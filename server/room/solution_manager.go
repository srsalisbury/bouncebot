package room

import (
	"fmt"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// SolutionManager handles solution submission.
type SolutionManager interface {
	// SubmitSolution validates and records a player's solution.
	// Returns (solution, signals) or error.
	SubmitSolution(room *Room, playerID string, moves []model.BotPosition) (*PlayerSolution, []Signal, error)

	// GetWinningSolution returns the winning solution from a list.
	// Public because GameLifecycle needs it.
	GetWinningSolution(solutions []PlayerSolution) *PlayerSolution
}

// solutionManager is the concrete implementation of SolutionManager.
type solutionManager struct{}

// NewSolutionManager creates a new SolutionManager.
func NewSolutionManager() SolutionManager {
	return &solutionManager{}
}

func (sm *solutionManager) SubmitSolution(room *Room, playerID string, moves []model.BotPosition) (*PlayerSolution, []Signal, error) {
	if room.CurrentGame == nil {
		return nil, nil, fmt.Errorf("no game in progress")
	}

	// Verify player exists
	if room.GetPlayerName(playerID) == "" {
		return nil, nil, fmt.Errorf("player not found: %s", playerID)
	}

	// Verify the solution
	isValid, _ := room.CurrentGame.CheckSolution(moves)
	if !isValid {
		return nil, nil, fmt.Errorf("invalid solution")
	}

	moveCount := len(moves)
	now := time.Now()
	room.LastActivityAt = now

	// Check if player already submitted a solution for this game
	for i := range room.Solutions {
		if room.Solutions[i].PlayerID == playerID {
			// Update if better solution
			if moveCount < room.Solutions[i].MoveCount() {
				room.Solutions[i].SolvedAt = now
				room.Solutions[i].Moves = moves

				signals := []Signal{
					BroadcastSignal{Event: PlayerSolvedEvent{
						RoomID:    room.ID,
						PlayerID:  playerID,
						MoveCount: moveCount,
					}},
				}
				return &room.Solutions[i], signals, nil
			}
			// Return existing solution (no update needed, no broadcast)
			return &room.Solutions[i], nil, nil
		}
	}

	// New solution
	solution := PlayerSolution{
		PlayerID: playerID,
		SolvedAt: now,
		Moves:    moves,
	}
	room.Solutions = append(room.Solutions, solution)

	signals := []Signal{
		BroadcastSignal{Event: PlayerSolvedEvent{
			RoomID:    room.ID,
			PlayerID:  playerID,
			MoveCount: moveCount,
		}},
	}

	return &solution, signals, nil
}

func (sm *solutionManager) GetWinningSolution(solutions []PlayerSolution) *PlayerSolution {
	if len(solutions) == 0 {
		return nil
	}

	best := &solutions[0]
	for i := range solutions[1:] {
		sol := &solutions[i+1]
		if sol.MoveCount() < best.MoveCount() {
			best = sol
		} else if sol.MoveCount() == best.MoveCount() && sol.SolvedAt.Before(best.SolvedAt) {
			best = sol
		}
	}
	return best
}
