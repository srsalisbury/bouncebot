# Solver Framework Implementation Plan

## Overview

Add a framework for multiple board solvers that run asynchronously, with configurable timeouts and results collected for comparison.

## Requirements

- **Async execution**: Solvers run in background, client polls for results
- **All results**: Wait for all solvers, return all solutions for comparison
- **Configurable timeout**: Per-solver timeout, returns best-so-far or fails
- **Multiple implementations**: Each solver in `solver/<name>/` subdirectory
- **Common interface**: Server can run any registered solver

---

- [ ] Room-wide settings (set by host)
  - [ ] Choose a solver
  - [ ] Show solver move count
  - [ ] Show solver solution (in multiplayer, it looks like another player but clearly distinguished, in single player, it shows before you go to the next puzzle in an end-game view like in multiplayer mode with your best and solver best solutions)

---

## Phase 1: Core Framework

### 1.1 Create solver interface and types

**File**: `solver/solver.go`

```go
type Solution struct {
    Moves    []model.BotPosition
}

type Result struct {
    SolverName   string
    Solution     *Solution
    Error        error
    Completed    bool
}

type Solver interface {
    Name() string
    Solve(ctx context.Context, game *model.Game) Result
}
```

### 1.2 Create solver registry

**File**: `solver/registry.go`

- Thread-safe registry with `Register()`, `Get()`, `All()`
- Global `DefaultRegistry` for convenience
- Solvers register via `init()` functions

### 1.3 Create async solver manager

**File**: `solver/manager.go`

- `Manager` struct with mutex-protected job storage
- `StartJob(roomID, game, timeout, solverName)` launches one solver
- Solver runs in its own goroutine with context timeout
- Callback on solver completion for WebSocket events
- `GetJob(jobID)` and `GetJobByRoom(roomID)` for retrieval

---

## Phase 2: First Solver Implementation

### 2.1 BFS Solver

**File**: `solver/bfs/bfs.go`

- Breadth-first search for optimal (shortest) solution
- State: bot positions as hashable key
- Transitions: for each bot, for each direction, compute destination via `game.ComputeDestination()`
- Track visited states to avoid cycles
- Check `ctx.Done()` periodically for timeout responsiveness
- Return first solution found (optimal by move count)

---

## Phase 3: Proto and RPC

### 3.1 Add proto definitions

**File**: `proto/bouncebot.proto`

Add messages:
- `SolverSolution` (moves)
- `SolverResult` (solver_name, solution, error, completed)

### 3.2 Regenerate proto code

Run `./proto/compile_protos.sh` and `npm run generate` in client

### 3.3 Add Solvers to manager

**File**: `server/main.go`

- Import solver packages with `_` for init registration
- Create `solver.Manager` in main

---

## Phase 4: WebSocket Integration

### 4.1 Add WebSocket events

**File**: `server/ws/hub.go`

Add broadcast methods:
- `BroadcastSolverComplete(roomID, solverResult)` - fired when solver is done

### 4.2 Update EventBroadcaster interface

**File**: `server/room/room.go`

Add methods to `EventBroadcaster` interface (or create separate interface for solver events)

---

## Phase 5: Client Integration

- Listen for WebSocket events or poll `GetSolverResult`
- Animated playback of solutions

---

## Files to Create

| File | Purpose |
|------|---------|
| `solver/solver.go` | Interface and types |
| `solver/registry.go` | Solver registration |
| `solver/manager.go` | Async job management |
| `solver/bfs/bfs.go` | BFS solver implementation |

## Files to Modify

| File | Changes |
|------|---------|
| `proto/bouncebot.proto` | Add solver messages and RPCs |
| `server/main.go` | Import solvers, create manager, add handlers |
| `server/ws/hub.go` | Add solver broadcast methods |
| `server/room/room.go` | Extend EventBroadcaster interface |

## Key Dependencies

- `model/game.go` - `Game.ComputeDestination()`, `Game.CheckSolution()`
- `model/game.go:13-16` - `BotPosition` struct for solution moves
- `server/room/room.go` - Patterns for mutex-protected stores

---

## Implementation Order

1. `solver/solver.go` - Define interface
2. `solver/registry.go` - Registration system
3. `solver/manager.go` - Async execution
4. `solver/bfs/bfs.go` - First working solver
5. `proto/bouncebot.proto` - API definitions
6. Regenerate proto code
7. `server/main.go` - Wire up handlers
8. `server/ws/hub.go` - Events
