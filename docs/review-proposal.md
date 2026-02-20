# Codebase Review Proposal

Generated 2026-02-19. This is a proposal, not a commit. Nothing gets changed until specific items are approved.

## Methodology

Four discovery agents reviewed the full codebase in parallel (security, correctness, performance, testing gaps), producing 81 findings. Three evaluator agents then scored every finding on a 1-5 scale through different lenses:

- **Pragmatist**: ship quality, minimize churn, effort-to-impact ratio
- **Purist**: correctness, long-term maintainability, defense in depth
- **Product-Minded**: user impact, data integrity, trust

Findings are sorted by weighted average score (all three evaluators weighted equally). Duplicate/overlapping findings are merged.

---

## Tier 1: Must Fix (avg 4.67-5.0)

These have unanimous or near-unanimous agreement across all three evaluators.

### 1. SEC-001 / Path Traversal in Daily Progress Storage
**Score: 5.0 (5/5/5) | Category: Security | Severity: Critical**

**What**: The `playerID` from client requests is used directly in `filepath.Join(dataDir, "users", prefix[:2], playerID+".json")` with zero validation. An attacker can supply a playerID like `../../etc/passwd` to read or write arbitrary files on the server filesystem.

**Where**: `server/daily/progress.go:150`

**Suggested fix**: Validate playerID against a strict regex like `^[a-f0-9]{16,64}$`. Reject any playerID containing path separators or dots. Optionally verify the resolved path is within the expected data directory.

**Pros**: Closes a textbook critical vulnerability. Single-file fix. Trivial effort.
**Cons**: None. Pure upside.

**Evaluator notes**: Unanimous 5/5/5. No disagreement.

---

### 2. SEC-003 + TEST-005 / Daily Solution Validation Skips Physics
**Score: 4.67+5.0 (4-5/5/5) | Category: Security + Testing | Severity: Critical**

**What**: The `validateSolution` function in the daily challenge endpoint simply places bots at the positions specified in the submitted moves and checks if the target bot ends at the target position. It does NOT verify that each move follows physics rules (sliding until hitting a wall or bot). A player can submit a single move that teleports the target bot directly to the goal. The multiplayer room path correctly uses `game.CheckSolution()`.

**Where**: `server/bouncebotserver.go:295`

**Suggested fix**: Replace the custom `validateSolution` with a call to `game.CheckSolution()`, which is already used by the multiplayer room path and properly validates physics. Add a test that submits a move where a robot passes through a wall and verifies rejection.

**Pros**: Restores integrity of the entire daily challenge feature. Reuses existing validated code. Small change.
**Cons**: None. The correct validation logic already exists.

**Evaluator notes**: Unanimous that this makes the daily challenge feature "meaningless" without the fix. Pragmatist and purist agree the fix is trivial.

---

### 3. BUG-001 / Double-Counting Wins
**Score: 5.0 (5/5/5) | Category: Correctness | Severity: Critical**

**What**: `StartGame` credits wins and increments `GamesPlayed` for the previous game (lines 48-58), but `EndGame` also credits wins and increments `GamesPlayed` (lines 150-156). If the normal flow runs `EndGame` then `StartGame` is called for the next round, wins from the previous game get counted twice.

**Where**: `server/room/game_lifecycle_manager.go:48` and `:150`

**Suggested fix**: Add a guard (e.g., a `gameEnded` flag or check if `GamesPlayed` was already incremented for the current round) so wins are credited exactly once. Alternatively, remove win-crediting from `StartGame` entirely and ensure `EndGame` is always called before transitioning. Add a regression test (TEST-011, TEST-015).

**Pros**: Fixes user-visible scoreboard corruption. Single file change.
**Cons**: Need to carefully trace both code paths to ensure exactly-once semantics. Low risk.

**Evaluator notes**: Unanimous 5/5/5. "Erodes trust in the game's fairness."

---

