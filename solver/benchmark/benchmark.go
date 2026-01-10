package benchmark

import (
	"context"
	"fmt"
	"time"

	"github.com/srsalisbury/bouncebot/solver"
)

// Config holds benchmark configuration.
type Config struct {
	Difficulties         []int         // Target solution lengths to benchmark
	PuzzlesPerDifficulty int           // Number of puzzles per difficulty
	Timeout              time.Duration // Timeout per puzzle solve
	NoCache              bool          // If true, skip cache and regenerate puzzles
	SolverName           string        // If set, only run this solver (default: all)
}

// DefaultConfig returns a reasonable default configuration.
func DefaultConfig() Config {
	return Config{
		Difficulties:         []int{5, 7, 10, 13},
		PuzzlesPerDifficulty: 10,
		Timeout:              30 * time.Second,
	}
}

// Result holds the benchmark result for one solver at one difficulty.
type Result struct {
	SolverName   string
	Difficulty   int
	PuzzleCount  int
	TotalTime    time.Duration
	AvgTime      time.Duration
	SolvedCount  int
	OptimalCount int
}

// Run executes the benchmark with the given configuration.
func Run(cfg Config) {
	fmt.Println("Solver Benchmark")
	fmt.Println("================")
	fmt.Println()

	// Get all registered solvers
	allSolvers := solver.DefaultRegistry.All()
	if len(allSolvers) == 0 {
		fmt.Println("No solvers registered!")
		return
	}

	// Filter to specific solver if requested
	var solvers []solver.Solver
	if cfg.SolverName != "" {
		for _, s := range allSolvers {
			if s.Name() == cfg.SolverName {
				solvers = append(solvers, s)
				break
			}
		}
		if len(solvers) == 0 {
			fmt.Printf("Solver %q not found. Available: ", cfg.SolverName)
			for i, s := range allSolvers {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(s.Name())
			}
			fmt.Println()
			return
		}
	} else {
		solvers = allSolvers
	}

	fmt.Printf("Registered solvers: ")
	for i, s := range solvers {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(s.Name())
	}
	fmt.Println()
	fmt.Println()

	// Generate puzzles (use cache by default unless NoCache is set)
	fmt.Print("Loading/generating puzzles...")
	puzzles := GeneratePuzzles(cfg.Difficulties, cfg.PuzzlesPerDifficulty, cfg.Timeout, !cfg.NoCache)
	fmt.Println(" done")
	fmt.Println()

	// Run benchmark for each difficulty
	for _, difficulty := range cfg.Difficulties {
		puzzleSet := puzzles[difficulty]
		if len(puzzleSet) == 0 {
			fmt.Printf("Difficulty: %d moves - No puzzles generated (difficulty too rare?)\n\n", difficulty)
			continue
		}

		fmt.Printf("Difficulty: %d moves (%d puzzles)\n", difficulty, len(puzzleSet))

		results := make([]Result, len(solvers))
		for i, s := range solvers {
			results[i] = benchmarkSolver(s, puzzleSet, cfg.Timeout)
		}

		printResultsTable(results)
		fmt.Println()
	}
}

// benchmarkSolver runs a single solver against a set of puzzles.
func benchmarkSolver(s solver.Solver, puzzles []Puzzle, timeout time.Duration) Result {
	result := Result{
		SolverName:  s.Name(),
		Difficulty:  puzzles[0].OptimalMoves,
		PuzzleCount: len(puzzles),
	}

	for _, puzzle := range puzzles {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		start := time.Now()
		solveResult := s.Solve(ctx, puzzle.Game)
		elapsed := time.Since(start)
		cancel()

		result.TotalTime += elapsed

		if solveResult.Completed && solveResult.Solution != nil {
			result.SolvedCount++
			foundMoves := len(solveResult.Solution.Moves)
			if foundMoves == puzzle.OptimalMoves {
				result.OptimalCount++
			} else {
				fmt.Printf("  [%s] puzzle %d: found %d moves, optimal %d\n",
					s.Name(), puzzle.Seed, foundMoves, puzzle.OptimalMoves)
			}
		}
	}

	if result.PuzzleCount > 0 {
		result.AvgTime = result.TotalTime / time.Duration(result.PuzzleCount)
	}

	return result
}
