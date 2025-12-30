# Code Refactoring Plan

## Goals

- Emphasize clarity with small, single-purpose functions and files
- Improve testability with tests for critical paths
- Reduce coupling between components

---

## Backend Refactoring

### Phase B1: Split room.go (813 lines → ~4 files)

**Current**: `server/room/room.go` is a god object handling rooms, players, solutions, and game lifecycle.

**Extract to**:

1. **`server/room/solution.go`** (~150 lines)
   - `SolutionManager` struct
   - `SubmitSolution()`, `RetractSolution()`, `GetBestSolution()`
   - Solution history tracking
   - **Tests**: Solution submission, retraction, history restoration

2. **`server/room/player.go`** (~120 lines)
   - `PlayerManager` struct
   - `AddPlayer()`, `RemovePlayer()`, `DisconnectPlayer()`, `ReconnectPlayer()`
   - Disconnect timer management
   - **Tests**: Disconnect grace period, reconnection

3. **`server/room/lifecycle.go`** (~100 lines)
   - `StartGame()`, `EndGame()`, `startNextGame()`
   - Winner determination logic
   - Game continuation vs new game logic
   - **Tests**: Game transitions, winner selection

4. **`server/room/room.go`** (~200 lines, reduced)
   - `Room` struct and `Store` struct
   - `CreateRoom()`, `JoinRoom()`, `GetRoom()`
   - Coordinates the managers above

### Phase B2: Simplify hub.go broadcast methods

**Current**: 9 nearly identical broadcast methods (lines 138-217)

**Refactor to**:
```go
// Generic event broadcaster
func (h *Hub) BroadcastEvent(roomID string, eventType string, payload any) {
    h.Broadcast(roomID, Event{Type: eventType, Payload: payload})
}
```

Keep type-safe wrappers but have them call the generic method.

**Add tests**: `server/ws/hub_test.go`
- Client registration/unregistration
- Broadcast delivery
- Room isolation (messages only go to correct room)

### Phase B3: Extract game helpers from model/

**`model/game.go`** - Extract:
- `checkBotCollision(pos Position) *BotId` - separate method
- `findBlockingWall(pos Position, dir Direction) bool` - cleaner wall check

**`model/games.go`** - Extract:
- `selectRandomTarget(board Board, bots map[BotId]Position) BotPosition`
- `placeBots(board Board, target BotPosition) map[BotId]Position`

**Tests**: Already have good coverage, ensure extracted functions are tested.

---

## Frontend Refactoring

### Phase F1: Split GameBoard.vue (924 lines → ~4 files)

**Current**: Rendering, input handling, replay logic, and game state all in one component.

**Extract to**:

1. **`composables/useGameInput.ts`** (~80 lines)
   - Keyboard event handling
   - Input blocking logic
   - Direction mapping
   - **Tests**: Key mapping, input blocking states

2. **`composables/useReplay.ts`** (~100 lines)
   - `unwindThenReplay()` logic
   - Position history building
   - Replay state management
   - **Tests**: Replay sequencing, position history accuracy

3. **`components/GameBoardRenderer.vue`** (~300 lines)
   - Pure rendering: grid, walls, robots, target
   - No game logic, just props → visual output

4. **`components/GameBoard.vue`** (~200 lines, reduced)
   - Orchestrates composables
   - Connects to store
   - Handles solution switching

### Phase F2: Split RoomView.vue (906 lines → ~3 files)

**Extract to**:

1. **`composables/useRoomConnection.ts`** (~100 lines)
   - Room loading and joining
   - WebSocket connection setup
   - Reconnection logic
   - **Tests**: Connection state management

2. **`composables/useGameActions.ts`** (~80 lines)
   - `submitSolution()`, `retractSolution()`
   - `markFinishedSolving()`, `markReadyForNext()`
   - API call wrappers

3. **`RoomView.vue`** (~400 lines, reduced)
   - Template and layout
   - Uses composables for logic

### Phase F3: Extract animation logic from gameStore.ts

**Current**: setTimeout calls scattered in store actions

**Extract to**: `services/AnimationService.ts` (~60 lines)
- `animateUnwind(callback, delay)`
- `animateReplay(positions, onStep, onComplete)`
- Centralized timing constants

---

## Files Summary

### Backend - Create
| File | Purpose |
|------|---------|
| `server/room/solution.go` | Solution tracking and history |
| `server/room/player.go` | Player lifecycle management |
| `server/room/lifecycle.go` | Game state transitions |
| `server/ws/hub_test.go` | WebSocket hub tests |

### Backend - Modify
| File | Changes |
|------|---------|
| `server/room/room.go` | Reduce to room CRUD, delegate to managers |
| `server/ws/hub.go` | Generic broadcast helper |
| `model/game.go` | Extract collision/wall helpers |
| `model/games.go` | Extract target selection, bot placement |

### Frontend - Create
| File | Purpose |
|------|---------|
| `composables/useGameInput.ts` | Keyboard handling |
| `composables/useReplay.ts` | Replay/unwind logic |
| `composables/useRoomConnection.ts` | Room/WebSocket setup |
| `composables/useGameActions.ts` | Game API calls |
| `services/AnimationService.ts` | Centralized animation timing |
| `components/GameBoardRenderer.vue` | Pure rendering component |

### Frontend - Modify
| File | Changes |
|------|---------|
| `components/GameBoard.vue` | Orchestration only, use composables |
| `views/RoomView.vue` | Use composables, reduce to template |
| `stores/gameStore.ts` | Remove animation logic |

---

## Implementation Order

### Week 1: Backend
1. Extract `server/room/solution.go` + tests
2. Extract `server/room/player.go` + tests
3. Extract `server/room/lifecycle.go` + tests
4. Add `server/ws/hub_test.go`
5. Refactor hub.go broadcast methods

### Week 2: Frontend
6. Create `composables/useGameInput.ts` + tests
7. Create `composables/useReplay.ts` + tests
8. Create `GameBoardRenderer.vue`, simplify `GameBoard.vue`
9. Create `composables/useRoomConnection.ts`
10. Create `composables/useGameActions.ts`
11. Simplify `RoomView.vue`
12. Extract `AnimationService.ts`

---

## Testing Strategy

**Critical paths requiring tests**:
- Solution submission/retraction/history (backend)
- Player disconnect/reconnect grace period (backend)
- Game lifecycle transitions (backend)
- WebSocket broadcast delivery (backend)
- Keyboard input mapping (frontend)
- Replay position sequencing (frontend)

**Skip tests for**:
- Simple extraction refactors with no logic changes
- Pure rendering components (visual testing preferred)
- Trivial wrappers/delegates
