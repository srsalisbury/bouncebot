# Backend Development

Implementation notes for developers working on the Go backend. For project overview and how to run, see the root [README.md](../README.md).

## Code Layout

```
server/
├── main.go             # HTTP server setup, RPC handlers, CORS, WebSocket endpoint
├── config/
│   └── config.go       # Server configuration (ports, persistence settings)
├── room/               # Multiplayer room management
│   ├── room.go         # Room struct, Store, Create/Join/Get operations
│   ├── player.go       # Player management, disconnect handling
│   ├── solution.go     # Solution submission, retraction, history
│   ├── lifecycle.go    # Game lifecycle (start, end, continuation)
│   ├── persistence.go  # JSON file save/load, auto-save, cleanup
│   └── *_test.go
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

## Key Packages

### `model/` - Game Logic
Pure game logic with no server dependencies. Can be tested independently.

- **Board**: 16x16 grid with walls, possible target positions
- **Game**: Robot positions, target, move validation, physics
- **Direction**: Up, Down, Left, Right movement
- **ComputeDestination**: Calculate where robot stops when sliding

### `server/room/` - Room Management
Multiplayer room state and operations.

- **Room**: Players, current game, solutions, wins tracking
- **Store**: Thread-safe room storage with persistence
- **Solution history**: Tracks all solutions for retraction support
- **Lifecycle**: Game start, end detection, continuation games

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
- `Store` uses `sync.RWMutex` for concurrent access
- Room modifications happen inside lock

### Testing
- Table-driven tests preferred
- Shared physics fixtures in `tests/physics_cases.json`
- Run all tests: `go test ./...`
