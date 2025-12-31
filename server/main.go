package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/cors"
	"github.com/srsalisbury/bouncebot/proto/protoconnect"
	"github.com/srsalisbury/bouncebot/server/config"
	"github.com/srsalisbury/bouncebot/server/room"
	"github.com/srsalisbury/bouncebot/server/ws"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	port     = flag.Int("port", 0, "The server port (overrides PORT env var)")
	dataFile = flag.String("data", "", "Path to room data file (overrides DATA_FILE env var)")
)

func main() {
	flag.Parse()

	// Load configuration from environment variables
	cfg := config.LoadFromEnv()

	// Allow flags to override env vars
	if *port != 0 {
		cfg.Port = *port
	}
	if *dataFile != "" {
		cfg.DataFile = *dataFile
	}

	log.Printf("Configuration: port=%d, data=%s, origins=%v", cfg.Port, cfg.DataFile, cfg.AllowedOrigins)

	rooms := room.NewRoomService()
	rooms.SetDisconnectGracePeriod(cfg.DisconnectGracePeriod)

	// Load existing rooms from disk (continue with empty list on failure)
	if err := rooms.Load(cfg.DataFile); err != nil {
		log.Printf("Warning: Failed to load rooms from %s: %v (starting with empty room list)", cfg.DataFile, err)
	}

	// Start auto-save goroutine
	stopAutoSave := rooms.StartAutoSave(cfg.DataFile, cfg.AutoSaveInterval)

	// Clean up stale rooms immediately, then start periodic cleanup
	rooms.CleanupStaleRooms(cfg.RoomMaxAge)
	stopCleanup := rooms.StartCleanup(cfg.CleanupInterval, cfg.RoomMaxAge)

	// Handle graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-shutdownChan
		log.Println("Shutting down, saving rooms...")
		close(stopCleanup)
		close(stopAutoSave) // This triggers final save
		os.Exit(0)
	}()

	wsHub := ws.NewHub(rooms, cfg)
	rooms.SetBroadcaster(wsHub)

	mux := http.NewServeMux()
	path, handler := protoconnect.NewBounceBotHandler(NewBounceBotServer(rooms))
	mux.Handle(path, handler)

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsHub.HandleWebSocket)

	// CORS configuration for browser access
	corsHandler := cors.New(cors.Options{
		AllowOriginFunc: cfg.IsOriginAllowed,
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