### 4. BUG-002 + PERF-010 / Broadcast Race Condition (Server Crash)
**Score: 5.0+4.0 (5/5/5 + 4/4/4) | Category: Correctness + Performance | Severity: Critical**

**What**: In `Broadcast`, the `RLock` is released (line 243) before the map iteration begins (line 245). Between releasing the lock and iterating, other goroutines can modify `h.rooms[roomID]` (register/unregister adding/deleting clients). This is a concurrent map read/write, which causes panics in Go. Additionally, `h.unregister(client)` is called during iteration (line 250), which acquires a write lock and modifies the map being iterated.

**Where**: `server/ws/hub.go:241-251`

**Suggested fix**: Either hold the `RLock` for the entire broadcast loop, or make a snapshot copy of the clients map under the lock and iterate the copy. Collect clients that need unregistering and do it after the loop completes.

**Pros**: Fixes a crash bug that takes down the entire server, disconnecting all players across all rooms.
**Cons**: Holding the lock during broadcast increases lock contention slightly. Snapshot copy approach has a small allocation. Both are trivial compared to a crash.

**Evaluator notes**: Unanimous 5/5/5 on the bug. All evaluators noted this is a crash risk. PERF-010 is the same issue from the performance angle.

---

### 5. BUG-003 / WebSocket Reconnection Race
**Score: 4.67 (4/5/5) | Category: Correctness | Severity: Critical**

**What**: When a player reconnects via WebSocket, there's a race between the new connection calling `ReconnectPlayer` and the old connection's `readPump` calling `unregister` -> `DisconnectPlayer`. Sequence: (1) new WS connects, checks player status = Disconnected, (2) calls `ReconnectPlayer` (sets Connected, cancels timer), (3) old connection's readPump finally calls `DisconnectPlayer` (sets Disconnected, starts removal timer). Now the player has an active WebSocket but is marked Disconnected with a removal timer running.

**Where**: `server/ws/hub.go:126, :284`

**Suggested fix**: Track the "current" client per player. In `unregister`, check if the client being unregistered is still the active client before calling `DisconnectPlayer`. Use a generation counter or compare client pointers so stale disconnects are ignored.

**Pros**: Fixes a real user-facing bug on flaky connections (mobile, poor wifi). Players would be unable to play without leaving and rejoining.
**Cons**: Requires careful sequencing changes. Moderate effort but high reliability impact.

**Evaluator notes**: Pragmatist scored 4 (noting effort), purist and product both scored 5.

---

### 6. PERF-001 / ReorderSolution Factorial Blowup
**Score: 5.0 (5/5/5) | Category: Performance | Severity: Critical**

**What**: `ReorderSolution` generates ALL valid interleavings of solution moves via recursive backtracking. The number of interleavings is multinomial: for a 12-move solution split evenly across 4 bots, this produces 369,600 interleavings, each stored in memory. A 14-move even split produces 63 million. This runs on every solver completion.

**Where**: `solver/reorder.go:29`

**Suggested fix**: Replace exhaustive enumeration with a greedy approach: always pick the next move from the same bot if valid, only switch bots when forced. This reduces from factorial to linear. If optimality matters, use branch-and-bound with early termination. Add a performance test (TEST-010) that verifies completion under a timeout with a 12+ move solution.

**Pros**: Prevents server hangs and potential OOM on non-trivial puzzles. Critical for production reliability.
**Cons**: Greedy approach may not find the globally optimal reordering, but "good enough" reordering is better than hanging.

**Evaluator notes**: Unanimous 5/5/5. "Latent DoS." "Must fix before production load."

---

## Tier 2: High Priority (avg 3.67-4.0)

### 7. TEST-001 / Daily Package Has Zero Tests
**Score: 4.0 (4/4/4) | Category: Testing | Severity: Critical**

**What**: The entire `server/daily/` package (Manager, ProgressManager, ClassifyDifficulty) has no test files. This package handles puzzle generation with seeded randomness, user progress persistence, caching, and concurrent access, all user-facing.

