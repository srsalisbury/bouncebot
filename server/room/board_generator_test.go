package room

import (
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// newTestGame returns a distinct, otherwise-empty *model.Game. generateBoard
// never inspects a candidate's contents, so tests identify "which candidate
// was returned" by pointer identity rather than by any field value.
func newTestGame() *model.Game {
	return &model.Game{}
}

func TestGenerateBoard_DisabledFastPath(t *testing.T) {
	var candidates []*model.Game
	candidateFn := func() *model.Game {
		g := newTestGame()
		candidates = append(candidates, g)
		return g
	}
	solveCalls := 0
	solve := func(*model.Game) ([]model.BotPosition, bool) {
		solveCalls++
		return validSolution(), true
	}

	game, moves := generateBoard(MinMinSolutionLength, candidateFn, solve)

	if len(candidates) != 1 {
		t.Errorf("expected exactly 1 candidateFn call at minLength=1, got %d", len(candidates))
	}
	if solveCalls != 0 {
		t.Errorf("expected solve to never be called at minLength=1, got %d calls", solveCalls)
	}
	if moves != nil {
		t.Errorf("expected nil moves on the fast path, got %v", moves)
	}
	if game != candidates[0] {
		t.Error("expected the single generated candidate to be returned")
	}
}

func TestGenerateBoard_NilSolveFastPath(t *testing.T) {
	var candidates []*model.Game
	candidateFn := func() *model.Game {
		g := newTestGame()
		candidates = append(candidates, g)
		return g
	}

	game, moves := generateBoard(10, candidateFn, nil)

	if len(candidates) != 1 {
		t.Errorf("expected exactly 1 candidateFn call when solve is nil, got %d", len(candidates))
	}
	if moves != nil {
		t.Errorf("expected nil moves when solve is nil, got %v", moves)
	}
	if game != candidates[0] {
		t.Error("expected the single generated candidate to be returned even with solve==nil")
	}
}

func TestGenerateBoard_ImmediateQualifyingFind(t *testing.T) {
	first := newTestGame()
	candidateFn := func() *model.Game { return first }
	solve := func(*model.Game) ([]model.BotPosition, bool) {
		return make([]model.BotPosition, 8), true
	}

	game, moves := generateBoard(8, candidateFn, solve)

	if len(moves) != 8 {
		t.Errorf("expected 8 moves, got %d", len(moves))
	}
	if game != first {
		t.Error("expected the first candidate to qualify immediately")
	}
}

func TestGenerateBoard_FoundAfterNAttempts(t *testing.T) {
	var candidates []*model.Game
	candidateFn := func() *model.Game {
		g := newTestGame()
		candidates = append(candidates, g)
		return g
	}
	// Only the 5th attempt qualifies for minLength=6.
	solve := func(*model.Game) ([]model.BotPosition, bool) {
		attempt := len(candidates)
		if attempt < 5 {
			return make([]model.BotPosition, attempt), true
		}
		return make([]model.BotPosition, 6), true
	}

	game, moves := generateBoard(6, candidateFn, solve)

	if len(candidates) != 5 {
		t.Errorf("expected exactly 5 attempts, got %d", len(candidates))
	}
	if len(moves) != 6 {
		t.Errorf("expected 6 moves, got %d", len(moves))
	}
	if game != candidates[4] {
		t.Error("expected the 5th candidate to be returned")
	}
}

func TestGenerateBoard_AttemptCapFallsBackToBestFound(t *testing.T) {
	var candidates []*model.Game
	candidateFn := func() *model.Game {
		g := newTestGame()
		candidates = append(candidates, g)
		return g
	}
	// Never qualifies for minLength=10; move counts cycle so the best (5) shows
	// up partway through and nothing later beats it.
	lengths := []int{2, 5, 3, 4, 1}
	solve := func(*model.Game) ([]model.BotPosition, bool) {
		attempt := len(candidates)
		return make([]model.BotPosition, lengths[(attempt-1)%len(lengths)]), true
	}

	game, moves := generateBoard(10, candidateFn, solve)

	if len(candidates) != boardGenAttemptCap {
		t.Errorf("expected the attempt cap (%d) to be exhausted, got %d attempts", boardGenAttemptCap, len(candidates))
	}
	if len(moves) != 5 {
		t.Errorf("expected fallback to the best-found (5 moves), got %d", len(moves))
	}
	if game != candidates[1] { // the 2nd candidate is the first to hit length 5
		t.Error("expected the best-found (longest-solved) candidate to be returned")
	}
}

func TestGenerateBoard_DeadlineFallsBackToBestFound(t *testing.T) {
	origDeadline := boardGenDeadline
	boardGenDeadline = time.Millisecond
	defer func() { boardGenDeadline = origDeadline }()

	var candidates []*model.Game
	candidateFn := func() *model.Game {
		g := newTestGame()
		candidates = append(candidates, g)
		time.Sleep(500 * time.Microsecond) // slow enough to exhaust the 1ms deadline in a few attempts
		return g
	}
	solve := func(*model.Game) ([]model.BotPosition, bool) {
		return make([]model.BotPosition, 3), true // never qualifies for minLength=10
	}

	game, moves := generateBoard(10, candidateFn, solve)

	if len(candidates) >= boardGenAttemptCap {
		t.Errorf("expected the deadline to be hit well before the attempt cap, got %d attempts", len(candidates))
	}
	if len(moves) != 3 {
		t.Errorf("expected fallback to the only move-length seen (3), got %d", len(moves))
	}
	if game != candidates[0] {
		t.Error("expected the (only) seen candidate to be returned as the fallback")
	}
}

func TestGenerateBoard_NothingSolvesFallsBackToLastCandidate(t *testing.T) {
	var candidates []*model.Game
	candidateFn := func() *model.Game {
		g := newTestGame()
		candidates = append(candidates, g)
		return g
	}
	solve := func(*model.Game) ([]model.BotPosition, bool) {
		return nil, false // never solves
	}

	game, moves := generateBoard(10, candidateFn, solve)

	if len(candidates) != boardGenAttemptCap {
		t.Errorf("expected the attempt cap (%d) to be exhausted, got %d attempts", boardGenAttemptCap, len(candidates))
	}
	if moves != nil {
		t.Errorf("expected nil moves when nothing ever solved, got %v", moves)
	}
	if game != candidates[len(candidates)-1] {
		t.Error("expected the last generated candidate to be returned")
	}
}
