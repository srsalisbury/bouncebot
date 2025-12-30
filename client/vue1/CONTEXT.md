# BounceBot Vue3 Client - Project Context

This document provides context for understanding the project and the decisions made during implementation. Useful for resuming work or onboarding new contributors.

## Project Overview

**BounceBot** is a Ricochet Robots-style puzzle game:
- 16x16 grid with internal walls
- 4 colored robots that slide until hitting obstacles (edges, walls, other robots)
- Goal: Move the target robot to the target cell in minimum moves
- Multiplayer: Compete in rooms to find the shortest solution fastest

## Architecture

### Backend (Go)
- Location: `/server/`, `/model/`, `/proto/`
- Connect RPC server for browser access (port 8080)
- Game logic in `/model/` - board, robots, move validation, physics
- Room management in `/server/room/` - multiplayer rooms, players, solutions
- Proto definitions in `/proto/bouncebot.proto`

### Frontend (Vue3)
- Location: `/client/vue1/`
- Vue 3 with Composition API
- TypeScript for type safety
- Vite for build/dev server
- Pinia for state management
- Connect RPC client for server communication
- WebSocket for real-time updates

## Key Decisions

### Why `client/vue1/` directory?
The `vue1` name allows for potential alternative frontend implementations in the future (e.g., `vue2`, `react1`). The project was built incrementally with small PRs that each showed visible progress.

### Why Vue 3 + TypeScript + Vite?
- Vue 3: Modern reactive framework, good for interactive UIs
- TypeScript: Type safety, better IDE support, matches proto types
- Vite: Fast dev server with HMR, simple configuration

### Why Connect protocol?
- gRPC requires a proxy for browser access
- Connect is browser-native HTTP that speaks the same proto definitions
- Avoids needing envoy/grpc-web proxy

### Why incremental steps?
- Each step is a reviewable PR
- Visible progress in browser at each step
- Easier to course-correct if needed
- Better learning opportunity

## File Structure

```
client/vue1/
├── index.html              # HTML entry point
├── src/
│   ├── main.ts             # Vue app initialization
│   ├── App.vue             # Root component
│   ├── router.ts           # Vue Router configuration
│   ├── constants.ts        # Shared constants (colors, board size)
│   ├── gamePhysics.ts      # Robot movement physics
│   ├── components/         # Reusable UI components
│   ├── composables/        # Vue composition functions
│   ├── views/              # Route views (HomeView, RoomView)
│   ├── stores/             # Pinia state stores
│   ├── services/           # API clients (Connect, WebSocket)
│   └── gen/                # Generated protobuf types
├── IMPLEMENTATION_PLAN.md  # Full step-by-step plan
├── PROGRESS.md             # Completed steps tracker
└── CONTEXT.md              # This file
```

## Conventions

### Component Naming
- `App.vue` stays as root component (Vue convention)
- Feature components get descriptive names: `GameBoard.vue`, `RobotMarker.vue`

### Styling
- Dark theme by default (matches game aesthetic)
- Scoped styles in components
- Global styles only in `style.css`

## Useful Commands

```bash
# Start dev server
cd client/vue1
npm run dev

# Build for production
npm run build

# Type check
npm run type-check
```

## Links

- Implementation Plan: `./IMPLEMENTATION_PLAN.md`
- Progress Tracker: `./PROGRESS.md`
- Proto definitions: `/proto/bouncebot.proto`
- Go game model: `/model/game.go`
