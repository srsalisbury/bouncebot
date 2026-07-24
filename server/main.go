package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/cors"
	"github.com/srsalisbury/bouncebot/model"
	"github.com/srsalisbury/bouncebot/proto/protoconnect"
	"github.com/srsalisbury/bouncebot/server/config"
	"github.com/srsalisbury/bouncebot/server/daily"
	"github.com/srsalisbury/bouncebot/server/ratelimit"
	"github.com/srsalisbury/bouncebot/server/room"
	"github.com/srsalisbury/bouncebot/server/ws"
	"github.com/srsalisbury/bouncebot/solver"
	"github.com/srsalisbury/bouncebot/solver/astar"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Version is set at build time via ldflags
var Version = "dev"

var (
	port = flag.Int("port", 0, "The server port (overrides PORT env var)")
)

func main() {
	flag.Parse()

	// Load configuration from environment variables
	cfg := config.LoadFromEnv()

	// Allow flags to override env vars
	if *port != 0 {
		cfg.Port = *port
	}

	log.Printf("BounceBot server version %s", Version)
	absDataDir, _ := filepath.Abs(cfg.DataDir)
	log.Printf("Configuration: port=%d, dataDir=%s, origins=%v", cfg.Port, absDataDir, cfg.AllowedOrigins)

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory %s: %v", cfg.DataDir, err)
	}

	rooms := room.NewRoomService()
	rooms.SetDisconnectGracePeriod(cfg.DisconnectGracePeriod)
	rooms.SetSoloDisconnectGracePeriod(cfg.SoloDisconnectGracePeriod)

	// Load existing rooms from disk (continue with empty list on failure)
	if err := rooms.Load(cfg.RoomsFile()); err != nil {
		log.Printf("Warning: Failed to load rooms from %s: %v (starting with empty room list)", cfg.RoomsFile(), err)
	}

	// Start auto-save goroutine
	stopAutoSave := rooms.StartAutoSave(cfg.RoomsFile(), cfg.AutoSaveInterval)

	// Clean up stale rooms immediately, then start periodic cleanup
	rooms.CleanupStaleRooms(cfg.RoomMaxAge)
	stopCleanup := rooms.StartCleanup(cfg.CleanupInterval, cfg.RoomMaxAge)

	// Create rate limiters
	getRoomLimiter := ratelimit.NewLimiter(100, time.Minute)      // 100 req/min per IP
	createRoomLimiter := ratelimit.NewLimiter(5, time.Minute)     // 5 req/min per IP
	submitDailyLimiter := ratelimit.NewLimiter(30, time.Minute)   // 30 req/min per IP
	stopRateLimitCleanup := getRoomLimiter.StartCleanup(5 * time.Minute)
	stopCreateRoomLimitCleanup := createRoomLimiter.StartCleanup(5 * time.Minute)
	stopSubmitDailyLimitCleanup := submitDailyLimiter.StartCleanup(5 * time.Minute)

	wsHub := ws.NewHub(rooms, cfg)
	rooms.SetBroadcaster(wsHub)

	// Create solver manager with completion callback and periodic cleanup
	solverMgr := solver.NewManager(solver.DefaultRegistry)
	stopSolverCleanup := solverMgr.StartCleanup(5 * time.Minute)

	// Handle graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-shutdownChan
		log.Println("Shutting down, saving rooms...")
		close(stopCleanup)
		close(stopAutoSave) // This triggers final save
		close(stopRateLimitCleanup)
		close(stopCreateRoomLimitCleanup)
		close(stopSubmitDailyLimitCleanup)
		stopSolverCleanup()
		os.Exit(0)
	}()

	// Check candidate boards against a room's configured minimum solution
	// length. Per-attempt timeout is intentionally much shorter than
	// cfg.SolverTimeout (used below for the post-hoc async solve) since this
	// runs synchronously, in a retry loop, while a player waits for game start.
	const boardSolveAttemptTimeout = 2 * time.Second
	rooms.SetBoardSolveFunc(func(game *model.Game) ([]model.BotPosition, bool) {
		ctx, cancel := context.WithTimeout(context.Background(), boardSolveAttemptTimeout)
		defer cancel()
		result := (&astar.AStarSolver{}).Solve(ctx, game)
		if !result.Completed || result.Solution == nil {
			return nil, false
		}
		return result.Solution.Moves, true
	})

	// Trigger all registered solvers when a game starts
	rooms.SetOnGameStart(func(r *room.Room) {
		if r.CurrentGame != nil {
			for _, solverName := range solver.DefaultRegistry.Names() {
				solverMgr.StartJob(r.ID, r.CurrentGame, cfg.SolverTimeout, solverName)
			}
		}
	})
	solverMgr.SetCompletionCallback(func(job *solver.Job) {
		if job.Result == nil {
			return
		}

		// Convert solution moves to MovePayload format
		var moves []room.MovePayload
		if job.Result.Solution != nil {
			for _, m := range job.Result.Solution.Moves {
				moves = append(moves, room.MovePayload{
					RobotId: int(m.Id),
					X:       int(m.Pos.X),
					Y:       int(m.Pos.Y),
				})
			}
		}

		errorMsg := ""
		if job.Result.Error != nil {
			errorMsg = job.Result.Error.Error()
		}

		// Store solver result in room for persistence across page reloads
		rooms.SetSolverResult(job.RoomID, &room.SolverResult{
			SolverName: job.Result.SolverName,
			Moves:      moves,
			Error:      errorMsg,
			Completed:  job.Result.Completed,
		})

		wsHub.BroadcastSolverComplete(job.RoomID, ws.SolverResultPayload{
			SolverName: job.Result.SolverName,
			Moves:      moves,
			Error:      errorMsg,
			Completed:  job.Result.Completed,
		})
	})

	// Initialize daily challenge system (behind feature flag)
	var dailyMgr *daily.Manager
	var dailyProgressMgr *daily.ProgressManager
	if cfg.EnableDailyChallenge {
		log.Println("Daily challenges enabled")
		dailyMgr = daily.NewManager(cfg.DataDir, solverMgr)
		dailyProgressMgr = daily.NewProgressManager(cfg.DataDir)

		// Start daily puzzle generation worker (2-day buffer)
		dailyCtx, dailyCancel := context.WithCancel(context.Background())
		dailyMgr.StartGenerationWorker(dailyCtx, 2)

		// Add daily cancel to shutdown handler
		go func() {
			<-shutdownChan
			dailyCancel()
		}()
	} else {
		log.Println("Daily challenges disabled (set ENABLE_DAILY_CHALLENGE=true to enable)")
	}

	mux := http.NewServeMux()
	path, handler := protoconnect.NewBounceBotHandler(NewBounceBotServer(rooms, getRoomLimiter, createRoomLimiter, submitDailyLimiter, dailyMgr, dailyProgressMgr))
	// Wrap handler with client IP middleware for rate limiting
	mux.Handle(path, ratelimit.InjectClientIPMiddleware(handler))

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsHub.HandleWebSocket)

	// Join-room link preview: a server-rendered page with real Open Graph
	// tags (crawlers don't run JS, so the SPA's static index.html can't
	// carry per-room content) that forwards real browsers into the app.
	mux.HandleFunc("GET /join/{roomId}", handleJoinPage(rooms, cfg.PublicClientURL, cfg.PublicServerURL))
	mux.HandleFunc("GET /join/{roomId}/preview.png", handleJoinPreviewImage(rooms))

	// CORS configuration for browser access
	corsHandler := cors.New(cors.Options{
		AllowOriginRequestFunc: func(r *http.Request, origin string) bool {
			return cfg.IsOriginAllowedForRequest(origin, r.Host)
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Connect-Protocol-Version",
			"Connect-Timeout-Ms",
			"Grpc-Timeout",
			"X-Grpc-Web",
			"X-User-Agent",
		},
		ExposedHeaders: []string{
			"Grpc-Status",
			"Grpc-Message",
			"Grpc-Status-Details-Bin",
		},
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("BounceBot Connect server listening at %s", addr)

	// Use h2c to support HTTP/2 without TLS (needed for gRPC clients)
	h2cHandler := h2c.NewHandler(corsHandler.Handler(mux), &http2.Server{})
	if err := http.ListenAndServe(addr, h2cHandler); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
