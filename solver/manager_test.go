package solver

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

type slowSolver struct {
	delay time.Duration
}

func (s *slowSolver) Name() string {
	return "slow"
}

func (s *slowSolver) Solve(ctx context.Context, game *model.Game) Result {
	select {
	case <-time.After(s.delay):
		return Result{
			SolverName: s.Name(),
			Solution:   &Solution{Moves: []model.BotPosition{}},
			Completed:  true,
		}
	case <-ctx.Done():
		return Result{
			SolverName: s.Name(),
			Error:      ctx.Err(),
			Completed:  false,
		}
	}
}

func TestManager_StartJob(t *testing.T) {
	r := NewRegistry()
	r.Register(&slowSolver{delay: 10 * time.Millisecond})
	m := NewManager(r)

	var wg sync.WaitGroup
	wg.Add(1)
	m.SetCompletionCallback(func(job *Job) {
		wg.Done()
	})

	game := model.Game1()
	jobID := m.StartJob("room1", game, 5*time.Second, "slow")

	if jobID == "" {
		t.Fatal("expected job ID")
	}

	job := m.GetJob(jobID)
	if job == nil {
		t.Fatal("expected job")
	}
	if job.RoomID != "room1" {
		t.Errorf("expected room 'room1', got '%s'", job.RoomID)
	}

	// Wait for completion
	wg.Wait()

	job = m.GetJob(jobID)
	if !job.Done {
		t.Error("expected job to be done")
	}
	if job.Result == nil {
		t.Error("expected result")
	}
}

func TestManager_GetJobByRoom(t *testing.T) {
	r := NewRegistry()
	r.Register(&slowSolver{delay: 10 * time.Millisecond})
	m := NewManager(r)

	var wg sync.WaitGroup
	wg.Add(1)
	m.SetCompletionCallback(func(job *Job) {
		wg.Done()
	})

	game := model.Game1()
	m.StartJob("room1", game, 5*time.Second, "slow")

	job := m.GetJobByRoom("room1")
	if job == nil {
		t.Fatal("expected job for room")
	}
	if job.RoomID != "room1" {
		t.Errorf("expected room 'room1', got '%s'", job.RoomID)
	}

	wg.Wait()
}

func TestManager_StartJob_UnknownSolver(t *testing.T) {
	r := NewRegistry()
	m := NewManager(r)

	game := model.Game1()
	jobID := m.StartJob("room1", game, 5*time.Second, "nonexistent")

	if jobID != "" {
		t.Error("expected empty job ID for unknown solver")
	}
}

func TestManager_CleanupRoom(t *testing.T) {
	r := NewRegistry()
	r.Register(&slowSolver{delay: 10 * time.Millisecond})
	m := NewManager(r)

	var wg sync.WaitGroup
	wg.Add(1)
	m.SetCompletionCallback(func(job *Job) {
		wg.Done()
	})

	game := model.Game1()
	jobID := m.StartJob("room1", game, 5*time.Second, "slow")

	wg.Wait()

	m.CleanupRoom("room1")

	if m.GetJob(jobID) != nil {
		t.Error("expected job to be cleaned up")
	}
	if m.GetJobByRoom("room1") != nil {
		t.Error("expected room job to be cleaned up")
	}
}

func TestManager_JobTimeout(t *testing.T) {
	r := NewRegistry()
	r.Register(&slowSolver{delay: 1 * time.Second})
	m := NewManager(r)

	var wg sync.WaitGroup
	wg.Add(1)
	m.SetCompletionCallback(func(job *Job) {
		wg.Done()
	})

	game := model.Game1()
	jobID := m.StartJob("room1", game, 50*time.Millisecond, "slow")

	wg.Wait()

	job := m.GetJob(jobID)
	if job.Result.Completed {
		t.Error("expected job to timeout, not complete")
	}
	if job.Result.Error == nil {
		t.Error("expected timeout error")
	}
}
