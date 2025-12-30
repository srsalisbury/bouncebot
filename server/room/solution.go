package room

import (
	"fmt"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// PlayerSolution represents a player's solution to the current game.
type PlayerSolution struct {
	PlayerID string
	SolvedAt time.Time
	Moves    []model.BotPosition // The actual moves that solved the puzzle
}

// MoveCount returns the number of moves in the solution.
func (s *PlayerSolution) MoveCount() int {
	return len(s.Moves)
}

// PlayerSolutionHistory tracks all solutions a player has found (for restoring after retraction).
type PlayerSolutionHistory struct {
	PlayerID  string
	Solutions []PlayerSolution
}

// SubmitSolution records a player's solution for the current game.
// If moves are provided, they are verified against the current game state.
func (store *Store) SubmitSolution(roomID, playerID string, moves []model.BotPosition) (*PlayerSolution, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	if room.CurrentGame == nil {
		return nil, fmt.Errorf("no game in progress")
	}

	// Verify player exists
	if room.GetPlayerName(playerID) == "" {
		return nil, fmt.Errorf("player not found: %s", playerID)
	}

	// Verify the solution
	isValid, _ := room.CurrentGame.CheckSolution(moves)
	if !isValid {
		return nil, fmt.Errorf("invalid solution")
	}

	moveCount := len(moves)
	now := time.Now()
	room.LastActivityAt = now

	// Add to solution history
	store.addToHistory(room, playerID, moves, now)

	// Check if player already submitted a solution for this game
	for i := range room.Solutions {
		if room.Solutions[i].PlayerID == playerID {
			// Update if better solution
			if moveCount < room.Solutions[i].MoveCount() {
				room.Solutions[i].SolvedAt = now
				room.Solutions[i].Moves = moves
				// Broadcast updated solution
				if store.broadcaster != nil {
					store.broadcaster.BroadcastPlayerSolved(roomID, playerID, moveCount)
				}
			}
			// Return existing (possibly updated) solution
			return &room.Solutions[i], nil
		}
	}

	solution := PlayerSolution{
		PlayerID: playerID,
		SolvedAt: now,
		Moves:    moves,
	}
	room.Solutions = append(room.Solutions, solution)

	// Broadcast player solved event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerSolved(roomID, playerID, moveCount)
	}

	return &solution, nil
}

// addToHistory adds a solution to the player's history (if not already present with same move count).
func (store *Store) addToHistory(room *Room, playerID string, moves []model.BotPosition, solvedAt time.Time) {
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

// RetractSolution removes a player's current best solution and restores the previous one from history.
func (store *Store) RetractSolution(roomID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	if room.CurrentGame == nil {
		return fmt.Errorf("no game in progress")
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
		return fmt.Errorf("no solution found for player: %s", playerID)
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

			// Broadcast the restored solution
			if store.broadcaster != nil {
				store.broadcaster.BroadcastPlayerSolved(roomID, playerID, history.Solutions[bestIdx].MoveCount())
			}
			return nil
		}
	}

	// No previous solution to restore - remove completely
	room.Solutions = append(room.Solutions[:solutionIndex], room.Solutions[solutionIndex+1:]...)

	// Broadcast solution retracted event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastSolutionRetracted(roomID, playerID)
	}

	return nil
}

// getWinningSolution finds the winning solution (lowest moves, earliest time as tiebreaker).
func (store *Store) getWinningSolution(solutions []PlayerSolution) *PlayerSolution {
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
