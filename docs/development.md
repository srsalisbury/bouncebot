# Development and Deployment Guide

Practical instructions for building, running, testing, and deploying BounceBot.

For architecture context, see [architecture.md](architecture.md). For server and client internals, see [server.md](server.md) and [client.md](client.md).

## Prerequisites

- **Go 1.24+**
- **Node.js 18+** with npm
- **protoc** (Protocol Buffers compiler) with plugins:
  - `protoc-gen-go`
  - `protoc-gen-go-grpc`
  - `protoc-gen-connect-go`
  - `buf` CLI (for TypeScript generation)
- **Docker** and **Docker Compose** (for containerized deployment)

## Local Development

### Starting the Server

Run from the **project root** (not from `server/`):

```bash
go run ./server
```

This ensures `rooms.json` and `data/` resolve relative to the project root. The server listens on port 8080 by default.

Configuration comes from `.env` (checked in) and `.env.local` (gitignored, for personal overrides). Real environment variables take priority over both files.

### Starting the Client

```bash
cd client/web
npm install
npm run dev
```

The Vite dev server starts on port 5173. It connects to the Go server at `http://localhost:8080` by default (configured in `public/config.js`).

### Environment Files

| File | Purpose | Committed |
|------|---------|-----------|
| `.env` | Shared local development defaults | Yes |
| `.env.local` | Personal overrides | No (gitignored) |

The `.env` file enables daily challenges by default (`ENABLE_DAILY_CHALLENGE=true`). All other values use code defaults (commented out in `.env` for reference).

## Proto Compilation

When modifying `proto/bouncebot.proto`, regenerate both Go and TypeScript code:

```bash
cd proto
./compile_protos.sh
```

This runs:
1. `protoc` with Go, gRPC, and Connect plugins (generates `.pb.go` and Connect files in `proto/`)
2. `npx buf generate` in `client/web/` (generates TypeScript files in `client/web/src/gen/`)

After regeneration, both the server and client need to be restarted.

## Testing

### Server (Go)

```bash
go test ./...
```

Tests cover:
- Room service operations and signal emissions
- Individual manager behavior (player, game lifecycle, solution, persistence)
- Game physics (movement, wall collision, solution validation)
- Solver correctness against known puzzles
- Board construction and panel rotation

### Client (Vitest)

```bash
cd client/web
npm test
```

Uses Vitest with jsdom environment. Tests cover:
- Physics calculations (wall blocking, robot sliding)
- Store logic (game state, solution management)
- Composable behavior

### Cross-Language Physics Tests

Some physics test cases are shared between Go and TypeScript via JSON files to ensure the client's optimistic movement calculations match the server's authoritative logic. Both test suites read the same test data and verify identical results.

## Docker Deployment

### Build and Run

```bash
docker compose up --build
```

This starts:
- **Server** on port 8080 (Go binary, persists data to a named volume)
- **Client** on port 80 (Nginx serving built Vue app)

### Environment Variables

#### Server Container

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server listen port |
| `DATA_DIR` | `data` | Base directory for data files (should match volume mount) |
| `ALLOWED_ORIGINS` | `localhost` | Comma-separated hostnames for CORS |
| `ALLOW_SAME_HOST` | `true` | Allow requests from same hostname |
| `AUTO_SAVE_INTERVAL` | `30` | Seconds between room auto-saves |
| `CLEANUP_INTERVAL` | `3600` | Seconds between stale room cleanup |
| `ROOM_MAX_AGE` | `86400` | Seconds before a room is considered stale |
| `DISCONNECT_GRACE_PERIOD` | `30` | Multiplayer disconnect grace period (seconds) |
| `SOLO_DISCONNECT_GRACE_PERIOD` | `1800` | Solo disconnect grace period (seconds) |
| `SOLVER_TIMEOUT` | `30` | Solver timeout (seconds) |
| `ENABLE_DAILY_CHALLENGE` | `false` | Enable daily challenge feature |

#### Client Container

| Variable | Default | Description |
|----------|---------|-------------|
| `BASE_PATH` | `/` | URL subpath for serving the app (e.g., `/app/`) |

`BASE_PATH` is applied at container startup by `docker-entrypoint.sh`, which rewrites asset paths in HTML, JS, CSS, and manifest.json. See [BASE_PATH_SETUP.md](BASE_PATH_SETUP.md) for the full guide.

### Data Volume

The `docker-compose.yml` creates a named volume `bouncebot-data` mounted at `/data` in the server container. This persists:
- `rooms.json` (active room state)
- `daily_puzzles/` (generated puzzles)
- `users/` (daily challenge progress)

## Data Directory Structure

```
data/
├── rooms.json                          # All active room state (auto-saved)
├── daily_puzzles/
│   └── YYYY/
│       └── MM/
│           └── DD.json                 # Three puzzles (easy, medium, hard)
└── users/
    └── XX/                             # Sharded by first 2 chars of player ID
        └── PLAYERID.json              # Per-user daily challenge progress
```

All writes use atomic rename (write to `.tmp`, then `os.Rename`) to prevent corruption from crashes.

## Feature Flags

Feature flags flow through a specific pipeline:

1. **Environment variable** (`ENABLE_DAILY_CHALLENGE`) loaded by `config.LoadFromEnv()`
2. **Server config** stores the boolean
3. **`GetServerInfo` RPC** exposes enabled features to the client
4. **`featureStore`** in the client calls `GetServerInfo` on first navigation
5. **Router guard** checks `featureStore.dailyChallengeEnabled` before allowing access to `/daily` routes

This allows enabling/disabling features per deployment without rebuilding the client.

## Graceful Shutdown

On SIGINT or SIGTERM, the server:
1. Stops background workers (auto-save, cleanup, daily generation)
2. Stops all disconnect timers
3. Performs a final room save to disk
4. Exits
