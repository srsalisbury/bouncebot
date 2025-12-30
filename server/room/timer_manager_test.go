package room

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTimerManager_StartTimer(t *testing.T) {
	tm := NewTimerManager()

	var fired atomic.Bool
	tm.StartTimer("room1", "player1", 50*time.Millisecond, func(roomID, playerID string) {
		if roomID != "room1" || playerID != "player1" {
			t.Errorf("callback got wrong args: %s, %s", roomID, playerID)
		}
		fired.Store(true)
	})

	if !tm.HasTimer("player1") {
		t.Error("expected timer to exist after StartTimer")
	}

	// Wait for timer to fire
	time.Sleep(100 * time.Millisecond)

	if !fired.Load() {
		t.Error("expected timer callback to be called")
	}
}

func TestTimerManager_CancelTimer(t *testing.T) {
	tm := NewTimerManager()

	var fired atomic.Bool
	tm.StartTimer("room1", "player1", 50*time.Millisecond, func(roomID, playerID string) {
		fired.Store(true)
	})

	tm.CancelTimer("player1")

	if tm.HasTimer("player1") {
		t.Error("expected timer to be cancelled")
	}

	// Wait to verify callback doesn't fire
	time.Sleep(100 * time.Millisecond)

	if fired.Load() {
		t.Error("expected cancelled timer not to fire")
	}
}

func TestTimerManager_CancelTimer_NonExistent(t *testing.T) {
	tm := NewTimerManager()

	// Should not panic
	tm.CancelTimer("nonexistent")
}

func TestTimerManager_StartTimer_ReplacesExisting(t *testing.T) {
	tm := NewTimerManager()

	var firstFired, secondFired atomic.Bool

	// Start first timer
	tm.StartTimer("room1", "player1", 200*time.Millisecond, func(roomID, playerID string) {
		firstFired.Store(true)
	})

	// Replace with second timer (shorter duration)
	tm.StartTimer("room1", "player1", 50*time.Millisecond, func(roomID, playerID string) {
		secondFired.Store(true)
	})

	// Wait for second timer to fire
	time.Sleep(100 * time.Millisecond)

	if firstFired.Load() {
		t.Error("first timer should have been cancelled")
	}
	if !secondFired.Load() {
		t.Error("second timer should have fired")
	}
}

func TestTimerManager_HasTimer(t *testing.T) {
	tm := NewTimerManager()

	if tm.HasTimer("player1") {
		t.Error("expected no timer initially")
	}

	tm.StartTimer("room1", "player1", time.Hour, func(roomID, playerID string) {})

	if !tm.HasTimer("player1") {
		t.Error("expected timer to exist")
	}

	tm.CancelTimer("player1")

	if tm.HasTimer("player1") {
		t.Error("expected timer to be gone after cancel")
	}
}

func TestTimerManager_StopAll(t *testing.T) {
	tm := NewTimerManager()

	var count atomic.Int32

	// Start multiple timers
	tm.StartTimer("room1", "player1", 50*time.Millisecond, func(roomID, playerID string) {
		count.Add(1)
	})
	tm.StartTimer("room1", "player2", 50*time.Millisecond, func(roomID, playerID string) {
		count.Add(1)
	})
	tm.StartTimer("room2", "player3", 50*time.Millisecond, func(roomID, playerID string) {
		count.Add(1)
	})

	tm.StopAll()

	if tm.HasTimer("player1") || tm.HasTimer("player2") || tm.HasTimer("player3") {
		t.Error("expected all timers to be stopped")
	}

	// Wait and verify no callbacks fired
	time.Sleep(100 * time.Millisecond)

	if count.Load() != 0 {
		t.Errorf("expected 0 callbacks, got %d", count.Load())
	}
}

func TestTimerManager_Concurrent(t *testing.T) {
	tm := NewTimerManager()

	var wg sync.WaitGroup
	var fired atomic.Int32

	// Concurrent operations on different players
	for i := 0; i < 10; i++ {
		wg.Add(1)
		playerID := string(rune('A' + i))
		go func(pid string) {
			defer wg.Done()
			tm.StartTimer("room1", pid, 50*time.Millisecond, func(roomID, playerID string) {
				fired.Add(1)
			})
		}(playerID)
	}

	wg.Wait()

	// All timers should exist
	for i := 0; i < 10; i++ {
		playerID := string(rune('A' + i))
		if !tm.HasTimer(playerID) {
			t.Errorf("expected timer for player %s", playerID)
		}
	}

	// Wait for all to fire
	time.Sleep(100 * time.Millisecond)

	if fired.Load() != 10 {
		t.Errorf("expected 10 callbacks, got %d", fired.Load())
	}
}

func TestTimerManager_ConcurrentStartCancel(t *testing.T) {
	tm := NewTimerManager()

	var wg sync.WaitGroup

	// Rapidly start and cancel same timer
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			tm.StartTimer("room1", "player1", time.Hour, func(roomID, playerID string) {})
		}()
		go func() {
			defer wg.Done()
			tm.CancelTimer("player1")
		}()
	}

	wg.Wait()
	// Should not deadlock or panic
}
