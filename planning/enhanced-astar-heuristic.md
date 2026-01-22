# Enhanced A* Solver (astar2) Implementation Plan

## Goal
Create a new A* solver with an improved heuristic that adds +1 when ALL shortest paths have imperfections (blocking bots or missing stoppers). This provides a tighter admissible heuristic.

## Background
- Current heuristic: Optimistic distance assuming no bot obstacles
- Enhanced heuristic: Same base distance, but +1 if every shortest path has issues
- A path is "imperfect" if:
  1. A bot blocks a cell the target bot slides through, OR
  2. A turning point has no wall and no bot to stop the target bot

## File Structure
```
solver/astar2/
  astar2.go                  # Main solver (modified copy of astar.go)
  enhanced_heuristic.go      # Path-tracking heuristic
  enhanced_heuristic_test.go # Tests for enhanced heuristic
  priority_queue.go          # Copy from astar (unchanged)
  astar2_test.go             # Solver tests
```

## Implementation Steps

### Step 1: Create Package Structure
- Create `solver/astar2/` directory
- Copy `priority_queue.go` from `solver/astar/`

### Step 2: Define Data Structures (`enhanced_heuristic.go`)
```go
type PathRequirement struct {
    // MustBeClear: stored as a set (precomputed) for O(1) bot position lookups at runtime.
    // At runtime, iterate ~3 other bots and check set membership.
    MustBeClear map[model.Position]bool

    // MustHaveBot: stored as a slice since we need to check each position individually
    // to verify a bot exists there.
    MustHaveBot []model.Position
}

type CellInfo struct {
    Distance int
    Paths    []PathRequirement
}

type EnhancedHeuristicTable struct {
    cells [256]CellInfo
    size  int
}
```

### Step 3: Implement Enhanced BFS (`enhanced_heuristic.go`)
Two-phase approach:

**Phase 1: BFS with predecessor tracking**
- Run reverse BFS from target (like current heuristic)
- Track ALL predecessors that reach each cell at minimum distance
- Store: fromPos, direction, and whether stop is at wall

**Phase 2: Path enumeration**
- For each cell, walk predecessor graph to enumerate all shortest paths
- Build `PathRequirement` for each path:
  - `MustBeClear`: build as `map[Position]bool` set containing cells between each slide's start and stop
  - `MustHaveBot`: slice of stopper positions where no wall exists (one cell past stop in slide direction)
- Limit to ~100 paths per cell to bound memory. **CRITICAL:** If limit is reached, mark cell as "complex" (e.g. `TooManyPaths = true`); runtime check must return base `Distance` (no +1 penalty) to preserve admissibility.

### Step 4: Implement Runtime Evaluation
```go
func (ht *EnhancedHeuristicTable) Get(
    targetBotPos model.Position,
    allBots map[model.BotId]model.Position,
    targetBotId model.BotId,
) int {
    cellInfo := ht.cells[posToIndex(targetBotPos)]

    // If we hit the path limit during precomputation, we can't be sure all paths are blocked.
    // Return base distance to remain admissible.
    if cellInfo.TooManyPaths {
        return cellInfo.Distance
    }

    // Check if ANY path is viable
    for _, path := range cellInfo.Paths {
        if isPathViable(path, allBots, targetBotId) {
            return cellInfo.Distance  // Perfect path exists
        }
    }
    return cellInfo.Distance + 1  // All paths imperfect
}

func isPathViable(path PathRequirement, bots map[BotId]Position, targetId BotId) bool {
    // 1. Check MustBeClear: iterate ~3 other bots, O(1) set lookup each
    for botId, botPos := range bots {
        if botId != targetId && path.MustBeClear[botPos] {
            return false  // Bot blocking the path
        }
    }

    // 2. Check MustHaveBot: for each required stopper, verify a bot exists there
    for _, needPos := range path.MustHaveBot {
        found := false
        for botId, botPos := range bots {
            if botId != targetId && botPos == needPos {
                found = true
                break
            }
        }
        if !found {
            return false  // Missing required stopper
        }
    }

    return true
}
```

### Step 5: Create Solver (`astar2.go`)
- Copy `astar.go` to `astar2.go`
- Change package to `astar2`
- Change solver name to `"A-Star-2"`
- Replace `NewHeuristicTable` with `NewEnhancedHeuristicTable`
- Update heuristic calls: `heuristic.Get(pos, bots, targetBotId)`

### Step 6: Register Solver
- Add `init()` function calling `solver.Register(&AStar2Solver{})`
- Add blank import to `cmd/benchmark/main.go`:
  ```go
  _ "github.com/srsalisbury/bouncebot/solver/astar2"
  ```

### Step 7: Write Tests
- `enhanced_heuristic_test.go`:
  - Test path enumeration correctness
  - Test blocked path detection (+1)
  - Test missing stopper detection (+1)
  - Test perfect path exists (base distance)
- `astar2_test.go`:
  - Verify optimal solutions still found
  - Compare with astar solver results

## Key Helper Functions Needed
- `oppositeDirection(dir)` - reverse direction
- `hasWallInDirection(game, pos, dir)` - check wall at position
- `getStopperPosition(stopPos, slideDir)` - compute P+1 position
- `getCellsBetween(a, b)` - cells strictly between two positions
- `collectReverseSlidePath(game, pos, dir)` - all source positions for a slide

## Critical Files to Reference
- `/Users/mike/dev/bouncebot/solver/astar/astar.go` - A* logic to copy
- `/Users/mike/dev/bouncebot/solver/astar/heuristic.go` - BFS pattern to extend
- `/Users/mike/dev/bouncebot/solver/solver.go` - Solver interface
- `/Users/mike/dev/bouncebot/model/game.go` - Game types

## Verification

### Unit Tests
```bash
go test ./solver/astar2/... -v
```

### Benchmark Comparison
```bash
# Run both solvers on same puzzles
go run cmd/benchmark/main.go
```

### Manual Verification
- Run both solvers on Game1
- Verify same solution length (both optimal)
- Compare node expansion counts (astar2 should be lower or equal)
- Log/print heuristic values to verify +1 applied correctly

## Edge Cases
1. Target bot already at target → distance 0, empty paths
2. Unreachable cells → distance -1
3. Stopper position off-board → shouldn't happen (wall stops first)
4. Too many paths → limit with MaxPathsPerCell constant
