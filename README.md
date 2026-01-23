# BounceBot

BounceBot is a real-time, multiplayer web implementation of the board game **Ricochet Robots**. 

Players compete in a shared room to find the shortest sequence of moves to guide a designated robot to its target. The catch? Robots only stop when they hit a wall or another robot.

While players calculate their moves, a server-side **A* solver** runs in the background to establish a "gold standard" for the puzzle's difficulty and provide hints.

![BounceBot game board mid-game](docs/screenshot.png)

## Core Architecture

BounceBot utilizes an **authoritative server model** to ensure game state integrity and synchronization across all players.

*   **Go Backend (Authoritative)**: Manages room state, validates all moves, and orchestrates the game lifecycle (lobby -> playing -> results). It uses a signal-based component architecture to decouple player management from game logic.
*   **Vue 3 Frontend (Optimistic)**: A responsive SPA that uses the same physics engine as the server to provide **optimistic UI updates**. Moves are calculated locally for instant feedback and then synchronized with the server.
*   **Dual-Channel Communication**:
    *   **Connect-RPC**: Handles transactional actions (creating rooms, submitting solutions) over HTTP/2.
    *   **WebSockets**: Provides a real-time stream of state changes (player joins, solver completions, game starts) to all connected clients.
*   **Background Solvers**: The backend registers automated solvers (like A*) that trigger immediately when a new board is generated, allowing the system to know the optimal solution count within seconds.

## Prerequisites

To develop or run BounceBot locally, you need the following tools:

### Backend & Tooling
*   **Go (v1.24+)**: The core language for the backend server and solver logic.
*   **Protocol Buffers (`protoc`)**: The compiler used to generate type-safe code from `.proto` definitions.
*   **Go Protobuf Plugins**: Required for the server to handle gRPC and Connect-RPC traffic:
    ```bash
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
    ```

### Frontend
*   **Node.js (v18+) & npm**: Required to build the Vue application and manage dependencies.
*   **Vite & Vue 3**: (Managed via `npm`) The frontend uses Vite for extremely fast HMR (Hot Module Replacement) and Vue 3 for component logic.

## Getting Started

### 1. Run the Backend Server
The server manages the game state and persistence.

```bash
# From the root directory
go run ./server
```
*The server defaults to port `8080`. Room state is persisted to `bouncebot_rooms.json`.*

### 2. Run the Frontend Client
The client provides the interactive game board and lobby.

```bash
cd client/vue1
npm install
npm run dev
```
*The client will typically start at `http://localhost:5173`. Open this in your browser to play.*

### 3. Development: Modifying the API
If you modify the API definitions in `proto/bouncebot.proto`, you must regenerate both the Go and TypeScript codebases. We provide a centralized script for this:

```bash
./proto/compile_protos.sh
```

## Deployment

### Docker (Recommended)

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

## Documentation

*   **[server/README.md](./server/README.md)** - Backend code layout, RPC endpoints, WebSocket events.
*   **[client/vue1/README.md](./client/vue1/README.md)** - Frontend architecture, Pinia stores, and composables.
*   **`model/`**: Pure game logic (physics, board generation) shared between the server and the solvers.
*   **`solver/`**: Implementation of the A* and BFS algorithms used to solve puzzles automatically.

## Appendix: Scaling to Multiple Servers

The current JSON file persistence (`bouncebot_rooms.json`) works well for single-server deployments. For multi-server deployments (e.g., Kubernetes with multiple replicas), you'll need a shared room store like Redis:

1. **Add Redis dependency**: `go get github.com/redis/go-redis/v9`
2. **Implement a Redis-backed RoomRepository**: Replace the file-based `Load`/`Save` methods with Redis operations.
3. **Use Redis pub/sub**: Replace the in-memory WebSocket hub with Redis pub/sub for cross-server broadcasting.
4. **Room affinity**: Alternatively, use sticky sessions to route players to the same server instance.