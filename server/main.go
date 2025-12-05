package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50055, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedBounceBotServer
}

func (s *server) MakeGame(_ context.Context, req *pb.MakeGameRequest) (*pb.Game, error) {
	game := model.Game1()
	return game.ToProto(), nil
}

func (s *server) CheckSolution(_ context.Context, req *pb.CheckSolutionRequest) (*pb.CheckSolutionResponse, error) {
	starting_game := model.NewGameFromProto(req.Game)
	moves := model.NewBotPositionsFromProto(req.Moves)
	is_valid, resulting_game := starting_game.CheckSolution(moves)
	return &pb.CheckSolutionResponse{
		IsValid:       is_valid,
		NumMoves:      int32(len(moves)),
		ResultingGame: resulting_game.ToProto(),
		// FirstBadMove: nil,
	}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBounceBotServer(s, &server{})
	log.Printf("BounceBot server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
