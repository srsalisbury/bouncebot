package daily

import (
	"testing"
	"time"
)

func TestClassifyDifficulty(t *testing.T) {
	tests := []struct {
		moves    int
		expected string
	}{
		{3, ""},
		{4, DifficultyEasy},
		{6, DifficultyEasy},
		{7, DifficultyMedium},
		{11, DifficultyMedium},
		{12, DifficultyHard},
		{20, DifficultyHard},
	}
	for _, tc := range tests {
		got := ClassifyDifficulty(tc.moves)
		if got != tc.expected {
			t.Errorf("ClassifyDifficulty(%d) = %q, want %q", tc.moves, got, tc.expected)
		}
	}
}

func TestDateToSeed_Deterministic(t *testing.T) {
	seed1 := dateToSeed("2026-01-15")
	seed2 := dateToSeed("2026-01-15")
	if seed1 != seed2 {
		t.Errorf("same date returned different seeds: %d vs %d", seed1, seed2)
	}
	if seed1 == 0 {
		t.Error("seed should not be zero for a valid date")
	}
}

func TestDateToSeed_DifferentDates(t *testing.T) {
	seed1 := dateToSeed("2026-01-15")
	seed2 := dateToSeed("2026-01-16")
	if seed1 == seed2 {
		t.Errorf("different dates returned same seed: %d", seed1)
	}
}

func TestGetPuzzlesForDate_CacheHit(t *testing.T) {
	m := &Manager{
		cache: make(map[string]puzzleCacheEntry),
	}

	// Manually insert a cache entry
	puzzles := &DailyPuzzles{Date: "2026-01-15"}
	m.cache["2026-01-15"] = puzzleCacheEntry{
		puzzles:  puzzles,
		cachedAt: time.Now(),
	}

	got, err := m.GetPuzzlesForDate("2026-01-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != puzzles {
		t.Error("expected cached puzzles to be returned")
	}
}

func TestGetPuzzlesForDate_CacheEviction(t *testing.T) {
	m := &Manager{
		cache: make(map[string]puzzleCacheEntry),
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Insert entries: one recent, one old (>7 days)
	recentDate := today.Format("2006-01-02")
	oldDate := today.AddDate(0, 0, -10).Format("2006-01-02")

	m.cache[recentDate] = puzzleCacheEntry{
		puzzles:  &DailyPuzzles{Date: recentDate},
		cachedAt: time.Now(),
	}
	m.cache[oldDate] = puzzleCacheEntry{
		puzzles:  &DailyPuzzles{Date: oldDate},
		cachedAt: time.Now().Add(-10 * 24 * time.Hour),
	}

	// Trigger eviction
	m.mu.Lock()
	m.evictCacheLocked()
	m.mu.Unlock()

	if _, ok := m.cache[recentDate]; !ok {
		t.Error("recent entry should not be evicted")
	}
	if _, ok := m.cache[oldDate]; ok {
		t.Error("old entry should be evicted")
	}
}
