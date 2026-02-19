# Client Deep Dive

The client is a Vue 3 single-page application built with Vite, using Pinia for state management and the Composition API throughout. It communicates with the server via Connect-Web (protobuf RPC) and WebSockets.

For the high-level architecture and how the client fits into the overall system, see [architecture.md](architecture.md).

## Entry Point and Setup

`client/web/src/main.ts` bootstraps the app: creates the Vue instance, registers Pinia and Vue Router, mounts to `#app`.

### Routing

`client/web/src/router/index.ts` defines five routes:

| Path | View | Notes |
|------|------|-------|
| `/` | `HomeView` | Landing page with create/join room |
| `/room/:roomId` | `RoomView` | Main game (roomId passed as prop) |
| `/help` | `HelpView` | Game instructions |
| `/daily` | `DailyChallengeView` | Daily puzzle selection |
| `/daily/:difficulty` | `DailyGameView` | Daily puzzle gameplay (difficulty as prop) |

A `beforeEach` navigation guard loads feature flags on first navigation and gates the `/daily` routes if `dailyChallengeEnabled` is false (redirects to home).

The router's base URL comes from `config.basePath`, which is set at runtime via `window.APP_CONFIG`.

### Runtime Configuration

`client/web/src/config.ts` reads from `window.APP_CONFIG`, which is populated by:
- `public/config.js` during local development
- `docker-entrypoint.sh` at container startup in production

Exported config:
- `httpBaseUrl`: API base URL (default: `http://localhost:8080`)
- `wsUrl`: WebSocket URL (derived from httpBaseUrl, switching http/https to ws/wss)
- `basePath`: Router base path (default: `/`)

## Views

### HomeView

Landing page with three main actions:
- **Solo mode**: Creates a single-player room and starts a game immediately
- **Create room**: Creates a multiplayer room, shows room ID for sharing
- **Join room**: Enter a room ID to join

Also shows: a "Return to Game" link if the player has a recent room (checked periodically), daily challenge link (if enabled), and persisted player name.

### RoomView

The main game container, managing the full session lifecycle:

1. **Not joined**: Shows join form (name input)
2. **Joined, no game**: Room lobby with start button (host only) and player list
3. **Game active**: `GameBoard` component with input handling
4. **Game ended**: Solution viewer with replay, leaderboard, and next game / leave options

Key composables wired here: `useRoomConnection` (server state sync), `useGameActions` (solution submission), `useReplay` (end-game playback).

The view detects pending players (joined mid-game, waiting for next round) and shows appropriate messaging.

### DailyChallengeView

Displays three puzzles (easy, medium, hard) with:
- Solved status and optimal move count (revealed after solving)
- Countdown timer to next daily reset (auto-refreshes when it hits zero)
- Progress bar showing solved/total

### DailyGameView

Single puzzle gameplay for a specific difficulty. Auto-submits when the puzzle is solved. Shows success modal with optimal move count on completion.

## Core Components

### GameBoard (`GameBoard.vue`)

The primary game UI. Renders the 16x16 grid with:
- **Cells**: Percentage-based sizing (6.25% per cell) for responsiveness
- **Walls**: Positioned on cell edges per wall semantics
- **Robots**: Colored circles on cells, selected robot highlighted
- **Target**: Marked cell showing which robot must reach it

Input handling:
- **Keyboard**: Via `useGameInput` composable (arrows/WASD for movement, 1-4 for robot selection, z/u for undo, etc.)
- **Touch/swipe**: Via `useSwipe` composable for mobile
- **Undo button**: Short tap undoes last move, press-and-hold (500ms) resets the board

Layout modes:
- **Desktop**: Horizontal layout with solution panels on the right
- **Mobile**: Vertical layout with bottom drawer for solutions

### PlayersPanel

Displays the player list during gameplay:
- Timer counting up from game start (capped at 30 minutes)
- Players sorted: solved first (by move count, then time), unsolved last
- Status badges: wins count, solution move count + time, finished-solving checkmark
- Host crown indicator
- Compact mode (dropdown) on mobile

### SolutionsDrawer

Mobile bottom drawer for managing solutions during gameplay:
- Collapsed: pill indicators showing move count and solved status per solution
- Expanded: full solution details with replay and delete actions
- Swipe gestures: up/down to toggle, left/right to switch solutions
- Add button for new solutions (up to 4 concurrent)

### SettingsModal

Room settings (host only):
- Toggle: show solver move count during game
- Toggle: show solver solutions in end-game viewer
- Boot player: dropdown to select, confirmation dialog before removal

## Composables

### useRoomConnection

The bridge between client state and server state. Manages the full connection lifecycle:

- **loadRoom()**: Fetches room state via `GetRoom` RPC. Detects stale localStorage (different game number), restores solver solutions from room state.
- **joinRoom()**: Calls `JoinRoom` RPC, stores credentials, connects WebSocket.
- **handleWebSocketEvent()**: Processes each event type (player_joined, game_started, player_solved, solver_complete, game_ended, etc.) by updating local state and calling appropriate callbacks.
- **Polling fallback**: Before joining, polls `GetRoom` periodically. After joining, relies on WebSocket events with visibility-change reload as a safety net.

### useGameInput

Maps keyboard input to game actions with mode awareness:

| Mode | Available Controls |
|------|-------------------|
| Normal (game active) | Movement (arrows/WASD), robot selection (1-4), undo (z/u/Escape), delete (Shift+D), new solution (n/+), switch solution (Shift+arrows), replay (p), help (?) |
| Game ended | Solution navigation (Shift+arrows), replay (p), help (?) |
| Modal open | Help toggle (?), close modal (Escape) |

