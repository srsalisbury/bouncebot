# BounceBot

BounceBot is a real-time, multiplayer web-based implementation of the board game "Ricochet Robots". Players compete to find the shortest sequence of moves to guide a colored robot to its target location on a grid-based board with walls.

## Architecture

The project follows a modern client-server architecture with a Go backend and a Vue.js single-page application frontend.

```
+------------------------+      +-------------------------+
|                        |      |                         |
|   Vue.js Frontend      |      |      Go Backend         |
| (Vite, Pinia, TS)      |      |                         |
|                        |      |                         |
+------------------------+      +-------------------------+
           |                             ^
           |         (gRPC / Connect)    | (WebSockets)
           +--------- RPC Actions ------->
           |                             |
           <------ Real-time State ------+
                     Updates
```

### Backend (Go)

The Go server is the authoritative source for all game logic and state.

-   **Responsibilities**:
    -   Managing game sessions (creating, joining).
    -   Enforcing game rules (validating moves, checking solutions).
    -   Persisting session state.
    -   Broadcasting state changes to clients.
-   **Key Technologies**:
    -   **Connect**: Used to generate a type-safe gRPC-style API for client-initiated actions (e.g., `SubmitSolution`).
    -   **WebSockets**: Used for real-time, server-to-client communication. When one player's action changes the game state, the server broadcasts the new state to all clients in the session.

### Frontend (Vue.js)

The frontend is a single-page application responsible for rendering the game state and capturing user input.

-   **Responsibilities**:
    -   Displaying the game board, robots, and targets.
    -   Accepting user input (primarily keyboard-based) to move the robots.
    -   Communicating with the backend via the Connect RPC and WebSocket services.
    -   Managing client-side state for responsive UI updates.
-   **Key Technologies**:
    -   **Vue 3**: Core frontend framework (using the Composition API).
    -   **Vite**: Build tooling.
    -   **Pinia**: Centralized state management.
    -   **TypeScript**: For type safety.

### Communication (Connect + WebSockets)

BounceBot uses a dual-channel communication model to provide a responsive user experience:

1.  **RPC Actions (Client-to-Server)**: The client uses a type-safe RPC client generated from the `.proto` definition to send specific commands to the server, such as creating a session or submitting a move sequence. This is a standard request/response model.
2.  **Real-time State Updates (Server-to-Client)**: The server uses a WebSocket to push updates to all clients in a session whenever the state changes. This ensures that all players see the same game state in real-time without needing to poll the server.

## Core Game Logic

The game is played on a 16x16 grid with walls. The objective is to move a designated robot to a target square.

-   **Movement**: Robots move in one of the four cardinal directions (Up, Down, Left, Right).
-   **Ricochet Mechanic**: Once a robot starts moving, it continues in a straight line until it hits an obstacle (another robot, a wall, or the edge of the board). It does not stop on an empty square.
-   **Solving**: Players find a sequence of moves (e.g., "Red Up, Blue Left, Red Right") and submit it. The server validates if the solution is correct and if it's the shortest one found so far.

## How to Run

### Option 1: Docker (Recommended)

**Pull and run from GitHub Container Registry:**
```sh
docker pull ghcr.io/srsalisbury/bouncebot-server:latest
docker pull ghcr.io/srsalisbury/bouncebot-client:latest

docker run -d -p 8080:8080 ghcr.io/srsalisbury/bouncebot-server:latest
docker run -d -p 80:80 ghcr.io/srsalisbury/bouncebot-client:latest
```

**Or use docker-compose:**
```sh
docker compose up
```
- Client: http://localhost (port 80)
- Server: http://localhost:8080

### Option 2: Local Development

#### Prerequisites
- Go 1.24+
- Node.js & npm
- `protoc-gen-go` and `protoc-gen-connect-go` (for compiling protos)

#### 1. Run the Backend Server

```sh
go run ./server
```
The server will start on `localhost:8080`.

```sh
# Or with custom port
go run ./server -port 9000

# Or build and run
go build -o bouncebot-server ./server
./bouncebot-server
```

#### 2. Run the Frontend Client

```sh
cd client/vue1
npm install
npm run dev
```
The frontend development server will start on `localhost:5173`. Open this URL in your browser to play.

#### 3. Compile Protobuf Definitions

If you make changes to `proto/bouncebot.proto`, regenerate the Go and TypeScript code:

**Go (from repo root):**
```sh
./proto/compile_protos.sh
```

**TypeScript (from client/vue1):**
```sh
npm run generate
```

## Documentation

- **[client/vue1/CONTEXT.md](./client/vue1/CONTEXT.md)** - Project overview and key decisions
- **[client/vue1/IMPLEMENTATION_PLAN.md](./client/vue1/IMPLEMENTATION_PLAN.md)** - Step-by-step build plan
- **[client/vue1/PROGRESS.md](./client/vue1/PROGRESS.md)** - Completed steps and PR history
