package room

import (
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
