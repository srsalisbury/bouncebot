# BounceBot Vue3 Client

Web client for the BounceBot puzzle game (Ricochet Robots style).

## Quick Start

**1. Start the Go server** (from repo root):
```bash
go run ./server
```
Server runs at http://localhost:8080

**2. Start the Vue client** (from this directory):
```bash
npm install
npm run dev
```
Client runs at http://localhost:5173

## Server

The Go backend provides puzzle generation and solution validation via Connect RPC.

**Run server:**
```bash
# From repo root
go run ./server

# Or with custom port
go run ./server -port 9000
```

**Build server:**
```bash
go build -o bouncebot-server ./server
./bouncebot-server
```

## Proto Generation

**Go (server):** Run from `proto/` directory:
```bash
./compile_protos.sh
```
Requires: `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-connect-go`

**TypeScript (client):** Run from this directory:
```bash
npm run generate
```
Generates types to `src/gen/bouncebot_pb.ts`

## Documentation

- **[CONTEXT.md](./CONTEXT.md)** - Project overview, architecture, and key decisions
- **[IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md)** - Step-by-step build plan
- **[PROGRESS.md](./PROGRESS.md)** - Completed steps and PR history

## Docker Deployment

**Pull from GitHub Container Registry:**
```bash
docker pull ghcr.io/srsalisbury/bouncebot-server:latest
docker pull ghcr.io/srsalisbury/bouncebot-client:latest
```

**Run with docker-compose** (from repo root):
```bash
docker compose up
```
- Client: http://localhost (port 80)
- Server: http://localhost:8080

**Build locally:**
```bash
docker compose up --build
```

## Tech Stack

- Vue 3 (Composition API)
- TypeScript
- Vite
- Connect RPC (browser-friendly gRPC)
