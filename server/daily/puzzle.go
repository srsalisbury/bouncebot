// Package daily provides daily challenge puzzle management.
package daily

// DailyPuzzle represents a single puzzle for a difficulty level.
type DailyPuzzle struct {
	OptimalMoves int      `json:"optimal_moves"`
	Solution     []string `json:"solution"`  // "B0:up", "B1:left", etc.
	Game         []string `json:"game"`      // ASCII art format (lines)
}

// DailyPuzzles represents the set of puzzles for a given day.
type DailyPuzzles struct {
	Date   string       `json:"date"` // "2026-01-22"
	Easy   *DailyPuzzle `json:"easy"`
	Medium *DailyPuzzle `json:"medium"`
	Hard   *DailyPuzzle `json:"hard"`
}

// Difficulty constants for puzzle classification.
const (
	DifficultyEasy   = "easy"
	DifficultyMedium = "medium"
	DifficultyHard   = "hard"
)

// DifficultyThresholds define the minimum move counts for each level.
const (
	EasyMinMoves   = 4  // Easy: 4-6 moves
	MediumMinMoves = 7  // Medium: 7-11 moves
	HardMinMoves   = 12 // Hard: 12+ moves
)

// ClassifyDifficulty returns the difficulty level based on optimal move count.
// Returns empty string for puzzles that don't fit any difficulty (< EasyMinMoves).
func ClassifyDifficulty(optimalMoves int) string {
	if optimalMoves >= HardMinMoves {
		return DifficultyHard
	}
	if optimalMoves >= MediumMinMoves {
		return DifficultyMedium
	}
	if optimalMoves >= EasyMinMoves {
		return DifficultyEasy
	}
	return ""
}
