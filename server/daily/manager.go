package daily

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/srsalisbury/bouncebot/model"
	"github.com/srsalisbury/bouncebot/solver"
	"github.com/srsalisbury/bouncebot/solver/bfs"
)

// Manager handles daily puzzle generation, storage, and retrieval.
type Manager struct {
	dataDir  string
	solver   *solver.Manager
	mu       sync.RWMutex
	cache    map[string]*DailyPuzzles // date -> puzzles
}

// NewManager creates a new daily puzzle manager.
func NewManager(dataDir string, solverMgr *solver.Manager) *Manager {
	return &Manager{
		dataDir: dataDir,
		solver:  solverMgr,
		cache:   make(map[string]*DailyPuzzles),
	}
}

// GetPuzzlesForDate loads puzzles for a specific date.
// Returns nil if puzzles don't exist for that date.
func (m *Manager) GetPuzzlesForDate(date string) (*DailyPuzzles, error) {
	// Check cache first
	m.mu.RLock()
	if puzzles, ok := m.cache[date]; ok {
		m.mu.RUnlock()
		return puzzles, nil
	}
	m.mu.RUnlock()

	// Load from disk
	path := m.puzzlePath(date)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read puzzle file: %w", err)
	}

	var puzzles DailyPuzzles
	if err := json.Unmarshal(data, &puzzles); err != nil {
		return nil, fmt.Errorf("failed to parse puzzle file: %w", err)
	}

	// Cache it
	m.mu.Lock()
	m.cache[date] = &puzzles
	m.mu.Unlock()

	return &puzzles, nil
}

// EnsureFuturePuzzles ensures puzzles exist for the next N days.
// Starts from yesterday UTC to ensure all timezones have puzzles for "today".
func (m *Manager) EnsureFuturePuzzles(days int) error {
	// Start from yesterday UTC to cover all timezones (e.g., UTC-12 to UTC+14)
	yesterday := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, -1)
	for i := range days + 1 { // +1 because we start from yesterday
		date := yesterday.AddDate(0, 0, i).Format("2006-01-02")
		puzzles, err := m.GetPuzzlesForDate(date)
		if err != nil {
			return fmt.Errorf("failed to check puzzles for %s: %w", date, err)
		}
		if puzzles == nil {
			log.Printf("Daily: generating puzzles for %s", date)
			if err := m.generateAndSavePuzzles(date); err != nil {
				return fmt.Errorf("failed to generate puzzles for %s: %w", date, err)
			}
		}
	}
	return nil
}

// StartGenerationWorker starts a background goroutine that ensures puzzles exist.
// It runs on startup (in background) and then daily at midnight UTC.
func (m *Manager) StartGenerationWorker(ctx context.Context, daysAhead int) {
	// Calculate time until next midnight UTC
	now := time.Now().UTC()
	nextMidnight := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	untilMidnight := nextMidnight.Sub(now)

	// Run everything in background so server can start immediately
	go func() {
		// Run immediately on startup
		if err := m.EnsureFuturePuzzles(daysAhead); err != nil {
			log.Printf("Daily: error generating puzzles on startup: %v", err)
		}

		// Wait until first midnight, then run daily
		select {
		case <-ctx.Done():
			return
		case <-time.After(untilMidnight):
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		// Run at first midnight
		if err := m.EnsureFuturePuzzles(daysAhead); err != nil {
			log.Printf("Daily: error generating puzzles: %v", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.EnsureFuturePuzzles(daysAhead); err != nil {
					log.Printf("Daily: error generating puzzles: %v", err)
				}
			}
		}
	}()
}

// puzzlePath returns the file path for a given date's puzzles.
func (m *Manager) puzzlePath(date string) string {
	// Parse date to get year/month/day
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		// Fallback to flat structure if parsing fails
		return filepath.Join(m.dataDir, "daily_puzzles", date+".json")
	}
	return filepath.Join(
		m.dataDir,
		"daily_puzzles",
		fmt.Sprintf("%d", t.Year()),
		fmt.Sprintf("%02d", t.Month()),
		fmt.Sprintf("%02d.json", t.Day()),
	)
}

