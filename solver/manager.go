package solver

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// Job represents an in-progress or completed solver job.
type Job struct {
	ID         string
	RoomID     string
	SolverName string
	StartedAt  time.Time
	Result     *Result
	Done       bool
}

// CompletionCallback is called when a solver job completes.
type CompletionCallback func(job *Job)

// Manager manages async solver execution.
type Manager struct {
	mu       sync.RWMutex
	jobs     map[string]*Job   // job ID -> job
	roomJobs map[string]string // room ID -> job ID (latest job for room)
	registry *Registry
	callback CompletionCallback
}

// NewManager creates a new solver manager.
func NewManager(registry *Registry) *Manager {
	return &Manager{
		jobs:     make(map[string]*Job),
		roomJobs: make(map[string]string),
		registry: registry,
	}
}

// SetCompletionCallback sets the callback for when jobs complete.
func (m *Manager) SetCompletionCallback(cb CompletionCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callback = cb
}

// StartJob starts a solver job for the given room.
// Returns the job ID.
func (m *Manager) StartJob(roomID string, game *model.Game, timeout time.Duration, solverName string) string {
	solver := m.registry.Get(solverName)
	if solver == nil {
		return ""
	}

	jobID := generateJobID()
	job := &Job{
		ID:         jobID,
		RoomID:     roomID,
		SolverName: solverName,
		StartedAt:  time.Now(),
		Done:       false,
	}

	m.mu.Lock()
	m.jobs[jobID] = job
	m.roomJobs[roomID] = jobID
	m.mu.Unlock()

	log.Printf("Solver: starting job %s for room %s using %s", jobID, roomID, solverName)

	// Run solver in goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		result := solver.Solve(ctx, game)

		// Reorder solution before storing to avoid race conditions
		if result.Completed && result.Solution != nil {
			result.Solution.Moves = ReorderSolution(game, result.Solution.Moves)
			log.Printf("Solver: job %s completed with %d moves", jobID, len(result.Solution.Moves))
		} else if result.Error != nil {
			log.Printf("Solver: job %s failed: %v", jobID, result.Error)
		}

		m.mu.Lock()
		job.Result = &result
		job.Done = true
		callback := m.callback
		m.mu.Unlock()

		if callback != nil {
			callback(job)
		}
	}()

	return jobID
}

// GetJob returns a job by ID.
func (m *Manager) GetJob(jobID string) *Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jobs[jobID]
}

// GetJobByRoom returns the latest job for a room.
func (m *Manager) GetJobByRoom(roomID string) *Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	jobID, ok := m.roomJobs[roomID]
	if !ok {
		return nil
	}
	return m.jobs[jobID]
}

// CleanupRoom removes jobs for a room.
func (m *Manager) CleanupRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if jobID, ok := m.roomJobs[roomID]; ok {
		delete(m.jobs, jobID)
		delete(m.roomJobs, roomID)
	}
}

// StartCleanup starts a background goroutine that periodically removes completed
// jobs older than 1 hour. Returns a function to stop the cleanup goroutine.
func (m *Manager) StartCleanup(interval time.Duration) func() {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.cleanup()
			case <-stop:
				return
			}
		}
	}()
	return func() { close(stop) }
}

func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	for id, job := range m.jobs {
		if job.Done && job.StartedAt.Before(cutoff) {
			delete(m.jobs, id)
		}
	}
	// Clean stale roomJobs entries pointing to deleted jobs
	for roomID, jobID := range m.roomJobs {
		if _, ok := m.jobs[jobID]; !ok {
			delete(m.roomJobs, roomID)
		}
	}
}

// generateJobID creates a unique job ID.
func generateJobID() string {
	return time.Now().Format("20060102150405.000000000")
}
