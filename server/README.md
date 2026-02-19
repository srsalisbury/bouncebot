# BounceBot Server

The Go backend for BounceBot. Manages multiplayer game state, real-time communication via WebSockets, and background puzzle solving.

For the full server documentation, see **[docs/server.md](../docs/server.md)**.

For architecture context, see **[docs/architecture.md](../docs/architecture.md)**.

## Quick Start

```bash
# From the project root (not server/)
go run ./server
```

## Tests

```bash
go test ./...
```
