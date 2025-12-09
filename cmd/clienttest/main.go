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
	addr = flag.String("addr", "localhost:8080", "the address to connect to")
)

// Why isn't this built in?
func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

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

	moves := []model.BotPosition{
		model.NewBotPosition(1, 0, 12),
		model.NewBotPosition(0, 5, 0),
		model.NewBotPosition(0, 2, 0),
		model.NewBotPosition(0, 2, 15),
		model.NewBotPosition(0, 0, 15),
		model.NewBotPosition(0, 0, 13),
		model.NewBotPosition(0, 5, 13),
	}
	log.Printf("Submitting moves: %v", moves)
	check_request := &pb.CheckSolutionRequest{
		Game:  gr,
		Moves: Map(moves, func(bp model.BotPosition) *pb.BotPos { return bp.ToProto() }),
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
