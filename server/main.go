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
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	port = flag.Int("port", 8080, "The server port")
)

type bounceBotServer struct{}

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

func main() {
	flag.Parse()

	mux := http.NewServeMux()
	path, handler := protoconnect.NewBounceBotHandler(&bounceBotServer{})
	mux.Handle(path, handler)

	// CORS configuration for browser access
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
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
