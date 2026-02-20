package daily

import (
	"sync"
	"testing"
)

func TestMarkSolved_IsSolved_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	pm := NewProgressManager(dir)

	playerID := "testplayer1"
	date := "2026-01-15"

	// Not solved initially
	solved, err := pm.IsSolved(playerID, date, DifficultyEasy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if solved {
		t.Error("expected not solved initially")
	}

	// Mark as solved
	if err := pm.MarkSolved(playerID, date, DifficultyEasy); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be solved now
	solved, err = pm.IsSolved(playerID, date, DifficultyEasy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !solved {
		t.Error("expected solved after MarkSolved")
	}

	// Other difficulties should not be solved
	solved, err = pm.IsSolved(playerID, date, DifficultyMedium)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if solved {
		t.Error("medium should not be solved")
	}
}

func TestGetUserProgress_Empty(t *testing.T) {
	dir := t.TempDir()
	pm := NewProgressManager(dir)

	progress, err := pm.GetUserProgress("newplayer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(progress) != 0 {
		t.Errorf("expected empty progress, got %d entries", len(progress))
	}
}

func TestMarkSolved_ConcurrentSamePlayer(t *testing.T) {
	dir := t.TempDir()
	pm := NewProgressManager(dir)

	playerID := "concurrentplayer"
	date := "2026-01-15"

	var wg sync.WaitGroup
	difficulties := []string{DifficultyEasy, DifficultyMedium, DifficultyHard}

	for _, diff := range difficulties {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			if err := pm.MarkSolved(playerID, date, d); err != nil {
				t.Errorf("MarkSolved(%s) error: %v", d, err)
			}
		}(diff)
	}
	wg.Wait()

	// All three difficulties should be solved
	for _, diff := range difficulties {
		solved, err := pm.IsSolved(playerID, date, diff)
		if err != nil {
			t.Fatalf("IsSolved(%s) error: %v", diff, err)
		}
		if !solved {
			t.Errorf("expected %s to be solved after concurrent MarkSolved", diff)
		}
	}
}

func TestValidatePlayerID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"abc123", true},
		{"player-1_test", true},
		{"A", true},
		{"", false},
		{"has spaces", false},
		{"has/slash", false},
		{"has..dots", false},
		// 65 chars (over 64 limit)
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
		// exactly 64 chars
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true},
	}
	for _, tc := range tests {
		err := validatePlayerID(tc.id)
		if tc.valid && err != nil {
			t.Errorf("validatePlayerID(%q) returned error: %v, expected valid", tc.id, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("validatePlayerID(%q) returned nil, expected error", tc.id)
		}
	}
}

func TestProgressCache_Eviction(t *testing.T) {
	dir := t.TempDir()
	pm := NewProgressManager(dir)

	// Fill cache beyond maxProgressCacheSize
	pm.mu.Lock()
	for i := range maxProgressCacheSize + 100 {
		id := "player" + string(rune('A'+i%26)) + string(rune('0'+i/26%10))
		// Use a unique string for each entry
		pm.cache[id] = progressCacheEntry{
			progress: make(UserProgress),
		}
	}
	pm.evictCacheLocked()
	pm.mu.Unlock()

	// Cache should have been cleared since all entries have zero lastAccessed
	// (before the 1-hour cutoff), and there were more than maxProgressCacheSize
	pm.mu.RLock()
	size := len(pm.cache)
	pm.mu.RUnlock()

	if size > maxProgressCacheSize {
		t.Errorf("cache size %d exceeds max %d after eviction", size, maxProgressCacheSize)
	}
}