**Where**: `server/daily/manager.go`, `server/daily/progress.go`, `server/daily/puzzle.go`

**Suggested fix**: Create `manager_test.go` and `progress_test.go` covering: ClassifyDifficulty boundaries, dateToSeed reproducibility, GetPuzzlesForDate cache behavior, MarkSolved/IsSolved round-trip, concurrent access, and seeded game determinism.

**Pros**: Catches SEC-001 and BUG-011 via tests. Prevents regressions in a new, user-facing feature. Tests can be written alongside the bug fixes.
**Cons**: Moderate effort (new test files from scratch). May need test fixtures for file I/O.

**Evaluator notes**: Unanimous 4/4/4. All agree the daily package's lack of tests is the most important testing gap.

---

### 8. SEC-002 / Daily Challenge Player ID Spoofing
**Score: 3.67 (3/4/4) | Category: Security | Severity: Critical**

**What**: `GetDailyChallenge` and `SubmitDailySolution` accept a client-supplied `player_id` with no authentication. Any user can view or modify any other user's daily challenge progress by supplying their player_id.

**Where**: `server/bouncebotserver.go:178`

**Suggested fix**: At minimum, ensure the client-generated player_id is a cryptographically random UUID stored in localStorage (so it cannot be guessed). Better: generate a server-side token on first daily challenge use, return it to the client, and require it for subsequent calls. Add rate limiting to prevent brute-force discovery.

**Pros**: Prevents griefing and progress tampering. Protects feature integrity.
**Cons**: Pragmatist notes this properly requires designing authentication, which is a larger effort. The UUID approach is a quick partial fix.

**Evaluator notes**: Pragmatist scored 3 ("flag for auth epic"), purist 4, product 4. Disagreement on scope: quick fix vs. proper auth.

---

### 9. PERF-002 / Solver Job Maps Grow Without Bound
**Score: 3.67 (4/4/3) | Category: Performance | Severity: Critical**

**What**: The solver Manager's `jobs` and `roomJobs` maps grow indefinitely. Every solver job is stored permanently. `CleanupRoom` only removes the latest job for a specific room, so older jobs leak. Over time, both maps accumulate entries proportional to total games ever played.

**Where**: `solver/manager.go:37`

**Suggested fix**: Add a periodic cleanup goroutine that removes jobs older than 1 hour, or limit maps to a fixed size with LRU eviction. Ensure `CleanupRoom` is called during stale room cleanup.

**Pros**: Prevents slow OOM on long-running servers. Simple periodic cleanup.
**Cons**: Minor implementation effort. Need to decide TTL policy.

