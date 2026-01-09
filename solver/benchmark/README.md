# Solver Benchmark

Compares solver performance across puzzles of varying difficulty.

## How It Works

1. **Puzzle Generation**: Generates puzzles using deterministic seeds (starting at 1000) for reproducibility
2. **Difficulty Classification**: Runs BFS to find the optimal solution length, then groups puzzles by move count
3. **Benchmarking**: Times each registered solver on the same puzzle set
4. **Results**: Outputs a comparison table showing average time, total time, solve rate, and optimal solution rate

## Usage

```bash
# From the repository root:
go run ./cmd/benchmark

# With options:
go run ./cmd/benchmark -puzzles=10 -difficulties=5,7,10,13 -timeout=30s
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `-difficulties` | `5,7,10,13` | Comma-separated list of difficulty levels (optimal move counts) |
| `-puzzles` | `10` | Number of puzzles per difficulty level |
| `-timeout` | `30s` | Timeout per puzzle solve |
| `-no-cache` | `false` | Skip cache and regenerate puzzles |

## Puzzle Caching

Generated puzzles are cached in `solver/benchmark/puzzles/` as JSON files. This avoids regeneration on subsequent runs, which can take significant time for higher difficulty puzzles.

- Cache files are named `difficulty_N.json` (e.g., `difficulty_13.json`)
- Use `-no-cache` flag to regenerate puzzles
- Delete cache files manually to force regeneration for specific difficulties

## Example Output

```
Solver Benchmark
================

Registered solvers: bfs, bfs2

Generating puzzles... done

Difficulty: 5 moves (10 puzzles)
┌──────────┬────────────┬────────────┬──────────┬───────────┐
│ Solver   │ Avg Time   │ Total      │ Solved   │ Optimal   │
├──────────┼────────────┼────────────┼──────────┼───────────┤
│ bfs      │     25.5ms │    255.0ms │    10/10 │     10/10 │
│ bfs2     │     23.2ms │    232.0ms │    10/10 │     10/10 │
└──────────┴────────────┴────────────┴──────────┴───────────┘
```

## Adding New Solvers

Solvers are automatically included if they register with the default registry via `init()`. To add a new solver to the benchmark:

1. Create your solver package (e.g., `solver/mysolver/`)
2. Register it in `init()`: `solver.Register(&MySolver{})`
3. Import it in `cmd/benchmark/main.go`: `_ "github.com/srsalisbury/bouncebot/solver/mysolver"`

## Reproducibility

The benchmark uses deterministic seeding, so running with the same parameters will produce the same puzzles. This allows for fair comparison across different solver implementations or code changes.
