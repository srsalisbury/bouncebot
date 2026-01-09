package solver

import (
	"context"

	"github.com/srsalisbury/bouncebot/model"
)

// Solution represents a valid solution to a puzzle.
type Solution struct {
	Moves []model.BotPosition
}

// MoveCount returns the number of moves in the solution.
func (s *Solution) MoveCount() int {
	if s == nil {
		return 0
	}
	return len(s.Moves)
}

// Result represents the outcome of a solver run.
type Result struct {
	SolverName string
	Solution   *Solution
	Error      error
	Completed  bool // true if solver finished (vs timed out)
}

// Solver is the interface that all puzzle solvers must implement.
type Solver interface {
	// Name returns the solver's identifier.
	Name() string

	// Solve attempts to find a solution to the given game.
	// The context is used for timeout/cancellation.
	// Returns a Result with the solution (if found) or error.
	Solve(ctx context.Context, game *model.Game) Result
}
