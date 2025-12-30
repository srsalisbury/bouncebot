# Backend Development

Implementation notes for developers working on the Go backend. For project overview and how to run, see the root [README.md](../README.md).

## Code Layout

```
server/
├── main.go             # HTTP server setup, RPC handlers, CORS, WebSocket endpoint
├── config/
│   └── config.go       # Server configuration (ports, persistence settings)
├── room/               # Multiplayer room management
│   ├── service.go      # RoomService orchestrator (main entry point)
│   ├── repository.go   # RoomRepository - CRUD with per-room locking
│   ├── player_manager.go    # PlayerManager - connection state
│   ├── game_lifecycle_manager.go  # GameLifecycle - game state transitions
│   ├── solution_manager.go  # SolutionManager - solution submission/retraction
│   ├── timer_manager.go     # TimerManager - disconnect grace timers
│   ├── persistence_manager.go  # PersistenceManager - save/load/cleanup
│   ├── signals.go      # Signal types for component communication
│   ├── room.go         # Room struct and helpers
│   ├── player.go       # Player struct, PlayerStatus
│   ├── solution.go     # PlayerSolution structs
│   └── *_test.go       # Unit tests per component + integration tests
└── ws/                 # WebSocket real-time events
    ├── hub.go          # Connection hub, event broadcasting
    └── hub_test.go

model/                  # Core game logic (no server dependencies)
├── position.go         # Position, BoardDim types
├── board.go            # Board interface, walls, possible targets
├── game.go             # Game struct, robot movement, validation
├── games.go            # Game generation (random, continuation)
├── render.go           # Board parsing from string representation
├── physics_test.go     # Shared physics test fixtures
└── *_test.go

proto/                  # Protocol buffer definitions
├── bouncebot.proto     # Message and RPC definitions
├── bouncebot.pb.go     # Generated Go types
├── bouncebot_grpc.pb.go
├── protoconnect/       # Generated Connect handlers
└── compile_protos.sh   # Regenerate Go code
```

## Architecture

The `room` package uses a component-based architecture with signal-driven communication:

```
┌─────────────────────────────────────────────────────────────┐
│                   RoomService (Orchestrator)                 │
│  - Receives API calls                                        │
│  - Delegates to components                                   │
│  - Processes returned signals                                │
│  - Coordinates cross-cutting concerns                        │
└─────────────────────────────────────────────────────────────┘
         │           │            │            │           │
         ▼           ▼            ▼            ▼           ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
   │   Room   │ │  Player  │ │   Game   │ │ Solution │ │  Timer   │
   │Repository│ │ Manager  │ │ Lifecycle│ │ Manager  │ │ Manager  │
   └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘
```

Components return **signals** instead of calling each other directly. The orchestrator interprets signals and coordinates actions:

```go
// Components return signals indicating what should happen
signals, err := playerMgr.RemovePlayer(room, playerID)
// -> may return EndGameSignal if all remaining players are finished

// Orchestrator processes signals
for _, sig := range signals {
    switch s := sig.(type) {
    case EndGameSignal:
        newSignals := gameMgr.EndGame(room)
        // process recursively...
    case BroadcastSignal:
        broadcaster.Broadcast(s.Event)
    }
}
```

## Key Packages

### `model/` - Game Logic
Pure game logic with no server dependencies. Can be tested independently.

- **Board**: 16x16 grid with walls, possible target positions
- **Game**: Robot positions, target, move validation, physics
- **Direction**: Up, Down, Left, Right movement
- **ComputeDestination**: Calculate where robot stops when sliding

### `server/room/` - Room Management
Multiplayer room state and operations, organized into components:

| Component | File | Responsibility |
|-----------|------|----------------|
| **RoomService** | `service.go` | Orchestrator - coordinates all components |
| **RoomRepository** | `repository.go` | CRUD operations with per-room locking |
| **PlayerManager** | `player_manager.go` | Add/disconnect/reconnect/remove players |
| **GameLifecycle** | `game_lifecycle_manager.go` | Start/end games, mark finished/ready |
| **SolutionManager** | `solution_manager.go` | Submit/retract solutions, determine winner |
| **TimerManager** | `timer_manager.go` | Disconnect grace period timers |
| **PersistenceManager** | `persistence_manager.go` | Save/load rooms, cleanup stale rooms |

### `server/ws/` - WebSocket Hub
Real-time event broadcasting to connected clients.

**Events broadcast:**
- `player_joined` - New player entered room
- `player_left` - Player disconnected
- `game_started` - New game began
- `player_solved` - Player submitted solution
- `solution_retracted` - Player retracted solution
- `player_finished_solving` - Player marked done
- `game_ended` - All players finished, winner determined

## RPC Endpoints

Defined in `proto/bouncebot.proto`, handled in `server/main.go`:

| RPC | Description |
|-----|-------------|
| `CreateRoom` | Create new room, returns room with player added |
| `JoinRoom` | Join existing room by ID |
| `GetRoom` | Get current room state |
| `StartGame` | Start new game (random or fixed board) |
| `SubmitSolution` | Submit solution moves (server validates) |
| `RetractSolution` | Retract submitted solution |
| `MarkFinishedSolving` | Player is done looking for solutions |
| `MarkReadyForNext` | Player ready for next game |

## Conventions

### Error Handling
- Return `connect.NewError(code, err)` for RPC errors
- Use `connect.CodeNotFound`, `connect.CodeInvalidArgument`, etc.

### Thread Safety
- `RoomRepository` uses per-room locking via `GetWithLock()`
- Each room operation locks only that room
- Timer callbacks and other rooms can proceed concurrently

### Testing
- Unit tests per component (e.g., `repository_test.go`, `player_manager_test.go`)
- Integration tests in `service_test.go`
- Shared test utilities in `helpers_test.go`
- Table-driven tests preferred
- Shared physics fixtures in `tests/physics_cases.json`
- Run all tests: `go test ./...`