**Evaluator notes**: Pragmatist and purist scored 4, product scored 3 (noting it's a slow leak with no immediate user impact).

---

### 10. TEST-004 / Physics Test Coverage Is Thin
**Score: 3.67 (3/4/4) | Category: Testing | Severity: Critical**

**What**: The shared `physics_cases.json` has only 16 basic test cases. No coverage for: multiple walls in a path, walls adjacent to board edges, corner cells, chains of robots blocking each other, or L-shaped wall patterns. Any of these could cause client/server physics divergence.

**Where**: `tests/physics_cases.json`, `model/game_movement.go`, `client/web/src/services/gamePhysics.ts`

**Suggested fix**: Expand `physics_cases.json` with cases for: gap between parallel walls, wall adjacent to board edge, corner positions (0,0 and 15,15), three robots in a line, robot blocked by both wall and robot at different distances. Both Go and TypeScript tests automatically pick up new cases.

**Pros**: Physics is the core game mechanic. More cases prevent subtle cross-language divergence. New cases are automatically tested on both sides.
**Cons**: Need to carefully construct valid test boards for each case. Moderate effort.

**Evaluator notes**: Purist and product scored 4, pragmatist scored 3. All agree physics correctness matters.

---

### 11. TEST-010 / ReorderSolution Needs Performance Test
**Score: 4.0 (4/4/4) | Category: Testing | Severity: Moderate**

**What**: There is no test that verifies `ReorderSolution` completes in reasonable time with larger inputs. Without this, PERF-001's factorial blowup ships undetected and could regress after a fix.

**Where**: `solver/reorder.go`

**Suggested fix**: Add `TestReorderSolution_LargeInput` that creates a 12+ move solution across 3-4 bots, runs with a 5-second timeout, and verifies completion. Also add `TestReorderSolution_SingleBot` and `TestReorderSolution_EmptyInput`.

**Pros**: Essential companion to the PERF-001 fix. Prevents regression.
**Cons**: Requires constructing a valid game and solution for the test fixture. Should be done alongside PERF-001.

**Evaluator notes**: Unanimous 4/4/4. All tie this directly to PERF-001.

---

## Tier 3: Medium Priority (avg 3.0-3.33)

### 12. SEC-008 / WebSocket No Size Limit or Read Deadline
**Score: 3.33 (3/4/3) | Category: Security | Severity: Moderate**

**What**: The WebSocket `readPump` does not call `conn.SetReadLimit()` or `conn.SetReadDeadline()`. A malicious client can send extremely large messages to consume server memory, or hold connections open indefinitely.

**Where**: `server/ws/hub.go:314`

**Suggested fix**: Add `c.conn.SetReadLimit(4096)` (server doesn't process client messages) and implement ping/pong handling with a 60-second read deadline.

**Pros**: Two-line fix that prevents a real DoS vector. Excellent effort-to-impact ratio.
**Cons**: None meaningful.

**Evaluator notes**: Purist scored 4 ("trivial fix"), pragmatist and product scored 3.

---

### 13. BUG-004 + BUG-021 + BUG-022 / Lock-Release-Before-Use Pattern
**Score: 3.33 each (3-4/3-4/3) | Category: Correctness | Severity: Moderate**

**What**: Multiple methods release the room lock and then either return a raw room pointer or read room data without a lock. This is a systemic pattern:
- **BUG-004**: `ValidateSessionToken` reads `room.Players` without holding a lock (`service.go:114`)
- **BUG-021**: `StartGame` returns raw room pointer after unlocking (`service.go:150`)
- **BUG-022**: `HandleWebSocket` reads player status outside any lock (`hub.go:269`)

**Suggested fix**: Fix as a batch. Either: (a) move session validation inside the room lock, (b) return serialized copies (proto) instead of raw pointers while still holding the lock, or (c) perform read-then-act operations within a single locked scope. These are the same class of bug and should be addressed together.

**Pros**: Eliminates an entire class of data races. Prevents panics and corrupted reads.
**Cons**: Requires restructuring lock scopes across several functions. Moderate effort.

**Evaluator notes**: Purist consistently scored 4 ("must hold the lock during the read"). Pragmatist and product scored 3, noting the races are narrow but real.

---

### 14. BUG-006 / Join Lock Release Before processSignals
**Score: 3.33 (3/4/3) | Category: Correctness | Severity: Moderate**

**What**: In `Join()`, the room lock is released (line 94) before `processSignals` is called (line 100). Between unlock and processSignals, other operations could modify the room. The returned room pointer is also read by the caller without a lock.

**Where**: `server/room/service.go:93-101`

**Suggested fix**: Return a proto/copy from Join while still holding the lock, consistent with how StartGame creates a roomCopy. Process signals inside the locked scope or immediately after unlock with a copied signal list.

**Pros**: Closes a TOCTOU window in a frequently-called path.
**Cons**: Single function reorder. Low risk.

**Evaluator notes**: Purist scored 4, others scored 3.

---

### 15. BUG-011 / Daily Progress Concurrent Write Race
**Score: 3.33 (3/3/4) | Category: Correctness | Severity: Moderate**

**What**: `MarkSolved` performs a non-atomic read-modify-write on user progress files. Two concurrent calls for the same player could both read the same file, each mark a different puzzle as solved, then both write, with the second overwriting the first's changes.

**Where**: `server/daily/progress.go:102`

**Suggested fix**: Use a per-player lock (the existing mutex is global) to make MarkSolved's read-modify-write atomic. Alternatively, use file locking or an atomic rename pattern.

**Pros**: Prevents loss of earned progress, one of the worst user experiences.
**Cons**: Requires adding a lock map keyed by playerID. Small effort.

**Evaluator notes**: Product scored 4 ("losing progress is one of the worst UX failures"). Others scored 3.

---

### 16. SEC-006 / Rate Limiting Gaps
**Score: 3.0 (3/3/3) | Category: Security | Severity: Moderate**

**What**: Rate limiting is only applied to `GetRoom`. All mutation endpoints (CreateRoom, JoinRoom, StartGame, SubmitDailySolution, etc.) have no rate limiting.

**Where**: `server/main.go:73`

**Suggested fix**: Add per-IP rate limits to at least `CreateRoom` (5/min), `JoinRoom` (10/min), and `SubmitDailySolution` (30/min). Consider a baseline global rate limiter middleware.

**Pros**: Prevents resource exhaustion via room spam or solution flooding.
**Cons**: Need to choose appropriate limits. Moderate implementation effort.

**Evaluator notes**: Unanimous 3/3/3. All agree CreateRoom and SubmitDailySolution are the most important gaps.

---

### 17. SEC-010 / No Input Validation on Player Names
**Score: 3.0 (3/3/3) | Category: Security | Severity: Moderate**

**What**: Player names from CreateRoom and JoinRoom are passed through with no length limits, character restrictions, or sanitization.

**Where**: `server/bouncebotserver.go:33`

**Suggested fix**: Enforce max length (30 chars), reject empty names, strip control characters.

**Pros**: One-line guard that prevents memory abuse and UI disruption.
**Cons**: None meaningful.

**Evaluator notes**: Unanimous 3/3/3.

---

### 18. SEC-012 / No Date Validation in Daily Solution Submission
**Score: 3.0 (3/3/3) | Category: Security | Severity: Moderate**

**What**: `SubmitDailySolution` accepts any date string. Players can submit solutions for past dates retroactively or for future dates.

**Where**: `server/bouncebotserver.go:235`

**Suggested fix**: Validate the date matches the player's current date (computed from timezone offset). Reject submissions for dates more than a day off.

**Pros**: Preserves daily challenge integrity. Simple validation.
**Cons**: Requires sending timezone offset in the submit request.

**Evaluator notes**: Unanimous 3/3/3.

---

### 19. BUG-007 / Pending Players Not Cleaned Up on Disconnect
**Score: 3.0 (2/3/4) | Category: Correctness | Severity: Moderate**

**What**: `DisconnectPlayer` only searches `room.Players`, not `PendingPlayers`. If a pending player's WebSocket disconnects, no disconnect timer is started, and they remain in `PendingPlayers` indefinitely.

**Where**: `server/room/player_manager.go:97`

**Suggested fix**: Extend `DisconnectPlayer` to also check `PendingPlayers`. For pending players, remove them immediately (since they haven't participated in a game) rather than starting a grace timer.

**Pros**: Prevents ghost pending players from blocking room capacity.
**Cons**: Small code change. Need to handle the PendingPlayers slice removal.

**Evaluator notes**: Product scored 4 (visible to host), purist 3, pragmatist 2 ("cleaned up on room expiry").

---

### 20. BUG-016 / Undo Doesn't Cancel Pending Move Timeout
**Score: 3.0 (3/3/3) | Category: Correctness | Severity: Minor**

**What**: When a player undoes a move, the pending timeout for that move is not cancelled. If the timeout fires after the undo, the move gets re-added to `committedMoves`, causing the undone move to reappear.

**Where**: `client/web/src/stores/gameStore.ts:333`

**Suggested fix**: In `undoMove`, check `pendingMoveTimeouts` for the last move and cancel + remove the timeout if found. Use `toRaw()` on the popped move for consistent identity.

**Pros**: Fixes a user-visible bug where undo appears to not work.
**Cons**: Small change in one store function.

**Evaluator notes**: Unanimous 3/3/3.

---

### 21. PERF-003 / A* State Encoding String Concatenation
**Score: 3.0 (3/3/3) | Category: Performance | Severity: Critical**

**What**: `encodeState` uses `fmt.Sprintf` for every bot for every state in the A* hot loop, producing millions of string allocations. String concatenation with `+=` is O(n^2) per call.

**Where**: `solver/astar/astar.go:35`

**Suggested fix**: Replace string-based state encoding with a compact `uint64` key. With 4 bots on 16x16, each position needs 8 bits. A single `uint64` encodes all 4 bot positions. Use `map[uint64]bool` for the closed set.

**Pros**: Eliminates millions of allocations. Measurable solver speedup.
**Cons**: Moderate refactor of the state representation. Needs careful testing.

**Evaluator notes**: Unanimous 3/3/3. All agree it's a meaningful optimization.

---

### 22. PERF-004 / A* Wall Lookup Creates Defensive Copies
**Score: 3.0 (3/3/3) | Category: Performance | Severity: Moderate**

**What**: `hasWallBlocking` does a linear scan through all walls on every step of every slide in every state expansion. `Board.HWalls()` and `Board.VWalls()` create a new slice copy on every call. Millions of allocations in a typical A* search.

**Where**: `solver/astar/astar.go:227`

**Suggested fix**: Precompute wall data into a `[256]bool` array (indexed by `y*16 + x`) at the start of each solve. Pass the precomputed structure into `computeDestination`.

**Pros**: Eliminates per-call allocations and reduces wall lookup to O(1). Good companion to PERF-003.
**Cons**: Need to precompute separate arrays for horizontal/vertical walls in each direction.

**Evaluator notes**: Unanimous 3/3/3. Should be done alongside PERF-003 for a comprehensive solver optimization pass.

---

### 23. PERF-007 / Daily Puzzle Cache Grows Without Bound
**Score: 3.0 (3/3/3) | Category: Performance | Severity: Moderate**

**What**: The daily Manager's `cache` map stores every date's puzzles permanently. After months/years of operation, this accumulates entries that will never be accessed again.

**Where**: `server/daily/manager.go:26`

**Suggested fix**: Only cache the last 7 days. Evict entries older than that on each access or via a periodic goroutine.

**Pros**: Prevents slow memory growth. Simple eviction logic.
**Cons**: Trivial implementation. Cache misses for old dates are rare and can be served from disk.

**Evaluator notes**: Unanimous 3/3/3.

---

### 24. TEST-003 / ForceRemovePlayer Has No Unit Tests
**Score: 3.0 (3/3/3) | Category: Testing | Severity: Critical**

**What**: `ForceRemovePlayer` (used by BootPlayer and LeaveRoom) has zero direct tests. Unlike `RemovePlayer`, it removes connected players too. If this behavioral difference is accidentally lost in a refactor, both BootPlayer and LeaveRoom silently fail for connected players.

**Where**: `server/room/player_manager.go:197`

**Suggested fix**: Add tests for: force-remove connected player, force-remove disconnected player, cleanup of FinishedSolving/ReadyForNext/Solutions, signal emissions, no-op for nonexistent player.

**Pros**: Covers a critical path with known gaps (BUG-008 for pending players).
**Cons**: Moderate test writing effort.

**Evaluator notes**: Unanimous 3/3/3.

---

## Tier 4: Lower Priority (avg 2.0-2.67)

These findings are real but either low-impact, narrow edge cases, or would be fixed incidentally by higher-tier work.

| ID | Score | Category | Summary | Note |
|----|-------|----------|---------|------|
| BUG-005 | 2.67 | correctness | RemovePlayer emits conflicting signals | Narrow timing window |
| BUG-008 | 2.67 | correctness | ForceRemovePlayer ignores PendingPlayers | Fix alongside BUG-007 |
| BUG-010 | 2.67 | correctness | Settings validation gaps | Add when auth is built |
| BUG-013 | 2.67 | correctness | Shallow copy in callbacks shares pointers | Fix with Tier 3 lock batch |
| PERF-008 | 2.67 | performance | Progress cache unbounded | Fix alongside PERF-007 |
| TEST-002 | 2.67 | testing | LeaveRoom RPC handler untested | Fix alongside TEST-006 |
| TEST-007 | 2.67 | testing | ValidateSessionToken untested | Fix alongside BUG-004 |
| TEST-011 | 2.67 | testing | StartGame continuation path untested | Fix alongside BUG-001 |
| TEST-015 | 2.67 | testing | EndGame double-credit untested | Fix alongside BUG-001 |
| SEC-005 | 2.33 | security | X-Forwarded-For trust | Deployment concern, not code bug |
| SEC-009 | 2.33 | security | Session tokens in rooms.json | One-line `json:"-"` tag |
| BUG-015 | 2.33 | correctness | 30s wait for disconnected players | Annoying but not incorrect |
| PERF-005 | 2.33 | performance | A* stores full move copies | Moderate refactor, fix in solver pass |
| PERF-006 | 2.33 | performance | A* map allocation per state | Fix in solver pass |
| TEST-006 | 2.33 | testing | LeaveRoom service method untested | Overlap with TEST-002 |
| TEST-016 | 2.33 | testing | processSignals ordering untested | Hard to unit test |
| SEC-004 | 2.00 | security | StartGame no auth | Architectural gap, not targeted fix |
| SEC-011 | 2.00 | security | Timezone offset unbounded | Clamp to +/-840 if convenient |
| BUG-009 | 2.00 | correctness | EventType missing 'settings_changed' | Quick TypeScript fix |
| PERF-009 | 2.00 | performance | Puzzle generation CPU cost | Bounded by timeout already |
| PERF-011 | 2.00 | performance | Model wall lookup linear scan | Less hot than solver path |
| PERF-014 | 2.00 | performance | Move timeout Map key identity | Small fix, low frequency |
| TEST-012 | 2.00 | testing | Better-solution update untested | Simple comparison logic |
| TEST-021 | 2.00 | testing | encodeState collisions untested | Low probability issue |

---

## Tier 5: Skip or Defer (avg < 2.0)

Not worth addressing now. Either cosmetic, theoretical, acceptable by design, or already fine at current scale.

| ID | Score | Summary | Reason to skip |
|----|-------|---------|----------------|
| SEC-007 | 1.67 | WS token in query param | Standard pattern, ephemeral tokens |
| SEC-013 | 1.00 | Player IDs use math/rand | Not security-sensitive |
| SEC-014 | 1.00 | Room IDs short/guessable | By design for usability |
| SEC-015 | 1.33 | AllowSameHost default | Dev convenience, configure in prod |
| SEC-016 | 1.33 | Console logs session token | Dev only, ephemeral |
| BUG-012 | 1.67 | Cache TOCTOU race | Wastes CPU, no correctness impact |
| BUG-014 | 1.67 | GetWithLock vs Delete race | Extremely narrow window, benign |
| BUG-017 | 1.67 | Stale loadRoom responses | Self-corrects on next update |
| BUG-018 | 1.33 | Generic error on room not found | Minor UX polish |
| BUG-019 | 1.67 | StartNextGame no LastActivityAt | One-line fix if convenient |
| BUG-020 | 1.00 | Inconsistent math/rand | Cosmetic |
| PERF-012 | 1.67 | Full room serialization on save | Fine at current scale |
| PERF-013 | 1.00 | localStorage write per move | Negligible |
| PERF-015 | 1.00 | 256 grid cell divs | Trivial for browsers |
| PERF-016 | 1.00 | Bot collision O(B) with B=4 | Already efficient |
| PERF-017 | 1.00 | Player lookup linear scan | Faster than a map for 2-8 items |
| PERF-018 | 1.33 | AnimationService possibly dead code | Verify and remove if unused |
| TEST-008 | 1.67 | Single-player join rejection | Simple guard clause |
| TEST-009 | 1.67 | Solo disconnect grace period | Niche feature |
| TEST-013 | 1.33 | ComputeDestination error paths | Defensive checks, simple |
| TEST-014 | 1.00 | assignColorIndex untested | Trivial modular arithmetic |
| TEST-017 | 1.67 | StartJob replacement untested | Simple map overwrite |
| TEST-018 | 1.00 | SetSolverResult lazy init | Trivial pattern |
| TEST-019 | 1.00 | BOARD_SIZE only 16 tested | Fixed constant, unlikely to change |
| TEST-020 | 1.67 | Pending player in MarkFinished | Edge case |
| TEST-022 | 1.67 | Timezone edge cases untested | Low urgency for casual game |
| TEST-023 | 1.00 | Job ID duplicates untested | Trivial utility |
| TEST-024 | 1.67 | Cleanup completeness untested | Covered implicitly |
| TEST-025 | 1.00 | Pending player rejection | Guard clause |

---

## Suggested Implementation Order

Work is grouped into natural batches that can be tackled as separate PRs:

### Batch A: Critical Security + Daily Challenge Integrity
*Findings: SEC-001, SEC-003/TEST-005, SEC-010, SEC-012*
*Effort: Small (1-2 hours)*

Fix path traversal, replace `validateSolution` with `game.CheckSolution()`, add player name validation, add date validation. All are targeted fixes in `bouncebotserver.go` and `progress.go`.

### Batch B: WebSocket Hub Correctness
*Findings: BUG-002/PERF-010, BUG-003, SEC-008*
*Effort: Medium (2-4 hours)*

Fix Broadcast race (snapshot copy pattern), fix reconnection race (track active client per player), add read limits and deadlines. All in `server/ws/hub.go`.

### Batch C: Game State Correctness
*Findings: BUG-001, BUG-007/BUG-008, BUG-016*
*Effort: Medium (2-3 hours)*

Fix double-win-counting, extend disconnect/remove to handle PendingPlayers, fix undo timeout leak. Tests for all three.

### Batch D: Lock Discipline
*Findings: BUG-004, BUG-006, BUG-021, BUG-022, BUG-013*
*Effort: Medium-Large (3-5 hours)*

Systematic fix for the lock-release-before-use pattern across service.go and hub.go. Return copies instead of raw pointers. Move validation inside lock scope.

### Batch E: Solver Performance
*Findings: PERF-001, PERF-003, PERF-004, TEST-010*
*Effort: Medium-Large (3-5 hours)*

Replace ReorderSolution with greedy algorithm, switch A* state encoding to uint64, precompute wall lookup tables. Performance tests to validate.

### Batch F: Daily Package Tests + Caches
*Findings: TEST-001, PERF-002, PERF-007, PERF-008, BUG-011, SEC-002, SEC-006*
*Effort: Medium (3-4 hours)*

Write daily package tests, add cache eviction to Manager and ProgressManager, fix concurrent write race, add rate limiting to CreateRoom/SubmitDailySolution. Consider a lightweight daily auth token.

### Batch G: Physics Test Expansion
*Findings: TEST-004, TEST-003*
*Effort: Small-Medium (1-2 hours)*

Expand `physics_cases.json` with edge case boards. Add ForceRemovePlayer unit tests.
