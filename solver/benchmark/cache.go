package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/srsalisbury/bouncebot/model"
)

// CachedPuzzle represents a puzzle stored in the cache.
type CachedPuzzle struct {
	Seed         int64  `json:"seed"`
	OptimalMoves int    `json:"optimal_moves"`
	GameString   string `json:"game"`
}

// CacheFile represents a cache file containing puzzles of a specific difficulty.
type CacheFile struct {
	Difficulty int            `json:"difficulty"`
	Puzzles    []CachedPuzzle `json:"puzzles"`
}

// GetPuzzlesDir returns the path to the puzzles cache directory.
func GetPuzzlesDir() string {
	// Get the directory where this source file is located
	// For runtime, we use a path relative to the working directory
	return filepath.Join("solver", "benchmark", "puzzles")
}

// GetCacheFilePath returns the path to a cache file for a specific difficulty.
func GetCacheFilePath(difficulty int) string {
	return filepath.Join(GetPuzzlesDir(), fmt.Sprintf("difficulty_%d.json", difficulty))
}

// LoadCachedPuzzles loads puzzles from cache for a specific difficulty.
// Returns nil if cache doesn't exist or is invalid.
func LoadCachedPuzzles(difficulty int) []Puzzle {
	path := GetCacheFilePath(difficulty)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cache CacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}

	if cache.Difficulty != difficulty {
		return nil
	}

	puzzles := make([]Puzzle, 0, len(cache.Puzzles))
	for _, cp := range cache.Puzzles {
		game, err := model.ParseGameString(cp.GameString)
		if err != nil {
			continue
		}
		puzzles = append(puzzles, Puzzle{
			Game:         game,
			OptimalMoves: cp.OptimalMoves,
			Seed:         cp.Seed,
		})
	}

	return puzzles
}

// SaveCachedPuzzles saves puzzles to cache for a specific difficulty.
func SaveCachedPuzzles(difficulty int, puzzles []Puzzle) error {
	cache := CacheFile{
		Difficulty: difficulty,
		Puzzles:    make([]CachedPuzzle, len(puzzles)),
	}

	for i, p := range puzzles {
		cache.Puzzles[i] = CachedPuzzle{
			Seed:         p.Seed,
			OptimalMoves: p.OptimalMoves,
			GameString:   p.Game.String(),
		}
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := GetPuzzlesDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(GetCacheFilePath(difficulty), data, 0644)
}

// LoadOrGeneratePuzzles loads puzzles from cache, or generates and caches them if not found.
func LoadOrGeneratePuzzles(difficulties []int, puzzlesPerDifficulty int, timeout, genTimeout func() bool) map[int][]Puzzle {
	result := make(map[int][]Puzzle)

	for _, d := range difficulties {
		// Try to load from cache
		cached := LoadCachedPuzzles(d)
		if len(cached) >= puzzlesPerDifficulty {
			result[d] = cached[:puzzlesPerDifficulty]
			continue
		}

		// Need to generate - will be done by caller
		result[d] = nil
	}

	return result
}