// generateAndSavePuzzles generates puzzles for a date and saves them.
func (m *Manager) generateAndSavePuzzles(date string) error {
	puzzles := &DailyPuzzles{Date: date}

	// Use date as seed for reproducibility
	seed := dateToSeed(date)

	// Generate puzzles for each difficulty
	var err error
	puzzles.Easy, err = m.generatePuzzle(DifficultyEasy, seed)
	if err != nil {
		return fmt.Errorf("failed to generate easy puzzle: %w", err)
	}

	puzzles.Medium, err = m.generatePuzzle(DifficultyMedium, seed+10000)
	if err != nil {
		return fmt.Errorf("failed to generate medium puzzle: %w", err)
	}

	puzzles.Hard, err = m.generatePuzzle(DifficultyHard, seed+20000)
	if err != nil {
		return fmt.Errorf("failed to generate hard puzzle: %w", err)
	}

	// Save to disk
	path := m.puzzlePath(date)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create puzzle directory: %w", err)
	}

	data, err := json.MarshalIndent(puzzles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal puzzles: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write puzzle file: %w", err)
	}

	// Cache it
	m.mu.Lock()
	m.cache[date] = puzzles
	m.mu.Unlock()

	log.Printf("Daily: saved puzzles for %s (easy=%d, medium=%d, hard=%d moves)",
		date, puzzles.Easy.OptimalMoves, puzzles.Medium.OptimalMoves, puzzles.Hard.OptimalMoves)

	return nil
}

// generatePuzzle generates a single puzzle matching the target difficulty.
func (m *Manager) generatePuzzle(difficulty string, baseSeed int64) (*DailyPuzzle, error) {
	bfsSolver := &bfs.BFSSolver{}
	timeout := 30 * time.Second
	maxAttempts := 10000

	for attempt := range maxAttempts {
		seed := baseSeed + int64(attempt)
		game := newSeededRandomGame(seed)

		// Solve to find optimal length
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		result := bfsSolver.Solve(ctx, game)
		cancel()

		if !result.Completed || result.Solution == nil {
			continue
		}

		optimalMoves := len(result.Solution.Moves)
		puzzleDifficulty := ClassifyDifficulty(optimalMoves)

		if puzzleDifficulty == difficulty {
			// Convert solution to string format
			solutionStrings := make([]string, len(result.Solution.Moves))
			for i, move := range result.Solution.Moves {
				// Calculate direction from position change
				dir := calculateDirection(game.Bots[move.Id], move.Pos)
				solutionStrings[i] = fmt.Sprintf("B%d:%s", move.Id, dir)
			}

			return &DailyPuzzle{
				OptimalMoves: optimalMoves,
				Solution:     solutionStrings,
				Game:         strings.Split(game.String(), "\n"),
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to generate %s puzzle after %d attempts", difficulty, maxAttempts)
}

// calculateDirection determines the direction of a move based on position change.
func calculateDirection(from, to model.Position) string {
	if to.Y < from.Y {
		return "up"
	}
	if to.Y > from.Y {
		return "down"
	}
	if to.X < from.X {
		return "left"
	}
	if to.X > from.X {
		return "right"
	}
	return "none"
}

// dateToSeed converts a date string to a reproducible seed.
func dateToSeed(date string) int64 {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return 0
	}
	// Use days since epoch as seed base
	return t.Unix() / 86400
}

// newSeededRandomGame generates a random game using a specific seed.
// This is copied from solver/benchmark/puzzle_gen.go for consistency.
func newSeededRandomGame(seed int64) *model.Game {
	rng := rand.New(rand.NewSource(seed))

	// Shuffle panels 1-4 into random positions
	panels := []int{1, 2, 3, 4}
	rng.Shuffle(len(panels), func(i, j int) {
		panels[i], panels[j] = panels[j], panels[i]
	})
	board := model.BuildBoard(panels[0], panels[1], panels[2], panels[3])

	// Pick a random target from possible targets
	possibleTargets := board.PossibleTargets()
	if len(possibleTargets) == 0 {
		panic("board has no possible targets")
	}
	targetPos := possibleTargets[rng.Intn(len(possibleTargets))]
	targetBotId := model.BotId(rng.Intn(4))
	target := model.BotPosition{Id: targetBotId, Pos: targetPos}

	// Place robots randomly, avoiding:
	// - Each other
	// - The target position
	// - The center 4 cells (for a 16x16 board: (7,7), (8,7), (7,8), (8,8))
	size := board.Size()
	centerCells := []model.Position{
		{X: size/2 - 1, Y: size/2 - 1},
		{X: size / 2, Y: size/2 - 1},
		{X: size/2 - 1, Y: size / 2},
		{X: size / 2, Y: size / 2},
	}

	isOccupied := func(pos model.Position, placedBots map[model.BotId]model.Position) bool {
		if pos == targetPos {
			return true
		}
		if slices.Contains(centerCells, pos) {
			return true
		}
		for _, botPos := range placedBots {
			if pos == botPos {
				return true
			}
		}
		return false
	}

	bots := make(map[model.BotId]model.Position)
	for botId := range model.BotId(4) {
		for {
			pos := model.Position{
				X: model.BoardDim(rng.Intn(int(size))),
				Y: model.BoardDim(rng.Intn(int(size))),
			}
			if !isOccupied(pos, bots) {
				bots[botId] = pos
				break
			}
		}
	}

	game, err := model.NewGame(board, bots, target)
	if err != nil {
		panic(err)
	}
	return game
}
