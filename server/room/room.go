// Package room provides multiplayer game room management.
package room

import (
	"time"

	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Room represents a multiplayer game room.
type Room struct {
	ID              string
	Players         []Player
	CreatedAt       time.Time
	LastActivityAt  time.Time               // Last user action timestamp (for cleanup)
	CurrentGame     *model.Game
	GameStartedAt   *time.Time
	Solutions       []PlayerSolution        // Current best solution per player
	SolutionHistory []PlayerSolutionHistory // All solutions per player (for retraction)
	Wins            map[string]int          // Wins per player ID
	GamesPlayed     int                     // Total games completed in room
	FinishedSolving []string                // Player IDs who are finished solving (triggers game end)
	ReadyForNext    []string                // Player IDs who are ready for next game
}

// GetPlayerName returns the name of the player with the given ID, or empty string if not found.
func (r *Room) GetPlayerName(playerID string) string {
	for _, p := range r.Players {
		if p.ID == playerID {
			return p.Name
		}
	}
	return ""
}

// FindPlayerIndex returns the index of the player with the given ID, or -1 if not found.
func (r *Room) FindPlayerIndex(playerID string) int {
	for i, p := range r.Players {
		if p.ID == playerID {
			return i
		}
	}
	return -1
}

// containsString returns true if the string is in the slice.
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// removeStringAt removes the element at index i from the slice.
func removeStringAt(slice []string, i int) []string {
	return append(slice[:i], slice[i+1:]...)
}

// ClearGameState resets the game-related state for a new game.
func (r *Room) ClearGameState() {
	r.Solutions = nil
	r.SolutionHistory = nil
	r.FinishedSolving = nil
	r.ReadyForNext = nil
}

// ToProto converts a Room to its protobuf representation.
func (r *Room) ToProto() *pb.Room {
	players := make([]*pb.Player, len(r.Players))
	for i, p := range r.Players {
		players[i] = &pb.Player{
			Id:   p.ID,
			Name: p.Name,
		}
	}

	solutions := make([]*pb.PlayerSolution, len(r.Solutions))
	for i, sol := range r.Solutions {
		moves := make([]*pb.BotPos, len(sol.Moves))
		for j, move := range sol.Moves {
			moves[j] = move.ToProto()
		}
		solutions[i] = &pb.PlayerSolution{
			PlayerId: sol.PlayerID,
			SolvedAt: timestamppb.New(sol.SolvedAt),
			Moves:    moves,
		}
	}

	// Convert wins map to proto
	scores := make([]*pb.PlayerScore, 0, len(r.Wins))
	for playerID, wins := range r.Wins {
		scores = append(scores, &pb.PlayerScore{
			PlayerId: playerID,
			Wins:     int32(wins),
		})
	}

	room := &pb.Room{
		Id:              r.ID,
		Players:         players,
		CreatedAt:       timestamppb.New(r.CreatedAt),
		Solutions:       solutions,
		Scores:          scores,
		GamesPlayed:     int32(r.GamesPlayed),
		FinishedSolving: r.FinishedSolving,
		ReadyForNext:    r.ReadyForNext,
	}

	if r.CurrentGame != nil {
		room.CurrentGame = r.CurrentGame.ToProto()
	}

	if r.GameStartedAt != nil {
		room.GameStartedAt = timestamppb.New(*r.GameStartedAt)
	}

	return room
}

// MovePayload represents a single move for WebSocket broadcast.
type MovePayload struct {
	RobotId int `json:"robotId"`
	X       int `json:"x"`
	Y       int `json:"y"`
}

// EventBroadcaster is an interface for broadcasting room events.
type EventBroadcaster interface {
	BroadcastPlayerJoined(roomID, playerID, playerName string)
	BroadcastPlayerLeft(roomID, playerID string)
	BroadcastGameStarted(roomID string)
	BroadcastPlayerFinishedSolving(roomID, playerID string)
	BroadcastPlayerReadyForNext(roomID, playerID string)
	BroadcastPlayerSolved(roomID, playerID string, moveCount int)
	BroadcastSolutionRetracted(roomID, playerID string)
	BroadcastGameEnded(roomID, winnerID, winnerName string, moves []MovePayload)
}
