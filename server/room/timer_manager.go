package room

import (
	"sync"
	"time"
)

// TimerCallback is called when a disconnect timer fires.
type TimerCallback func(roomID, playerID string)

// TimerManager manages disconnect grace period timers.
type TimerManager interface {
	// StartTimer starts a timer for the given player.
	// Cancels any existing timer for this player.
	StartTimer(roomID, playerID string, duration time.Duration, callback TimerCallback)

	// CancelTimer cancels the timer for the given player.
	CancelTimer(playerID string)

	// StopAll cancels all timers.
	StopAll()

	// HasTimer returns true if a timer exists for the given player (for testing).
	HasTimer(playerID string) bool
}

// timerManager is the concrete implementation of TimerManager.
type timerManager struct {
	mu     sync.Mutex
	timers map[string]*time.Timer
}

// NewTimerManager creates a new TimerManager.
func NewTimerManager() TimerManager {
	return &timerManager{
		timers: make(map[string]*time.Timer),
	}
}

func (tm *timerManager) StartTimer(roomID, playerID string, duration time.Duration, callback TimerCallback) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Cancel existing timer for this player
	if old, ok := tm.timers[playerID]; ok {
		old.Stop()
		delete(tm.timers, playerID)
	}

	timer := time.AfterFunc(duration, func() {
		callback(roomID, playerID)
	})
	tm.timers[playerID] = timer
}

func (tm *timerManager) CancelTimer(playerID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if timer, ok := tm.timers[playerID]; ok {
		timer.Stop()
		delete(tm.timers, playerID)
	}
}

func (tm *timerManager) StopAll() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for id, timer := range tm.timers {
		timer.Stop()
		delete(tm.timers, id)
	}
}

func (tm *timerManager) HasTimer(playerID string) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	_, ok := tm.timers[playerID]
	return ok
}
