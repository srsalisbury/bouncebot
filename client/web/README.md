# BounceBot Client Documentation

The BounceBot client is a Single Page Application (SPA) built with **Vue 3**, **TypeScript**, and **Vite**. It interacts with the Go server using **Connect-Web** (gRPC-compatible) and **WebSockets**.

## Getting Started

### Prerequisites
- Node.js 18+
- npm

### Running Locally

```bash
cd client/web
npm install
npm run dev
```

The app will start at `http://localhost:5173`.

## Project Structure

The project follows a standard Vue 3 structure, organized by feature type rather than file type where possible.

```
client/web/src/
├── main.ts               # App entry point, mounts the Vue app
├── App.vue               # Root component, handles layout & global overlays
├── router.ts             # Route definitions (/, /room/:roomId)
├── config.ts             # Runtime configuration
├── constants.ts          # Shared constants (colors, board size, timings)
├── gamePhysics.ts        # Core movement logic (mirrors server)
├── stores/               # Pinia state management
├── services/             # External communication (RPC & WebSockets)
├── composables/          # Reusable logic (Vue Composition API)
├── components/           # UI Components
├── views/                # Route-level Page Components
└── gen/                  # Generated Protobuf code (do not edit)
```

## Application Architecture

### Views (Pages)

The application has two primary views:

1.  **`HomeView.vue`**: The landing page.
    -   **Responsibility**: Allows users to create new rooms or join existing ones. Handles "Solo Play" (creating a single-player room) and "Play with Friends".
    -   **Key Logic**: Form validation for player names and room IDs.

2.  **`RoomView.vue`**: The main game container.
    -   **Responsibility**: Manages the entire lifecycle of a game session. It acts as a controller, switching between states: "Loading", "Lobby/Joining", "Playing", and "Waiting Room" (between games).
    -   **Key Logic**: Orchestrates `useRoomConnection` to sync with the server. It handles global UI elements like the "Leaderboard" and "Settings" modals.

### Components

*   **`GameBoard.vue`**: The core gameplay component.
    -   **Responsibility**: Renders the 16x16 grid, walls, targets, and robots.
    -   **Key Logic**: It integrates heavily with composables to handle user input (`useGameInput` for keyboard, `useSwipe` for touch). It manages the visual state of the board, including animations and solution replays (`useReplay`).
*   **`PlayersPanel.vue`**: Displays the list of players, their connectivity status, and scores.
*   **`LeaderboardModal.vue`**: Shows cumulative scores for the room.
*   **`SettingsModal.vue`**: Allows the room host to configure settings (e.g., showing solver hints).

### Composables (Logic)

Logic is extracted into composables to keep components clean and testable.

*   **`useRoomConnection.ts`**: The bridge between the View and the Network.
    -   Manages loading room state via RPC.
    -   Establishes and manages the WebSocket connection.
    -   Handles real-time events (`player_joined`, `game_started`) and updates the local state.
*   **`useGameInput.ts`**: Centralized input handler.
    -   Maps keyboard events (Arrows, WASD, Numbers) to game actions.
    -   Handles global shortcuts (Undo, Reset, New Solution).
*   **`useReplay.ts`**: Manages the playback of solutions.
    -   Used by `GameBoard` to animate a sequence of moves (e.g., watching a bot's solution).
*   **`useGameActions.ts`**: Encapsulates specific transactional actions like submitting a solution or marking a player as ready.

### State Management (Pinia)

The application uses two primary stores:

1.  **`roomStore`**: Manages the **Session Context**.
    -   **State**: Current Room ID, Player ID, Player Name, Room Settings.
    -   **Persistence**: Uses `localStorage` to remember the user across reloads.
    -   **Role**: "Who am I and where am I?"

2.  **`gameStore`**: Manages the **Active Puzzle**.
    -   **State**: Robot positions, Wall locations, Target, Move History (Undo stack).
    -   **Physics**: Directly integrates with `gamePhysics.ts` to calculate move results.
    -   **Role**: "What is the state of the board right now?"

### Services (Networking)

*   **`connectClient.ts`**: Configures the **Connect-Web** client for making RPC calls to the server.
*   **`websocket.ts`**: A singleton service that manages the raw WebSocket connection, handles reconnection logic, and dispatches typed events to listeners.

### Physics & Game Logic

To ensure instant feedback, the client implements the exact same physics engine as the server.

*   **`gamePhysics.ts`**: Contains the rules for robot movement.
    -   `calculateDestination`: Determines where a robot stops sliding based on walls and other robots.
    -   Used by `gameStore` to validate moves immediately on the client side, ensuring the UI is responsive (optimistic UI) before confirming with the server.

## Development Workflow

### Protocol Buffers
If the `proto/bouncebot.proto` file changes, you must regenerate the code for both the client and server using the script in the `proto/` directory:

```bash
# From the project root
./proto/compile_protos.sh
```

This script generates the Go handlers for the server and uses `buf` to generate the TypeScript stubs in `src/gen/`.

### Testing
Unit tests use **Vitest**.

```bash
npm test
```

Tests cover critical logic like `gamePhysics`, `gameStore`, and input composables.