Arrow key default behavior is prevented only during active input handling.

### useGameActions

Manages solution submission and server communication:

- **submitSolution()**: Only submits if the current solution is better than the player's previous best (fewer moves). Tracks best submission in localStorage.
- **markFinishedSolving()**: Signals the player is done attempting.
- **markReadyForNext()**: Signals readiness for the next game.
- **Best solution protection**: Before modifying a submitted solution (undo, delete), warns the user via callback.

### useReplay

End-game solution playback:

- Replays winning moves one at a time with 420ms delays between moves
- Can switch between player solutions and replay each
- Auto-starts replay when the game ends
- Tracks robot positions during replay to show intermediate states

### useSolutionDisplay

Computes display state for the end-game solution viewer:

- **Multiplayer order**: Top 3 solutions, then solver solutions, then current player's solution if not in top 3
- **Single player order**: Player solutions, then solver solutions
- **Time formatting**: Deduplicates solutions by player ID before formatting (handles solution updates)

## State Management

### roomStore (Pinia)

Player identity and session persistence. All state persisted to localStorage via individual watchers.

| State | Purpose |
|-------|---------|
| `currentPlayerName` | Player's chosen display name |
| `currentPlayerId` | Player's UUID |
| `currentSessionToken` | Auth token for RPCs |
| `lastRoomId` | Most recent room for "Return to Game" |
| `isSinglePlayer` | Solo mode flag |
| `puzzlesSolved` / `puzzlesAttempted` | Daily challenge counters |
| `showSolverMoveCount` / `showSolverSolutions` | Persisted room settings |

### gameStore (Pinia)

Active puzzle state and physics. The most complex store.

**Core state**: `robots`, `initialRobots`, `vWalls`, `hWalls`, `target`, `solutions` (array of up to 4), `activeSolutionIndex`, `selectedRobotId`, `committedMoves` (for animation sync).

**Key behaviors**:

- **moveRobot(direction)**: Core puzzle logic. Calls `calculateDestination()` from gamePhysics. Includes **input forgiveness**: if the player moves the same robot in the opposite direction as their last move, it undoes that move instead of adding a new one. Marks the solution as solved when the target robot reaches the target position.
- **Multi-solution editing**: Players work on up to 4 solutions. `startNewSolution()` auto-deletes the worst unsubmitted solution if at capacity.
- **Solution persistence**: Solutions save to localStorage keyed by room ID. On game load, restores if the game number matches, otherwise clears.
- **Animation tracking**: `committedMoves` are populated with a delay to synchronize with CSS transitions. Timeouts are tracked and cancelled on undo/forgiveness to prevent visual glitches.

### dailyStore (Pinia)

Daily challenge state:
- Generates or restores a stable device-based player ID
- `fetchDaily()`: Fetches puzzles from server with timezone offset
- `submitSolution()`: Submits solution, returns whether correct and whether it is a new completion

### featureStore (Pinia)

Feature flag management:
- `fetchFeatures()`: Calls `GetServerInfo` RPC, sets `dailyChallengeEnabled`
- Loaded once on first navigation via router guard

## Services

### Connect-Web Client (`connectClient.ts`)

Singleton RPC client generated from the proto definition. Created with `createConnectTransport` pointing at `config.httpBaseUrl`.

### WebSocket Service (`websocket.ts`)

Real-time event connection with reconnection logic:

- **Connection**: Opens WebSocket to `/ws?roomId=X&token=Y`
- **Events**: Parses JSON messages with `type` field, dispatches to callback
- **Reconnection strategy**:
  - If never connected successfully: max 5 attempts (room may not exist)
  - If previously connected: unlimited retries (likely a network issue)
  - 3-second delay between attempts
  - Notifies consumer after max attempts exceeded
- **Disconnect**: Clean shutdown sets `shouldReconnect = false`

### AnimationService (`animationService.ts`)

Timing constants synchronized with CSS transitions:

| Constant | Value | Used For |
|----------|-------|----------|
| `MOVE_DELAY` | 150ms | Delay between moves during in-game solution switching |
| `MOVE_ANIMATION` | 150ms | Robot movement CSS transition duration |
| `REPLAY_DELAY` | 420ms | Delay between moves during end-game replay |
| `REPLAY_ANIMATION` | 400ms | Robot movement CSS transition during replay |

Helper functions: `animateSequence()`, `animateSequenceReverse()`, `scheduleAfter()`, `chainAnimations()`.

### gamePhysics (`gamePhysics.ts`)

Client-side mirror of server physics for optimistic UI:

- **hasWall(x, y, direction, vWalls, hWalls)**: Checks wall and board edge blocking. Board edges at x=0 (left), x=15 (right), y=0 (up), y=15 (down). Internal walls checked against vWalls (left/right) and hWalls (up/down).
- **isOccupied(x, y, excludeRobotId, robots)**: Checks if another robot occupies a cell.
- **calculateDestination(robot, direction, robots, vWalls, hWalls)**: Slides robot one cell at a time until blocked. Returns final position.

Uses `BOARD_SIZE = 16` from constants.

## Testing Patterns

Client tests use Vitest with jsdom environment:

- **Physics tests**: Verify `calculateDestination()` against known board configurations. Some test cases are shared with the server via JSON for cross-language verification.
- **Store tests**: Test Pinia stores in isolation with mock data.
- **Composable tests**: Test composable logic by providing mock dependencies.
