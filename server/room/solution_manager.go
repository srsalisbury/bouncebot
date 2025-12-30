package room

import (
	"fmt"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// SolutionManager handles solution submission and retraction.
type SolutionManager interface {
	// SubmitSolution validates and records a player's solution.
	// Returns (solution, signals) or error.
	SubmitSolution(room *Room, playerID string, moves []model.BotPosition) (*PlayerSolution, []Signal, error)

	// RetractSolution removes a player's current solution.
	// Returns signals or error.
	RetractSolution(room *Room, playerID string) ([]Signal, error)

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

	// Add to solution history
	sm.addToHistory(room, playerID, moves, now)

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

// addToHistory adds a solution to the player's history (if not already present with same move count).
func (sm *solutionManager) addToHistory(room *Room, playerID string, moves []model.BotPosition, solvedAt time.Time) {
	// Find or create history entry for this player
	var history *PlayerSolutionHistory
	for i := range room.SolutionHistory {
		if room.SolutionHistory[i].PlayerID == playerID {
			history = &room.SolutionHistory[i]
			break
		}
	}
	if history == nil {
		room.SolutionHistory = append(room.SolutionHistory, PlayerSolutionHistory{
			PlayerID: playerID,
		})
		history = &room.SolutionHistory[len(room.SolutionHistory)-1]
	}

	// Check if we already have this move count in history
	moveCount := len(moves)
	for _, sol := range history.Solutions {
		if sol.MoveCount() == moveCount {
			return // Already have this solution
		}
	}

	// Add to history
	history.Solutions = append(history.Solutions, PlayerSolution{
		PlayerID: playerID,
		SolvedAt: solvedAt,
		Moves:    moves,
	})
}

func (sm *solutionManager) RetractSolution(room *Room, playerID string) ([]Signal, error) {
	if room.CurrentGame == nil {
		return nil, fmt.Errorf("no game in progress")
	}

	room.LastActivityAt = time.Now()

	// Find the player's current solution
	var currentMoveCount int
	var solutionIndex int = -1
	for i, sol := range room.Solutions {
		if sol.PlayerID == playerID {
			currentMoveCount = sol.MoveCount()
			solutionIndex = i
			break
		}
	}

	if solutionIndex == -1 {
		return nil, fmt.Errorf("no solution found for player: %s", playerID)
	}

	// Find the player's history and remove the current solution from it
	var history *PlayerSolutionHistory
	for i := range room.SolutionHistory {
		if room.SolutionHistory[i].PlayerID == playerID {
			history = &room.SolutionHistory[i]
			break
		}
	}

	// Remove current move count from history
	if history != nil {
		for i, sol := range history.Solutions {
			if sol.MoveCount() == currentMoveCount {
				history.Solutions = append(history.Solutions[:i], history.Solutions[i+1:]...)
				break
			}
		}

		// Find the next best solution in history (smallest move count remaining)
		if len(history.Solutions) > 0 {
			bestIdx := 0
			for i, sol := range history.Solutions {
				if sol.MoveCount() < history.Solutions[bestIdx].MoveCount() {
					bestIdx = i
				}
			}
			// Restore the previous best solution
			room.Solutions[solutionIndex] = history.Solutions[bestIdx]

			signals := []Signal{
				BroadcastSignal{Event: PlayerSolvedEvent{
					RoomID:    room.ID,
					PlayerID:  playerID,
					MoveCount: history.Solutions[bestIdx].MoveCount(),
				}},
			}
			return signals, nil
		}
	}

	// No previous solution to restore - remove completely
	room.Solutions = append(room.Solutions[:solutionIndex], room.Solutions[solutionIndex+1:]...)

	signals := []Signal{
		BroadcastSignal{Event: SolutionRetractedEvent{
			RoomID:   room.ID,
			PlayerID: playerID,
		}},
	}

	return signals, nil
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
