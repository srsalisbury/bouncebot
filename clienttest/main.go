package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50055", "the address to connect to")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBounceBotClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	gr, err := c.MakeGame(ctx, &pb.MakeGameRequest{Size: 16})
	if err != nil {
		log.Fatalf("could not make game: %v", err)
	}
	game := model.NewGameFromProto(gr)
	log.Printf("Got Game:\n%s", game)

	// TODO: Replace with BotPosition and convert to pb.
	moves := []*pb.BotPos{
		{Id: 1, Pos: &pb.Position{X: 0, Y: 12}},
		{Id: 0, Pos: &pb.Position{X: 5, Y: 0}},
		{Id: 0, Pos: &pb.Position{X: 2, Y: 0}},
		{Id: 0, Pos: &pb.Position{X: 2, Y: 15}},
		{Id: 0, Pos: &pb.Position{X: 0, Y: 15}},
		{Id: 0, Pos: &pb.Position{X: 0, Y: 13}},
		{Id: 0, Pos: &pb.Position{X: 5, Y: 13}},
	}
	log.Printf("Submitting moves: %#v", moves)
	check_request := &pb.CheckSolutionRequest{
		Game:  gr,
		Moves: moves,
	}
	cr, err := c.CheckSolution(ctx, check_request)
	if err != nil {
		log.Fatalf("could not check solution: %v", err)
	}
	log.Printf("Checked solution: isValid=%v, numMoves=%d", cr.IsValid, cr.NumMoves)
	if cr.ResultingGame != nil {
		resulting_game := model.NewGameFromProto(cr.ResultingGame)
		log.Printf("Resulting game:\n%s", resulting_game)
	}
}
