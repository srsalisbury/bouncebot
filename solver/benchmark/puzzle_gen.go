package benchmark

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/srsalisbury/bouncebot/model"
	"github.com/srsalisbury/bouncebot/solver/bfs"
)

// Puzzle represents a generated puzzle with its known optimal solution length.
type Puzzle struct {
	Game         *model.Game
	OptimalMoves int
	Seed         int64
}

// GeneratePuzzles generates puzzles grouped by difficulty (optimal solution length).
// Returns a map from difficulty (move count) to slice of puzzles.
// If useCache is true, loads from cache when available and saves newly generated puzzles.
func GeneratePuzzles(difficulties []int, puzzlesPerDifficulty int, timeout time.Duration, useCache bool) map[int][]Puzzle {
	result := make(map[int][]Puzzle)
	needsGeneration := make([]int, 0)

	// Try to load from cache first
	if useCache {
		for _, d := range difficulties {
			cached := LoadCachedPuzzles(d)
			if len(cached) >= puzzlesPerDifficulty {
				result[d] = cached[:puzzlesPerDifficulty]
			} else {
				result[d] = make([]Puzzle, 0, puzzlesPerDifficulty)
				needsGeneration = append(needsGeneration, d)
			}
		}
	} else {
		for _, d := range difficulties {
			result[d] = make([]Puzzle, 0, puzzlesPerDifficulty)
			needsGeneration = append(needsGeneration, d)
		}
	}

	// If all difficulties are loaded from cache, we're done
	if len(needsGeneration) == 0 {
		return result
	}

	// Use BFS solver to determine optimal solution length
	bfsSolver := &bfs.BFSSolver{}

	// Start from a known seed base for reproducibility
	seed := int64(1000)
	maxAttempts := puzzlesPerDifficulty * 100 * len(needsGeneration)

	lastProgressPrint := time.Now()
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check if we have enough puzzles for all needed difficulties
		allFilled := true
		for _, d := range needsGeneration {
			if len(result[d]) < puzzlesPerDifficulty {
				allFilled = false
				break
			}
		}
		if allFilled {
			break
		}

		// Print progress every 5 seconds
		if time.Since(lastProgressPrint) > 5*time.Second {
			fmt.Printf("\n  Progress: ")
			for _, d := range needsGeneration {
				fmt.Printf("d%d=%d/%d ", d, len(result[d]), puzzlesPerDifficulty)
			}
			lastProgressPrint = time.Now()
		}

		// Generate a puzzle with this seed
		game := newSeededRandomGame(seed)
		seed++

		// Solve to find optimal length
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		solveResult := bfsSolver.Solve(ctx, game)
		cancel()

		if !solveResult.Completed || solveResult.Solution == nil {
			continue // Skip unsolvable or timed-out puzzles
		}

		optimalMoves := len(solveResult.Solution.Moves)

		// Check if this puzzle fits a difficulty bucket we need
		for _, d := range needsGeneration {
			if optimalMoves == d && len(result[d]) < puzzlesPerDifficulty {
				result[d] = append(result[d], Puzzle{
					Game:         game,
					OptimalMoves: optimalMoves,
					Seed:         seed - 1, // The seed we just used
				})
				break
			}
		}
	}

	// Save newly generated puzzles to cache
	if useCache {
		for _, d := range needsGeneration {
			if len(result[d]) > 0 {
				if err := SaveCachedPuzzles(d, result[d]); err != nil {
					fmt.Printf("Warning: failed to cache puzzles for difficulty %d: %v\n", d, err)
				}
			}
		}
	}

	return result
}

// newSeededRandomGame generates a random game using a specific seed.
// This is a seeded version of model.NewRandomGame().
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
		for _, center := range centerCells {
			if pos == center {
				return true
			}
		}
		for _, botPos := range placedBots {
			if pos == botPos {
				return true
			}
		}
		return false
	}

	bots := make(map[model.BotId]model.Position)
	for botId := model.BotId(0); botId < 4; botId++ {
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
