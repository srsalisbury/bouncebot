package solver

import (
	"context"
	"testing"

	"github.com/srsalisbury/bouncebot/model"
)

type mockSolver struct {
	name string
}

func (s *mockSolver) Name() string {
	return s.name
}

func (s *mockSolver) Solve(ctx context.Context, game *model.Game) Result {
	return Result{SolverName: s.name, Completed: true}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	s := &mockSolver{name: "test"}

	r.Register(s)

	got := r.Get("test")
	if got == nil {
		t.Fatal("expected solver to be registered")
	}
	if got.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", got.Name())
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()

	got := r.Get("nonexistent")
	if got != nil {
		t.Error("expected nil for nonexistent solver")
	}
}

func TestRegistry_All(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockSolver{name: "a"})
	r.Register(&mockSolver{name: "b"})

	all := r.All()
	if len(all) != 2 {
		t.Errorf("expected 2 solvers, got %d", len(all))
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockSolver{name: "a"})
	r.Register(&mockSolver{name: "b"})

	names := r.Names()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestRegistry_Replace(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockSolver{name: "test"})
	r.Register(&mockSolver{name: "test"}) // Should replace

	all := r.All()
	if len(all) != 1 {
		t.Errorf("expected 1 solver after replace, got %d", len(all))
	}
}
