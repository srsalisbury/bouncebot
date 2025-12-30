# Development Context

Implementation notes for developers working on the Vue client. For project overview and how to run, see the root [README.md](../../README.md).

## Code Layout

```
client/vue1/src/
├── main.ts             # Vue app initialization
├── App.vue             # Root component (router outlet)
├── router.ts           # Route definitions (/, /room/:roomId)
├── config.ts           # Runtime configuration (API URLs)
├── constants.ts        # Shared constants (colors, board size)
├── gamePhysics.ts      # Robot movement physics (shared with backend tests)
│
├── views/              # Route-level components
│   ├── HomeView.vue    # Create/join room
│   └── RoomView.vue    # Game room (lobby + game)
│
├── components/         # Reusable UI components
│   ├── GameBoard.vue   # Main game board with robots, walls, target
│   ├── PlayersPanel.vue    # Player list and leaderboard
│   ├── HowToPlayModal.vue  # Help modal
│   └── LeaderboardModal.vue
│
├── composables/        # Vue composition functions
│   ├── useGameInput.ts     # Keyboard event handling
│   ├── useReplay.ts        # Solution replay/unwind animations
│   ├── useRoomConnection.ts # Room loading and WebSocket setup
│   └── useGameActions.ts   # Solution submission API calls
│
├── stores/             # Pinia state stores
│   ├── gameStore.ts    # Game state (robots, walls, moves, solutions)
│   └── roomStore.ts    # Room state (player ID, name persistence)
│
├── services/           # External communication
│   ├── connectClient.ts    # Connect RPC client
│   ├── websocket.ts        # WebSocket for real-time events
│   └── AnimationService.ts # Animation timing constants
│
└── gen/                # Generated from proto (do not edit)
    └── bouncebot_pb.ts
```

## Key Decisions

### Why `client/vue1/` directory?
Allows for alternative frontend implementations (e.g., `react1`). Built incrementally with small PRs.

### Why Connect protocol?
gRPC requires a proxy for browsers. Connect is browser-native HTTP using the same proto definitions.

### Why separate composables?
Extracted from large components (GameBoard.vue, RoomView.vue) during refactoring to improve testability and reduce coupling.

## Conventions

### Component Naming
- Feature components: `GameBoard.vue`, `PlayersPanel.vue`
- Composables: `useFeatureName.ts`

### Styling
- Dark theme by default
- Scoped styles in components
- Global styles only in `style.css`

## Related Documentation

- [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) - Original step-by-step build plan
- [PROGRESS.md](./PROGRESS.md) - Completed steps and PR history
