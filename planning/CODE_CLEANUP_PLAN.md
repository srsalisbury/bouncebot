# Code Cleanup Plan

A comprehensive list of cleanup and refactoring opportunities for the BounceBot codebase.

---

## Frontend (Vue Client)

### High Priority

#### 1. Split Large Components

**~~GameBoard.vue (1,297 lines)~~** ✅ PR #160
- ~~Extract solutions panel into `GameBoardSolutions.vue`~~

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

#### ~~3. Move Hardcoded Values to Constants~~ ✅ PR #159

### 4. Standardize Error Handling

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

#### ~~3. Refactor Service Lock Pattern~~ ✅ PR #161, #164

~~Extract to `withRoomLock(roomID, fn)` helper~~

### Medium Priority

#### ~~4. Split Large Files~~ ✅ PR #162

~~**`service.go`** - Split signal processing into `service_signals.go`~~

#### 5. Extract Duplicate String Helpers

**`server/room/room.go`** (lines 66-78):
- `containsString()`
- `removeStringAt()`

→ Move to `server/room/helpers.go` or `server/util/strings.go`

#### 6. Reduce Function Complexity

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

#### ~~8. Test Coverage Gaps~~ ✅ PR #166

~~Added tests for `bouncebotserver.go` and `solver/astar/astar.go`~~

Remaining gaps (lower priority):
- `solver/benchmark/*` (benchmark tooling)

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
| Frontend | 2 items | 2 items | 3 items |
| Backend | 2 items | 2 items | 3 items |

### Completed PRs

- PR #159: Move hardcoded values to constants (frontend)
- PR #160: Split GameBoard.vue (extract solutions panel)
- PR #161: Refactor service lock pattern
- PR #162: Split service.go (extract service_signals.go)
- PR #164: Simplify withRoomLock (combine with withRoomLockErr)
- PR #166: Add test coverage (bouncebotserver, A* solver)
