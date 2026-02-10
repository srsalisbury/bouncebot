# BounceBot Server Documentation

The BounceBot server is a Go application that manages multiplayer game state, real-time communication, and game logic solving. It uses **Connect** (gRPC-compatible) for transactional API calls and **WebSockets** for real-time state updates.

## Getting Started

### Prerequisites
- Go 1.24+
- Protocol Buffers compiler (if modifying `.proto` files)

### Running Locally
You can run the server directly from the `server` directory or the project root.

```bash
# From project root
go run server/main.go --port 8080
```

**Configuration Flags:**
- `--port`: Port to listen on (default: `8080` or `PORT` env var).

All data files (rooms, daily puzzles, user progress) are stored under `DATA_DIR` (default: `data`).

## Project Structure

```
server/
├── main.go               # Application entry point, wiring, and HTTP server
├── bouncebotserver.go    # Connect RPC handler implementation
├── config/               # Configuration loading
├── room/                 # Core domain logic and room management
│   ├── service.go        # RoomService (Orchestrator)
│   ├── repository.go     # Thread-safe room storage
│   ├── *_manager.go      # Logic components (Player, Game, Solution, etc.)
│   └── signals.go        # Internal event system
└── ws/                   # WebSocket hub and event broadcasting
```

## Architecture

The server uses a **Component-Based Orchestrator** pattern within the `room` package to manage complexity.

### The Orchestrator (`RoomService`)
The `RoomService` (`server/room/service.go`) is the single entry point for all room modifications. It does not contain deep business logic itself but coordinates specialized components:

1.  **RoomRepository**: Manages concurrent access to room data.
2.  **PlayerManager**: Handles joining, leaving, and connection status.
3.  **GameLifecycle**: Manages game state (waiting -> playing -> finished).
4.  **SolutionManager**: Validates and ranks player solutions.
5.  **PersistenceManager**: Handles saving/loading rooms to disk.

### The Signal System
To keep components decoupled, they do not call each other directly. Instead, they return **Signals** (e.g., `PlayerJoinedSignal`, `GameEndedSignal`) to the `RoomService`. The service then decides what to do next, such as broadcasting an event or triggering another component.

### Concurrency Model
- **Per-Room Locking**: `RoomRepository` provides `GetWithLock(id)`, ensuring only one operation happens on a specific room at a time.
- **Global Safety**: Operations across different rooms run in parallel.
- **Background Jobs**: Solvers and cleanup tasks run in separate goroutines.

## Key Systems

### Real-Time Updates (`ws`)
Clients connect via WebSocket to `/ws?room_id=...&player_id=...`. The `ws.Hub` broadcasts events to all players in a room when state changes.

**Common Events:**
- `player_joined`, `player_left`
- `game_started`, `game_ended`
- `player_solved`: A player submitted a valid solution.
- `solver_complete`: The server-side bot found a solution.

### Solver Integration
The server runs an automated solver (A*) alongside human players to verify puzzle difficulty and provide hints.
1.  **Trigger**: On `GameStart`, `main.go` triggers registered solvers.
2.  **Execution**: Solvers run in background goroutines (managed by `solver/manager.go`).
3.  **Completion**: Results are sent back to `main` via callback, saved to the room, and broadcasted to clients.

### Persistence
Game state is periodically saved to a JSON file (specified by `--data`).
- **Auto-save**: Occurs on a configurable interval.
- **Graceful Shutdown**: State is always saved when the server receives `SIGINT` or `SIGTERM`.
- **Cleanup**: Stale rooms (inactive for >24h) are automatically removed.

## API Reference

### Connect RPCs
Defined in `proto/bouncebot.proto`. These handle all user actions.

| RPC | Description |
| :--- | :--- |
| `CreateRoom` | Creates a new room (Multiplayer or Single Player). |
| `JoinRoom` | Adds a player to an existing room. |
| `GetRoom` | Fetches the full current state of a room. |
| `StartGame` | Generates a new board and starts the timer. |
| `SubmitSolution` | Validates moves. Returns success if valid. |
| `MarkFinishedSolving`| Player gives up or is done improving their solution. |
| `MarkReadyForNext` | Player votes to start the next round. |
| `UpdateRoomSettings` | Host updates settings (e.g., show solver moves). |
| `BootPlayer` | Host removes a player from the room. |

## Development Workflow

### specific Tests
Run unit and integration tests for the server:

```bash
go test ./server/...
```

The `server/room` package contains extensive tests verifying the complex state transitions of the game lifecycle.

### Regenerating Protobufs
If you modify `proto/bouncebot.proto`, regenerate the Go code:

```bash
cd proto
./compile_protos.sh
```