// Package room provides multiplayer game room management.
package room

import (
	"time"

	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MinMinSolutionLength and MaxMinSolutionLength bound RoomSettings.MinSolutionLength.
const (
	MinMinSolutionLength = 1
	MaxMinSolutionLength = 10
)

// RoomSettings contains configurable settings for a room.
type RoomSettings struct {
	ShowSolverMoveCount bool // Show solver move count in header during game
	ShowSolverSolutions bool // Show solver solutions in end game screen
	// MinSolutionLength is the minimum accepted optimal-solution length, in moves
	// (inclusive). Range [MinMinSolutionLength, MaxMinSolutionLength]; values
	// outside that range (including the proto zero-value) are treated as
	// MinMinSolutionLength.
	MinSolutionLength int
}

// SolverResult represents the result from the automated solver.
type SolverResult struct {
	SolverName string
	Moves      []MovePayload
	Error      string
	Completed  bool
}

// Room represents a multiplayer game room.
type Room struct {
	ID              string
	Players         []Player
	PendingPlayers  []Player                  // Players waiting for next game to start
	CreatedAt       time.Time
	LastActivityAt  time.Time                 // Last user action timestamp (for cleanup)
	CurrentGame     *model.Game
	GameStartedAt   *time.Time
	Solutions       []PlayerSolution          // Current best solution per player
	Wins            map[string]int            // Wins per player ID
	GamesPlayed     int                       // Total games completed in room
	FinishedSolving []string                  // Player IDs who are finished solving (triggers game end)
	ReadyForNext    []string                  // Player IDs who are ready for next game
	IsSinglePlayer  bool                      // If true, only the creator can be in this room
	SolverResults   map[string]*SolverResult  // Solver solutions keyed by solver name
	Settings        RoomSettings              // Room settings configurable by host
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

// FindPendingPlayerIndex returns the index of the pending player with the given ID, or -1 if not found.
func (r *Room) FindPendingPlayerIndex(playerID string) int {
	for i, p := range r.PendingPlayers {
		if p.ID == playerID {
			return i
		}
	}
	return -1
}

// FindPlayerBySessionToken returns the player with the given session token, or nil if not found.
// Searches both Players and PendingPlayers.
func (r *Room) FindPlayerBySessionToken(token string) *Player {
	for i := range r.Players {
		if r.Players[i].SessionToken == token {
			return &r.Players[i]
		}
	}
	for i := range r.PendingPlayers {
		if r.PendingPlayers[i].SessionToken == token {
			return &r.PendingPlayers[i]
		}
	}
	return nil
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
	r.FinishedSolving = nil
	r.ReadyForNext = nil
	r.SolverResults = nil
}

// ToProto converts a Room to its protobuf representation.
func (r *Room) ToProto() *pb.Room {
	players := make([]*pb.Player, len(r.Players))
	for i, p := range r.Players {
		players[i] = &pb.Player{
			Id:         p.ID,
			Name:       p.Name,
			ColorIndex: p.ColorIndex,
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

	// Convert pending players
	pendingPlayers := make([]*pb.Player, len(r.PendingPlayers))
	for i, p := range r.PendingPlayers {
		pendingPlayers[i] = &pb.Player{
			Id:         p.ID,
			Name:       p.Name,
			ColorIndex: p.ColorIndex,
		}
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
		IsSinglePlayer:  r.IsSinglePlayer,
		PendingPlayers:  pendingPlayers,
		Settings: &pb.RoomSettings{
			ShowSolverMoveCount: r.Settings.ShowSolverMoveCount,
			ShowSolverSolutions: r.Settings.ShowSolverSolutions,
			MinSolutionLength:   int32(r.Settings.MinSolutionLength),
		},
	}

	if r.CurrentGame != nil {
		room.CurrentGame = r.CurrentGame.ToProto()
	}

	if r.GameStartedAt != nil {
		room.GameStartedAt = timestamppb.New(*r.GameStartedAt)
	}

	if len(r.SolverResults) > 0 {
		room.SolverResults = make([]*pb.SolverResult, 0, len(r.SolverResults))
		for _, sr := range r.SolverResults {
			moves := make([]*pb.BotPos, len(sr.Moves))
			for i, move := range sr.Moves {
				moves[i] = &pb.BotPos{
					Id:  int32(move.RobotId),
					Pos: &pb.Position{X: int32(move.X), Y: int32(move.Y)},
				}
			}
			room.SolverResults = append(room.SolverResults, &pb.SolverResult{
				SolverName: sr.SolverName,
				Moves:      moves,
				Error:      sr.Error,
				Completed:  sr.Completed,
			})
		}
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
	BroadcastGameEnded(roomID, winnerID, winnerName string, moves []MovePayload)
	BroadcastRoomSettingsChanged(roomID string)
}
