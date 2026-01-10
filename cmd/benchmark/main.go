package main

import (
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/srsalisbury/bouncebot/solver/benchmark"

	// Register solvers via init()
	_ "github.com/srsalisbury/bouncebot/solver/astar_steven"
	_ "github.com/srsalisbury/bouncebot/solver/bfs"
)

func main() {
	// Parse command line flags
	difficulties := flag.String("difficulties", "5,7,10,13", "Comma-separated list of difficulty levels (optimal move counts)")
	puzzlesPerDifficulty := flag.Int("puzzles", 10, "Number of puzzles per difficulty level")
	timeout := flag.Duration("timeout", 30*time.Second, "Timeout per puzzle")
	noCache := flag.Bool("no-cache", false, "Skip cache and regenerate puzzles")
	solverName := flag.String("solver", "", "Run only this solver (default: all)")
	flag.Parse()

	// Parse difficulties
	diffList := parseDifficulties(*difficulties)

	cfg := benchmark.Config{
		Difficulties:         diffList,
		PuzzlesPerDifficulty: *puzzlesPerDifficulty,
		Timeout:              *timeout,
		NoCache:              *noCache,
		SolverName:           *solverName,
	}

	benchmark.Run(cfg)
}

func parseDifficulties(s string) []int {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if n, err := strconv.Atoi(p); err == nil {
			result = append(result, n)
		}
	}
	return result
}
