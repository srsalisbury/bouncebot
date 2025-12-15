package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/rs/cors"
	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"github.com/srsalisbury/bouncebot/proto/protoconnect"
	"github.com/srsalisbury/bouncebot/server/session"
	"github.com/srsalisbury/bouncebot/server/ws"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	port = flag.Int("port", 8080, "The server port")
)

type bounceBotServer struct {
	sessions *session.Store
}

func (s *bounceBotServer) MakeGame(_ context.Context, req *connect.Request[pb.MakeGameRequest]) (*connect.Response[pb.Game], error) {
	game := model.Game1()
	return connect.NewResponse(game.ToProto()), nil
}

func (s *bounceBotServer) CheckSolution(_ context.Context, req *connect.Request[pb.CheckSolutionRequest]) (*connect.Response[pb.CheckSolutionResponse], error) {
	starting_game := model.NewGameFromProto(req.Msg.Game)
	moves := model.NewBotPositionsFromProto(req.Msg.Moves)
	is_valid, resulting_game := starting_game.CheckSolution(moves)
	return connect.NewResponse(&pb.CheckSolutionResponse{
		IsValid:       is_valid,
		NumMoves:      int32(len(moves)),
		ResultingGame: resulting_game.ToProto(),
	}), nil
}

func (s *bounceBotServer) CreateSession(_ context.Context, req *connect.Request[pb.CreateSessionRequest]) (*connect.Response[pb.Session], error) {
	sess := s.sessions.Create(req.Msg.PlayerName)
	return connect.NewResponse(sess.ToProto()), nil
}

func (s *bounceBotServer) JoinSession(_ context.Context, req *connect.Request[pb.JoinSessionRequest]) (*connect.Response[pb.Session], error) {
	sess, err := s.sessions.Join(req.Msg.SessionId, req.Msg.PlayerName)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(sess.ToProto()), nil
}

func (s *bounceBotServer) GetSession(_ context.Context, req *connect.Request[pb.GetSessionRequest]) (*connect.Response[pb.Session], error) {
	sess, err := s.sessions.Get(req.Msg.SessionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(sess.ToProto()), nil
}

func (s *bounceBotServer) StartGame(_ context.Context, req *connect.Request[pb.StartGameRequest]) (*connect.Response[pb.Session], error) {
	sess, err := s.sessions.StartGame(req.Msg.SessionId, req.Msg.UseFixedBoard)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(sess.ToProto()), nil
}

func (s *bounceBotServer) SubmitSolution(_ context.Context, req *connect.Request[pb.SubmitSolutionRequest]) (*connect.Response[pb.SubmitSolutionResponse], error) {
	moves := model.NewBotPositionsFromProto(req.Msg.Moves)
	solution, err := s.sessions.SubmitSolution(req.Msg.SessionId, req.Msg.PlayerId, moves)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Convert moves back to proto for response
	protoMoves := make([]*pb.BotPos, len(solution.Moves))
	for i, move := range solution.Moves {
		protoMoves[i] = move.ToProto()
	}

	return connect.NewResponse(&pb.SubmitSolutionResponse{
		Solution: &pb.PlayerSolution{
			PlayerId: solution.PlayerID,
			Moves:    protoMoves,
		},
	}), nil
}

func (s *bounceBotServer) RetractSolution(_ context.Context, req *connect.Request[pb.RetractSolutionRequest]) (*connect.Response[pb.RetractSolutionResponse], error) {
	err := s.sessions.RetractSolution(req.Msg.SessionId, req.Msg.PlayerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&pb.RetractSolutionResponse{
		Success: true,
	}), nil
}

func (s *bounceBotServer) MarkFinishedSolving(_ context.Context, req *connect.Request[pb.MarkFinishedSolvingRequest]) (*connect.Response[pb.MarkFinishedSolvingResponse], error) {
	err := s.sessions.MarkFinishedSolving(req.Msg.SessionId, req.Msg.PlayerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&pb.MarkFinishedSolvingResponse{
		Success: true,
	}), nil
}

func (s *bounceBotServer) MarkReadyForNext(_ context.Context, req *connect.Request[pb.MarkReadyForNextRequest]) (*connect.Response[pb.MarkReadyForNextResponse], error) {
	err := s.sessions.MarkReadyForNext(req.Msg.SessionId, req.Msg.PlayerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&pb.MarkReadyForNextResponse{
		Success: true,
	}), nil
}

func main() {
	flag.Parse()

	sessions := session.NewStore()
	wsHub := ws.NewHub()
	sessions.SetBroadcaster(wsHub)

	mux := http.NewServeMux()
	path, handler := protoconnect.NewBounceBotHandler(&bounceBotServer{sessions: sessions})
	mux.Handle(path, handler)

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsHub.HandleWebSocket)

	// CORS configuration for browser access
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://guido.local:5173"},
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

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("BounceBot Connect server listening at %s", addr)

	// Use h2c to support HTTP/2 without TLS (needed for gRPC clients)
	h2cHandler := h2c.NewHandler(corsHandler.Handler(mux), &http2.Server{})
	if err := http.ListenAndServe(addr, h2cHandler); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
