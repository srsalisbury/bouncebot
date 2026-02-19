# Server Deep Dive

The Go server handles room management, game logic, real-time communication, and daily challenge generation. It uses [Connect RPC](https://connectrpc.com/) for the API layer and raw WebSockets for real-time events.

For the high-level architecture and how the server fits into the overall system, see [architecture.md](architecture.md).

## Entry Point and Wiring

`server/main.go` bootstraps the server:

1. Loads configuration from environment variables (via `config.LoadFromEnv()`)
2. Creates the `RoomService` (core game orchestrator)
3. Creates the WebSocket `Hub` and wires it as the `EventBroadcaster`
4. Creates the `SolverManager`, wires `onGameStart` callback to launch solver jobs, and wires completion callback to store results and broadcast
5. Optionally creates `DailyManager` and `DailyProgressManager` (if `ENABLE_DAILY_CHALLENGE=true`)
6. Loads persisted rooms from disk
7. Starts background workers (auto-save, cleanup, daily puzzle generation)
8. Registers HTTP handlers: Connect RPC service, WebSocket endpoint (`/ws`)
9. Starts an HTTP/2 server with h2c support
10. Handles graceful shutdown on SIGINT/SIGTERM (final save, stop workers, stop timers)

The server uses `golang.org/x/net/http2/h2c` for HTTP/2 cleartext, which Connect RPC clients may use.

## RPC Handler Reference

`server/bouncebotserver.go` implements all RPCs. Each handler validates inputs, delegates to `RoomService` or the daily challenge managers, and returns Connect RPC errors with appropriate status codes.

| RPC | Auth | Delegates To | Notes |
|-----|------|-------------|-------|
| `CreateRoom` | None | `RoomService.Create()` | Returns room + player ID + session token |
| `JoinRoom` | None | `RoomService.Join()` | Returns room + player ID + session token |
| `GetRoom` | None | `RoomService.Get()` | Rate-limited: 100/min per IP |
| `StartGame` | None | `RoomService.StartGame()` | Triggers solver via callback |
| `SubmitSolution` | Session token | `RoomService.SubmitSolution()` | Validates moves by simulation |
| `MarkFinishedSolving` | Session token | `RoomService.MarkFinishedSolving()` | |
| `MarkReadyForNext` | Session token | `RoomService.MarkReadyForNext()` | |
| `UpdateRoomSettings` | Session token (host) | `RoomService.UpdateRoomSettings()` | Host-only |
| `BootPlayer` | Session token (host) | `RoomService.BootPlayer()` | Host-only |
| `LeaveRoom` | Session token | `RoomService.LeaveRoom()` | |
| `GetDailyChallenge` | Player ID | `DailyManager.GetPuzzlesForDate()` | Timezone-aware date calculation |
| `SubmitDailySolution` | Player ID | `ProgressManager.MarkSolved()` | Validates solution by simulation |
| `GetServerInfo` | None | Config check | Returns feature flags |

Session token validation: the handler calls `RoomService.ValidateSessionToken()` which looks up the token in the repository and returns the associated player ID, or an error.

## Room Orchestration (`server/room/`)

### RoomService

`server/room/service.go` is the central facade. It holds references to all managers and coordinates operations using the `withRoomLock` pattern:

```
withRoomLock(roomID, func(room *Room) ([]Signal, error) {
    // Operate on room while holding lock
    // Return signals describing side effects
})
// Lock released, signals processed
```

This pattern ensures:
- Room state is only modified under lock
- The lock is released before processing signals (which may need to acquire locks themselves)
- Errors prevent signal processing
- Deadlocks are avoided through consistent lock ordering

### Signal System

Components return `Signal` values instead of calling each other directly. After the room lock is released, `processSignals()` in `server/room/service_signals.go` interprets each signal:

| Signal Type | Processing |
|-------------|------------|
| `BroadcastSignal` | Dispatches to `EventBroadcaster` (WebSocket hub) by event type |
| `EndGameSignal` | Calls `gameMgr.EndGame()` under lock, which determines the winner and emits `GameEndedEvent` |
| `StartNextGameSignal` | Calls `gameMgr.StartNextGame()` under lock, promotes pending players, generates next puzzle, fires `onGameStart` callback |
| `StartTimerSignal` | Starts a disconnect grace period timer (duration depends on solo vs multiplayer) |
| `CancelTimerSignal` | Cancels a player's disconnect timer |

Signal processing is **recursive**: processing an `EndGameSignal` may produce `BroadcastSignal`s, which are then processed in the same loop.

### PlayerManager

`server/room/player_manager.go` manages player lifecycle:

- **AddPlayer**: Creates player with name and color (8 colors, wrapping). If a game is active, player goes to `PendingPlayers` instead of active players.
- **DisconnectPlayer**: Marks player as disconnected, emits `StartTimerSignal` for grace period.
- **ReconnectPlayer**: Restores connected status, emits `CancelTimerSignal`.
- **RemovePlayer**: Physically removes a disconnected player. Cleans up their entries in `FinishedSolving`, `ReadyForNext`, and `Solutions`. May emit `EndGameSignal` (if all remaining players are now finished) or `StartNextGameSignal` (if all remaining are ready).
- **ForceRemovePlayer**: Same as remove but works regardless of connection status. Used by host boot and leave operations.

### GameLifecycle

`server/room/game_lifecycle_manager.go` manages the game state machine:

```
No Game ──StartGame──> Playing ──EndGame──> Finished ──StartNextGame──> Playing
                                                    └── (no ready) ──> Waiting
```

Key behaviors:
- **StartGame**: Generates a random game (random board, random robot positions, random target). Emits `GameStartedEvent`.
- **MarkFinishedSolving**: Records player as finished. If all active players are finished, emits `EndGameSignal`.
- **EndGame**: Determines winner (fewest moves, then earliest submission). Credits a win to the winner. Emits `GameEndedEvent` with winner info and winning moves.
- **StartNextGame**: Promotes `PendingPlayers` to active. Generates a continuation game: same board, robots at positions from winning solution, new random target. Emits `GameStartedEvent`.
- **MarkReadyForNext**: Records player as ready. If all active players are ready, emits `StartNextGameSignal`.

### SolutionManager

`server/room/solution_manager.go` handles solution submission:

- Validates the solution against the current game using `game.CheckSolution()`
- If the player already submitted, only accepts improvements (fewer moves)
- Non-improvements are silently ignored (no broadcast)
- New or improved solutions emit `PlayerSolvedEvent` with the move count

Winner determination: fewest moves wins. Ties broken by earliest submission time.

### TimerManager

`server/room/timer_manager.go` manages per-player disconnect grace period timers:

- Each player can have at most one active timer
- Starting a new timer for a player cancels any existing one
- When a timer fires, it calls `RoomService.RemovePlayer()`
- Grace periods: 30 seconds for multiplayer, 30 minutes for solo mode

### PersistenceManager

`server/room/persistence_manager.go` handles room serialization:

- **Save**: Writes all rooms to `data/rooms.json` as JSON. Uses atomic writes (write to `.tmp`, then rename).
- **Load**: Reads `rooms.json` on startup. Handles missing files, empty files, and initializes empty maps for backward compatibility.
- **Cleanup**: `FindStaleRooms()` identifies rooms with no activity past `ROOM_MAX_AGE`.

Background workers (started from `server/room/service_persistence.go`):
- **Auto-save**: Runs every `AUTO_SAVE_INTERVAL` (default 30s)
- **Cleanup**: Runs every `CLEANUP_INTERVAL` (default 1h), deletes stale rooms

### RoomRepository

`server/room/repository.go` provides thread-safe room storage:

- **Two-level locking**: A repository-level `RWMutex` protects the room map. Each room has its own `Mutex` for state mutations.
- **Room IDs**: 4-character uppercase strings from a reduced charset (no 0, 1, I, O to avoid confusion).
- **Player IDs**: 16-hex string from `rand.Uint64()`.
- **Session tokens**: 64-hex string from 32 bytes of `crypto/rand`.
- **Case-insensitive lookup**: Room IDs normalized to uppercase.

`GetWithLock()` is the key method: acquires the repository read lock to find the room, gets a reference to the room's mutex, releases the repository lock, then acquires the room lock. This allows concurrent operations on different rooms.

## WebSocket Hub

`server/ws/hub.go` manages real-time connections:

- **Room-scoped client sets**: Each room has a `map[*Client]bool` of connected clients.
- **Client lifecycle**: On WebSocket upgrade, validates room ID and session token, reconnects the player in `RoomService`, creates a `Client` with a buffered send channel (256 messages), and starts read/write pump goroutines.
- **Broadcasting**: Sends JSON events to all clients in a room. If a client's send buffer is full, the client is unregistered (prevents slow clients from blocking).
- **Disconnection**: When the read pump detects a closed connection, unregisters the client and calls `RoomService.DisconnectPlayer()`.
- **Origin validation**: Checks configured allowed origins plus same-host policy.

Event types broadcast as JSON with a `type` field and event-specific data.

## Configuration

All configuration is via environment variables. `.env` provides checked-in defaults; `.env.local` (gitignored) provides personal overrides. Real environment variables always take priority.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server listen port |
| `DATA_DIR` | `data` | Base directory for all data files |
| `ALLOWED_ORIGINS` | `localhost` | Comma-separated hostnames for CORS |
| `ALLOW_SAME_HOST` | `true` | Allow requests from same hostname |
| `AUTO_SAVE_INTERVAL` | `30` | Seconds between room auto-saves |
| `CLEANUP_INTERVAL` | `3600` | Seconds between stale room cleanup |
| `ROOM_MAX_AGE` | `86400` | Seconds before a room is considered stale |
| `DISCONNECT_GRACE_PERIOD` | `30` | Seconds before disconnected player is removed (multiplayer) |
| `SOLO_DISCONNECT_GRACE_PERIOD` | `1800` | Seconds before disconnected player is removed (solo) |
| `SOLVER_TIMEOUT` | `30` | Seconds before solver is cancelled |
| `ENABLE_DAILY_CHALLENGE` | `false` | Enable daily challenge feature |

## Rate Limiting

`GetRoom` is rate-limited to 100 requests per minute per IP address, applied as middleware in `server/main.go`. This prevents abuse from polling clients. Other RPCs are not rate-limited since they require session tokens.

## Daily Challenge

### Manager (`server/daily/manager.go`)

Generates and serves daily puzzles:

- **Cache**: In-memory map of date to puzzles, with disk fallback.
- **Storage**: `data/daily_puzzles/YYYY/MM/DD.json`. Atomic writes.
- **Generation**: For each difficulty, tries up to 10,000 random games, solves with A\*, keeps the first matching the difficulty range. Date-seeded for reproducibility.
- **Background worker**: Runs immediately on startup, then daily at midnight UTC. Generates puzzles for the next 2 days to ensure all timezones have coverage.

Difficulty classification by optimal move count:

| Difficulty | Move Range |
|------------|-----------|
| Easy | 4-6 |
| Medium | 7-11 |
| Hard | 12+ |

### ProgressManager (`server/daily/progress.go`)

Tracks per-user completion:

- **Storage**: `data/users/XX/PLAYERID.json` (sharded by first 2 characters of player ID for filesystem scalability).
- **Cache**: In-memory map of player ID to progress.
- **Data model**: Map of date string to `DayProgress` (easy/medium/hard booleans).

## Testing Patterns

Server tests use Go's standard `testing` package:

- **Room service tests**: Create a `RoomService` with in-memory repository, exercise operations, assert state changes and signal emissions.
- **Manager tests**: Test each manager in isolation with constructed `Room` objects.
- **Physics tests**: `model/` package tests cover movement, wall collision, and solution validation. Some test cases are shared with the client via JSON files for cross-language verification.
- **Solver tests**: Test known puzzles against expected optimal solutions.
