package astar

import (
	"context"
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

func TestAStarSolver_Name(t *testing.T) {
	s := &AStarSolver{}
	if s.Name() != "A-Star" {
		t.Errorf("expected name 'A-Star', got '%s'", s.Name())
	}
}

func TestAStarSolver_AlreadySolved(t *testing.T) {
	s := &AStarSolver{}
	game := model.Game1()

	// Move target bot to target position
	game.Bots[game.Target.Id] = game.Target.Pos

	ctx := context.Background()
	result := s.Solve(ctx, game)

	if !result.Completed {
		t.Error("expected completed")
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if result.Solution == nil {
		t.Fatal("expected solution")
	}
	if len(result.Solution.Moves) != 0 {
		t.Errorf("expected 0 moves for already solved, got %d", len(result.Solution.Moves))
	}
}

func TestAStarSolver_FindsSolution(t *testing.T) {
	s := &AStarSolver{}
	game := model.Game1()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := s.Solve(ctx, game)

	if !result.Completed {
		t.Errorf("expected completed, got error: %v", result.Error)
	}
	if result.Solution == nil {
		t.Fatal("expected solution")
	}
	if len(result.Solution.Moves) == 0 {
		t.Error("expected at least one move")
	}

	// Verify the solution is valid
	valid, _ := game.CheckSolution(result.Solution.Moves)
	if !valid {
		t.Error("solution is not valid")
	}

	t.Logf("Found solution with %d moves", len(result.Solution.Moves))
}

func TestAStarSolver_Timeout(t *testing.T) {
	s := &AStarSolver{}
	game := model.Game1()

	// Very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(1 * time.Millisecond)

	result := s.Solve(ctx, game)

	if result.Completed {
		// Might find solution before timeout check - that's OK
		t.Log("Solver completed before timeout (this is acceptable)")
	} else {
		if result.Error == nil {
			t.Error("expected error for timeout")
		}
	}
}

func TestAStarSolver_OptimalSolution(t *testing.T) {
	s := &AStarSolver{}
	game := model.Game1()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := s.Solve(ctx, game)

	if !result.Completed || result.Solution == nil {
		t.Skip("Could not find solution")
	}

	// A* should find the optimal (shortest) solution
	// Game1 has a known optimal solution length of 7
	expectedMoves := 7
	if len(result.Solution.Moves) != expectedMoves {
		t.Errorf("expected optimal solution of %d moves, got %d", expectedMoves, len(result.Solution.Moves))
	}
}

func TestAStarSolver_SolverName(t *testing.T) {
	s := &AStarSolver{}
	game := model.Game1()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := s.Solve(ctx, game)

	if result.SolverName != "A-Star" {
		t.Errorf("expected solver name 'A-Star', got '%s'", result.SolverName)
	}
}
