# Code Cleanup Plan

A comprehensive list of cleanup and refactoring opportunities for the BounceBot codebase.

---

## Frontend (Vue Client)

### High Priority

#### 1. Split Large Components

**GameBoard.vue (1,297 lines)**
- Extract board rendering into `GameBoardContent.vue`
- Extract solutions panel into `GameBoardSolutions.vue`
- Move style helper functions to `gameBoardStyles.ts`
- Move 700+ lines of styles to separate file or use CSS modules

**RoomView.vue (1,276 lines)**
- Extract game header into `GameHeaderBar.vue` with solo/multiplayer/ended variants
- Create `useKeyboardEventManager.ts` composable for centralized keyboard handling
- Extract dialog management into separate composables

**PlayersPanel.vue (559 lines)**
- Extract timer logic into `useGameTimer.ts` composable

#### 2. Extract Duplicate Code

**`isMobile()` function** - duplicated in 3 files:
- `SolutionsDrawer.vue`
- `PlayerSolutionsDrawer.vue`
- `LeaderboardModal.vue`

→ Extract to `src/composables/useIsMobile.ts`

**`formatSolveTime()` function** - duplicated in:
- `GameBoard.vue` (lines 193-201)
- `PlayerSolutionsDrawer.vue` (lines 80-88)

→ Extract to `src/utils/timeUtils.ts`

**`MoveWithDirection` interface** - duplicated in:
- `useReplay.ts`
- `PlayerSolutionsDrawer.vue`

→ Define once in `src/types/index.ts`

### Medium Priority

#### 3. Move Hardcoded Values to Constants

| Value | Location | Suggested Constant |
|-------|----------|-------------------|
| `1000` (undo hold timer) | `GameBoard.vue:103` | `UNDO_HOLD_DURATION_MS` |
| `300` (double-tap threshold) | `GameBoard.vue:217` | `DOUBLE_TAP_THRESHOLD_MS` |
| `30 * 60` (max timer) | `PlayersPanel.vue:30` | `MAX_GAME_TIMER_SECONDS` |
| `3000` (poll interval) | `useRoomConnection.ts:211` | `ROOM_POLL_INTERVAL_MS` |
| `500` (reset delay) | `gameStore.ts` | `SOLUTION_SWITCH_DELAY_MS` |
| `'__solver__'` | `RoomView.vue:117` | `SOLVER_PLAYER_ID` |
| `'__solo__'` | `RoomView.vue:119` | `SOLO_PLAYER_ID` |
| `6/5` (aspect ratio) | Multiple files | `MOBILE_ASPECT_RATIO` |
| `1050` (width breakpoint) | Multiple files | `MOBILE_WIDTH_BREAKPOINT` |

#### 4. Standardize Error Handling

- Create `useErrorHandler.ts` composable for centralized error handling
- Some components log errors, some swallow silently - make consistent
- Add retry logic where appropriate

#### 5. Remove/Wrap Console Logs

14 console.log statements to address:
- `websocket.ts` (lines 89, 94, 102, 110, 118, 130, 139, 143)
- `useRoomConnection.ts` (line 179)
- `useGameActions.ts` (line 76)
- `RoomView.vue` (lines 73, 265)

→ Create environment-based logger utility

### Low Priority

#### 6. Extract gameStore Persistence

- Move localStorage persistence logic from `gameStore.ts` to `gamePersistenceService.ts`
- Extract animation/timeout management to helper

#### 7. Type Safety Improvements

- Add dedicated types for special player IDs
- Reduce `any`/`unknown` usage (16 instances)

#### 8. Documentation

- Add JSDoc comments to complex functions (e.g., input forgiveness logic in gameStore)

---

## Backend (Go Server)

### High Priority

#### 1. Extract Duplicated Solver Code

**Duplicated in `solver/astar/` and `solver/bfs/`:**
- `copyBots()` function
- `isWin()` function
- `encodeState()` function

→ Extract to `solver/common.go`

#### 2. Convert Panics to Error Returns

**`model/games.go`** uses panic in non-emergency contexts:
- Line 24: `panic("all Panels must have the same Size")`
- Line 155: `panic(fmt.Sprintf("unknown panel id: %d", id))`
- Line 170: `panic(err)`
- Lines 190, 265: `panic("board has no possible targets")`

→ Convert to proper error returns

#### 3. Refactor Service Lock Pattern

**`server/room/service.go`** repeats this pattern 8+ times:
```go
room, unlock := s.repo.GetWithLock(roomID)
if room != nil {
    // do work
    unlock()
}
// process signals
```

→ Extract to `withRoomLock(roomID, fn)` helper

### Medium Priority

#### 4. Split Large Files

**`service.go` (459 lines)** - combines too many responsibilities:
- Signal processing
- Public API
- Persistence management
- Room cleanup

→ Split into:
- `service.go` - API and orchestration
- `persistence_service.go` - Auto-save, cleanup, load/save
- `signal_processor.go` - Signal handling

**`model/game.go` (378 lines)**:
→ Split into:
- `game.go` - Core struct and public API
- `game_movement.go` - Movement computation
- `game_solution.go` - Solution verification

**`model/games.go` (293 lines)**:
→ Split into:
- `game_generator.go` - NewRandomGame, NewContinuationGame
- `board_panels.go` - Panel definitions

#### 5. Extract Duplicate String Helpers

**`server/room/room.go`** (lines 66-78):
- `containsString()`
- `removeStringAt()`

→ Move to `server/room/helpers.go` or `server/util/strings.go`

#### 6. Reduce Function Complexity

**`RoomService.processSignals()`** - large switch with 5 cases
→ Extract each case to separate handler methods

**`PlayerManager.RemovePlayer()`** - handles too many concerns
→ Split into `cleanupPlayerState()` and `determineGameStateSignals()`

**`GameLifecycle.StartGame()` and `StartNextGame()`** - 70% code duplication
→ Extract common logic to `generateNextGame()` helper

### Low Priority

#### 7. Hardcoded Values to Constants

| Value | Location | Suggested Constant |
|-------|----------|-------------------|
| `30*time.Second` | `main.go:87` | `config.SolverTimeout` |
| `256` (buffer size) | `ws/hub.go:320` | `WSClientBufferSize` |

#### 8. Test Coverage Gaps

Missing tests for:
- `bouncebotserver.go` (105 lines)
- `solver/solver.go`, `solver/registry.go`, `solver/manager.go`
- `model/game.go`, `model/games.go`, `model/render.go`, `model/board.go`
- `solver/benchmark/*` (>500 lines)

#### 9. Remove Dead Code

- `model/game.go:370` - commented out error return
- Incomplete mock in `helpers_test.go`

#### 10. Standardize Error Messages

Error messages use inconsistent formats throughout `service.go`:
- "room not found: %s"
- "cannot join single player room"
- "only host can change settings"

→ Create constant error messages or error types package

---

## Summary

| Area | High Priority | Medium Priority | Low Priority |
|------|--------------|-----------------|--------------|
| Frontend | 2 items | 3 items | 3 items |
| Backend | 3 items | 3 items | 4 items |

### Suggested Order of Implementation

1. Extract duplicated solver code (backend) - quick win
2. Move hardcoded values to constants (both) - easy, reduces bugs
3. Split GameBoard.vue (frontend) - biggest impact on maintainability
4. Refactor service lock pattern (backend) - reduces boilerplate
5. Convert panics to errors (backend) - improves reliability
6. Extract duplicate frontend utilities - reduces duplication
7. Split large backend files - improves navigation
8. Add missing tests - improves confidence
