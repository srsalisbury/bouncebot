# System Architecture

BounceBot is a web-based implementation of [Ricochet Robots](https://en.wikipedia.org/wiki/Ricochet_Robots), a puzzle game where players slide robots on a grid to reach a target. It supports both multiplayer rooms (real-time competitive solving) and single-player daily challenges. A background A\* solver provides optimal solutions for each puzzle.

## High-Level Architecture

```
┌──────────────────┐          ┌──────────────────────────────────┐
│   Vue 3 Client   │◄── WS ──►│          Go Server               │
│  (Vite + Pinia)  │◄─ RPC ──►│  (Connect RPC + WebSocket Hub)   │
└──────────────────┘          │                                  │
                              │  ┌─────────┐  ┌──────────────┐  │
                              │  │  model/  │  │   solver/    │  │
                              │  │ (shared  │  │ (A*, BFS,    │  │
                              │  │  physics)│  │  registry)   │  │
                              │  └─────────┘  └──────────────┘  │
                              └──────────────────────────────────┘
```

Key packages:

- **`server/`**: HTTP server, RPC handlers, room orchestration, WebSocket hub, daily challenge system
- **`client/web/`**: Vue 3 SPA with Pinia stores, composables for game logic, Connect-Web RPC client
- **`model/`**: Pure Go game logic (board construction, robot physics, solution validation). No server dependencies.
- **`solver/`**: Pluggable solver system (A\* with heuristic, BFS, registry, manager). Depends on `model/` only.
- **`proto/`**: Protobuf service definition and generated Go code

## Dual Communication Channels

The client and server communicate over two channels, each serving a distinct purpose:

**Connect RPC (request/response transactions)**
Used for operations that need acknowledgement: creating/joining rooms, starting games, submitting solutions, marking ready. The client sends a request, the server validates and mutates state, and returns the result. All RPCs are defined in `proto/bouncebot.proto`.

**WebSocket (real-time event broadcast)**
Used for pushing state changes to all players in a room: a player joined, someone solved the puzzle, the game ended. The client opens a WebSocket connection after joining a room and receives JSON events. The server's WebSocket hub maintains a set of connected clients per room and broadcasts events as they occur.

This split keeps the RPC handlers simple (validate, mutate, respond) while the WebSocket layer handles fan-out without coupling to game logic.

### Proto Service Summary

| RPC | Purpose |
|-----|---------|
| `CreateRoom` | Create a new room, returns room state + player ID + session token |
| `JoinRoom` | Add player to existing room |
| `GetRoom` | Fetch current room state (rate-limited: 100 req/min per IP) |
| `StartGame` | Begin a new game in the room |
| `SubmitSolution` | Validate and record a player's solution |
| `MarkFinishedSolving` | Player signals they are done looking for solutions |
| `MarkReadyForNext` | Player is ready for the next game |
| `UpdateRoomSettings` | Host changes room settings |
| `BootPlayer` | Host removes another player |
| `LeaveRoom` | Player leaves the room |
| `GetDailyChallenge` | Fetch daily puzzles for the user's local date |
| `SubmitDailySolution` | Validate and record a daily puzzle solution |
| `GetServerInfo` | Returns server capabilities (feature flags) |

### WebSocket Events

| Event | Trigger |
|-------|---------|
| `player_joined` | Player enters room |
| `player_left` | Player removed from room |
| `game_started` | New game begins |
| `player_solved` | Player submits a solution (includes move count) |
| `player_finished_solving` | Player done attempting the puzzle |
| `player_ready_for_next` | Player ready for the next game |
| `game_ended` | Game concludes (includes winner and winning moves) |
| `solver_complete` | Background solver finished (includes solver name and moves) |
| `settings_changed` | Room settings updated by host |

## Game Model

### Board Construction

Boards are built from four 8x8 **panels** assembled into a 16x16 grid, matching the physical Ricochet Robots game. The codebase includes 12 predefined panels (defined as ASCII art in `model/board_builder.go`). A random board selects four panels and rotates them into the four quadrants.

**Wall semantics**: Walls exist on cell edges. A vertical wall at position (X, Y) sits to the right of that cell. A horizontal wall at position (X, Y) sits below that cell. Board edges have implicit walls.

### Robot Physics

Robots **slide** in a cardinal direction until they hit a wall, a board edge, or another robot. They cannot stop mid-slide. This is the core puzzle mechanic.

The physics logic is implemented twice:

- **Server** (`model/game_movement.go`): Authoritative. `ComputeDestination()` calculates where a robot ends up. `CheckSolution()` replays an entire move sequence and validates the target is reached.
- **Client** (`client/web/src/services/gamePhysics.ts`): Mirror implementation for optimistic UI. `calculateDestination()` uses the same wall-checking logic so the client can show moves instantly without waiting for the server.

Both implementations use identical wall-checking logic: for each step in the slide direction, check whether a wall or board edge blocks movement, then check whether another robot occupies the next cell.

### Game Flow

A **Game** consists of a board, four robots at specific positions, and a target (which robot must reach which cell). When a game continues from a previous round, the board stays the same but robots start from wherever the winning solution left them, and a new random target is chosen.

## Server Architecture

See [server.md](server.md) for the full deep dive.

The server follows a **component-based orchestrator** pattern. `RoomService` is the central facade that coordinates six specialized managers:

- **PlayerManager**: Player join/disconnect/reconnect/remove lifecycle
- **GameLifecycle**: State machine (no game, playing, finished, waiting for next)
- **SolutionManager**: Solution validation, upsert (only improvements accepted), winner determination
- **TimerManager**: Per-player disconnect grace period timers
- **PersistenceManager**: JSON file auto-save and room cleanup
- **RoomRepository**: Thread-safe room CRUD with per-room locking

### Signal System

Components do not call each other directly. Instead, each operation returns a list of **signals** that the orchestrator processes after releasing the room lock. This prevents deadlocks and keeps components decoupled.

Five signal types:

| Signal | Effect |
|--------|--------|
| `BroadcastSignal` | Send an event to all WebSocket clients in the room |
| `EndGameSignal` | Trigger game end (determine winner, broadcast result) |
| `StartNextGameSignal` | Start the next game (promote pending players, generate puzzle) |
| `StartTimerSignal` | Start a disconnect grace period timer for a player |
| `CancelTimerSignal` | Cancel a player's disconnect timer (they reconnected) |

The pattern: acquire room lock, execute operation, release lock, then process signals. Signal processing may itself acquire locks and produce more signals (recursive processing).

### Concurrency Model

- **Repository-level RWMutex**: Protects the room map (adding/removing rooms)
- **Per-room Mutex**: Each room has its own lock for state mutations. Operations on different rooms run in parallel.
- **Lock ordering**: Room lock is always released before processing signals. This prevents deadlocks when signal processing needs to acquire locks.
- **Background goroutines**: Auto-save (periodic), cleanup (periodic), solver jobs (per-game), daily puzzle generation (daily at midnight UTC)

## Solver System

Solvers live in `solver/` and follow a plugin pattern:

- **Registry** (`solver/registry.go`): Thread-safe map of solver name to `Solver` interface. Solvers auto-register via `init()` functions. The server imports solver packages via blank imports to trigger registration.
- **Manager** (`solver/manager.go`): Launches solver jobs in background goroutines. Each job runs with a context timeout (default 30s). On completion, calls a callback that stores the result on the room and broadcasts to clients. All registered solvers run for each game.
- **A\* solver** (`solver/astar/`): The primary solver. Uses a priority queue with an admissible heuristic based on reverse BFS from the target with rook movement. The heuristic accounts for actual board physics (walls, sliding) rather than simple Manhattan distance. Guaranteed to find optimal solutions.
- **BFS solver** (`solver/bfs/`): Simpler exhaustive search. Has an `init()` function but is not imported by the server (only used in benchmarks). A\* is preferred for production.
- **Solution reordering** (`solver/reorder.go`): After finding a solution, reorders moves to minimize "robot switches" (e.g., move robot A twice, then robot B, rather than alternating). This makes solutions easier for humans to follow.

Solver integration flow: game starts, `RoomService` calls `onGameStart` callback, `SolverManager.StartJob()` launches a goroutine, solver runs, result stored on room, `solver_complete` event broadcast.

## Daily Challenge System

When enabled (`ENABLE_DAILY_CHALLENGE=true`), the server generates three puzzles per day (easy, medium, hard) classified by optimal move count:

| Difficulty | Optimal Moves |
|------------|---------------|
| Easy | 4-6 |
| Medium | 7-11 |
| Hard | 12+ |

**Generation**: A background worker runs at midnight UTC. For each difficulty, it tries up to 10,000 random puzzles, solves each with A\*, and keeps the first one that falls in the target range. Puzzles are seeded by date for reproducibility.

**Storage**: Puzzles are stored as JSON at `data/daily_puzzles/YYYY/MM/DD.json`. User progress is stored at `data/users/XX/PLAYERID.json` (sharded by first two characters of the player ID).

**Client integration**: The client generates a stable device-based player ID (stored in localStorage) and sends it with timezone offset. The server determines the user's local date and returns that day's puzzles along with their completion status.

## Security Model

- **Session tokens**: Generated on CreateRoom/JoinRoom (32 bytes from crypto/rand, hex-encoded to 64 characters). Required for all state-mutating RPCs and WebSocket connections. Validated per-request.
- **Host permissions**: The first player in a room is the host. Only the host can change settings and boot players. Starting a game does not require host privileges.
- **Rate limiting**: `GetRoom` is limited to 100 requests per minute per IP address.
- **CORS**: Configured via `ALLOWED_ORIGINS` env var. The `ALLOW_SAME_HOST` option (default true) automatically allows requests from the same hostname, which simplifies deployments where the client and server share a host.
- **WebSocket authentication**: Session token is validated on WebSocket upgrade. Invalid tokens are rejected before the connection is established.

## Client Architecture

See [client.md](client.md) for the full deep dive.

The client is a Vue 3 SPA using Pinia for state management and the Composition API throughout.

Key layers:

- **Views**: HomeView (lobby), RoomView (main game), DailyChallengeView/DailyGameView (daily puzzles), HelpView
- **Composables**: `useRoomConnection` (server sync bridge), `useGameInput` (keyboard/touch), `useGameActions` (solution submission), `useReplay` (end-game solution playback), `useSolutionDisplay` (leaderboard formatting)
- **Stores**: `roomStore` (player identity, persisted to localStorage), `gameStore` (puzzle state, robot positions, solutions), `dailyStore` (daily challenge state), `featureStore` (feature flags from server)
- **Services**: Connect-Web RPC client, WebSocket service (reconnection logic), AnimationService (timing constants), gamePhysics (movement calculation)

The client supports **multi-solution editing**: players can work on up to 4 solutions simultaneously, switching between them. Solutions persist to localStorage per game so they survive page refreshes. An **input forgiveness** system detects when a player moves the same robot in the opposite direction and automatically undoes the previous move instead.

## Planning Documents

The `planning/` directory contains design documents. Some describe implemented features, others describe future work:

| Document | Status |
|----------|--------|
| `DAILY_PUZZLE_DESIGN.md` | Implemented |
| `PLAYER_AUTHENTICATION_DESIGN.md` | Not implemented |
| `USER_LOGIN_DESIGN.md` | Not implemented |
| `FUTURE_FEATURES.md` | Roadmap |
| `enhanced-astar-heuristic.md` | Not implemented |
| `smart-solution-time-formatting.md` | Partially implemented |
