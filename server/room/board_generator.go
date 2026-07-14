package room

import (
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// BoardSolveFunc attempts to find an optimal (shortest) solution for a candidate
// board. Implementations must enforce their own per-attempt timeout and must be
// safe to call from any goroutine without any room lock held. Injected from
// main.go so this package has no compile-time dependency on the solver package.
type BoardSolveFunc func(game *model.Game) (moves []model.BotPosition, solved bool)

// boardGenAttemptCap bounds the number of candidate boards generateBoard will
// try, independent of the wall-clock cap below.
const boardGenAttemptCap = 300

// boardGenDeadline bounds the total wall-clock time generateBoard will spend
// searching, regardless of attempt count. A var (not const) so tests can
// override it to exercise the deadline-exhaustion path without waiting.
var boardGenDeadline = 5 * time.Second

// generateBoard repeatedly calls candidateFn to produce a new candidate board
// and, if solve is set and minLength > 1, uses solve to check whether the
// candidate's optimal solution is at least minLength moves, retrying until a
// qualifying board is found or the attempt/deadline budget is exhausted.
//
// Fallback policy: if no qualifying board is found within budget, the best
// (longest-solved) candidate seen is returned instead of failing. If nothing
// solved at all within budget, the most recently generated candidate is
// returned unsolved (moves == nil) rather than nil, so callers can always
// proceed with game start.
//
// minLength <= 1 (or solve == nil) is a fast path: every board produced by
// model.NewRandomGame/NewContinuationGame is solvable in at least 1 move by
// construction, so candidateFn is called exactly once with no solve at all.
func generateBoard(minLength int, candidateFn func() *model.Game, solve BoardSolveFunc) (*model.Game, []model.BotPosition) {
	if minLength <= MinMinSolutionLength || solve == nil {
		return candidateFn(), nil
	}

	deadline := time.Now().Add(boardGenDeadline)
	var best *model.Game
	var bestMoves []model.BotPosition
	var last *model.Game

	for attempt := 0; attempt < boardGenAttemptCap && time.Now().Before(deadline); attempt++ {
		candidate := candidateFn()
		last = candidate

		moves, solved := solve(candidate)
		if !solved {
			continue
		}
		if len(moves) >= minLength {
			return candidate, moves
		}
		if len(moves) > len(bestMoves) {
			best, bestMoves = candidate, moves
		}
	}

	if best != nil {
		return best, bestMoves
	}
	return last, nil
}
