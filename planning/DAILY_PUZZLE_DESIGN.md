# Daily Puzzle System Design

## Overview
This feature introduces a "Daily Challenge" mode where users can play three curated puzzles (Easy, Medium, Hard) every day. The puzzles reset at local midnight based on the user's timezone.

## Goals
1.  **Engagement**: Give users a reason to return daily.
2.  **Quality Control**: Puzzles must be guaranteed to match their difficulty label (verified by the A* solver).
3.  **Progression**: Track daily completion status per user (device-based).
4.  **Future-Proofing**: Store puzzle history to allow for a future "Archives" feature.

## Architecture

### 1. The Puzzle Generator (Background Worker)
Since we need specific difficulty guarantees, we cannot generate puzzles on-the-fly when a user requests them. Instead, the server will maintain a buffer of future puzzles.

**Configuration:**
- **Buffer size**: 7 days of future puzzles
- **Generation timeout**: 10 minutes per puzzle (fail if not found in time)
- **Check frequency**: Once per day (or on server startup)

**Logic:**
A background goroutine runs daily (and on startup) to ensure puzzles exist for the next 7 days.

1.  For each date missing any difficulty:
    a.  Generate a random board/game state using a fresh random seed.
    b.  Run the **A* Solver** to find the optimal solution length ($L$).
    c.  Classify the puzzle:
        *   **Easy**: $L < 8$
        *   **Medium**: $8 \le L < 12$
        *   **Hard**: $L \ge 12$
    d.  If the puzzle fits a needed difficulty bucket, save it. Otherwise, discard and retry.
2.  If a puzzle cannot be generated within the timeout, log an error and notify (allows human intervention or retry on next cycle).

**Avoiding repeats**: Each puzzle is generated from a fresh random seed. Given the enormous state space (board configurations × robot placements × target selection), the probability of generating a duplicate is negligible without explicit checking.

### 2. Data Storage
We use a file-per-entity approach with directory hierarchies that keep directories small and browsable.

#### A. Daily Puzzles: `data/daily_puzzles/{year}/{month}/{day}.json`

Each day's puzzles are stored in a separate file, organized by year and month:

```
data/
  daily_puzzles/
    2026/
      01/
        22.json
        23.json
      02/
        01.json
```

Each file uses the human-readable ASCII art format (same as `solver/benchmark/puzzles/`):

```json
{
  "date": "2026-01-22",
  "easy": {
    "optimal_moves": 5,
    "solution": ["B0:up", "B1:left", "B0:right", "B1:down", "B0:down"],
    "game": [
      "+----+----+----+----+----+----+----+----+----+----+----+----+----+----+----+----+",
      "|                   |                                  |                        |",
      "+    +    +    +    +    +    +    +    +    +    +    +    +    +    +    +    +",
      "|    B0        |              T0                            |              B1   |",
      "..."
    ]
  },
  "medium": {
    "optimal_moves": 11,
    "solution": [ "..." ],
    "game": [ "..." ]
  },
  "hard": {
    "optimal_moves": 17,
    "solution": [ "..." ],
    "game": [ "..." ]
  }
}
```

The ASCII format encodes walls, bot positions (`B0`-`B3`), and targets (`T0`-`T3`) in a visually inspectable way. This ensures puzzles remain human-readable and that changes to generation code don't break historical puzzles.

**Directory bounds**: Max 12 subdirs per year, max 31 files per month.

#### B. User Progress: `data/users/{prefix}/{uuid}.json`

Each user has a single file containing all their progress. Files are sharded into subdirectories by the first 2 hex characters of the UUID (256 possible buckets) to keep directories small.

```
data/
  users/
    a3/
      a3f82b91-1234-5678-9abc-def012345678.json
    f7/
      f7c91a02-abcd-ef01-2345-6789abcdef01.json
```

Each user file tracks completion by date:

```json
{
  "2026-01-22": {"easy": true, "medium": true, "hard": false},
  "2026-01-23": {"easy": true, "medium": false, "hard": false}
}
```

**Why this structure**:
- One file per user = atomic reads/writes with no cross-file coordination
- Sharding prevents any directory from exceeding ~1/256th of total users
- A user playing daily for 10 years ≈ 3,650 entries — still a tiny JSON file
```

### 3. API Changes (`proto/bouncebot.proto`)

We need new RPCs to handle the daily context.

```protobuf
service BounceBot {
  // ... existing RPCs ...

  // Fetch today's puzzles and the user's status
  rpc GetDailyChallenge(GetDailyChallengeRequest) returns (GetDailyChallengeResponse);

  // Submit a solution specifically for a daily puzzle
  rpc SubmitDailySolution(SubmitDailySolutionRequest) returns (SubmitDailySolutionResponse);
}

message DailyPuzzleInfo {
  string difficulty = 1; // "easy", "medium", "hard"
  Game game = 2;         // The board and robot setup
  bool solved = 3;       // Has this user solved it?
  int32 optimal_moves = 4; // Only populated if solved == true
}

message GetDailyChallengeRequest {
  string player_id = 1;
  int32 timezone_offset_minutes = 2; // Minutes from UTC (e.g., -480 for PST). Server uses this to determine the user's local date.
}

message GetDailyChallengeResponse {
  string date = 1; // "2026-01-22" (user's local date)
  repeated DailyPuzzleInfo puzzles = 2;
  int32 seconds_until_reset = 3; // Seconds until user's local midnight
}

message SubmitDailySolutionRequest {
  string player_id = 1;
  string date = 2;
  string difficulty = 3;
  repeated BotPos moves = 4;
}

message SubmitDailySolutionResponse {
  bool correct = 1;
  bool new_completion = 2; // True if this was the first time solving it
}
```

### 4. Error Handling

**Missing or corrupted puzzle file** (when user requests a daily challenge):
- Log/notify error to system for developer attention
- Return an error response to the client
- Client shows an error dialog: "Today's puzzle is temporarily unavailable. We've noted the issue — please try again tomorrow."

**User submits solution for non-existent puzzle** (wrong date or difficulty):
- Log the error with details (player_id, date, difficulty) to detect patterns
- Return error to client: "We couldn't verify that solution. Please refresh and try again."

**Solution validation**:
- Server replays the submitted moves on the puzzle's initial state
- Valid if the target bot ends on the target cell after all moves
- A solution is "optimal" if its move count equals `optimal_moves` (any sequence of minimal length is equally good)

**Security (V1 limitation)**:
- `player_id` is client-provided from LocalStorage with no server-side authentication
- Users can spoof IDs or mark puzzles solved without actually solving
- Accepted for V1 — this is a single-player experience with no competitive element yet

### 5. Server Implementation Plan

1.  **`server/daily/manager.go`**:
    *   Manages the `data/daily_puzzles/` directory hierarchy.
    *   `GetPuzzlesForDate(date)`: Loads and parses `{year}/{month}/{day}.json`.
    *   `EnsureFuturePuzzles(days)`: The generation loop.

2.  **`server/daily/progress.go`**:
    *   Manages the `data/users/` directory hierarchy.
    *   `GetUserProgress(uuid)`: Loads `{prefix}/{uuid}.json`.
    *   `SaveUserProgress(uuid, progress)`: Atomic write to user file.
    *   Thread-safe updates for user status.

3.  **`server/main.go`**:
    *   Initialize `DailyManager`.
    *   Register the new RPC handlers.
    *   Start the background generation worker.

### 6. Frontend Implementation Plan

1.  **New Store `dailyStore.ts`**:
    *   Stores the current day's puzzles and completion status.
    *   Actions: `fetchDaily()`, `submitSolution()`.

2.  **UI Components**:
    *   **Home Screen Entry**: A prominent "Daily Challenge" card.
    *   **`DailyChallengeView.vue`**:
        *   Shows the 3 difficulty cards.
        *   Status indicators (Checkmark for solved).
        *   Countdown timer to next reset.
    *   **`DailyGameView.vue`**:
        *   A stripped-down version of `RoomView`.
        *   No multiplayer elements (no player list, no chat).
        *   "Submit" checks against the daily API validation.
        *   On success: Show "Puzzle Solved!" modal and return to Daily Menu.

## Future Considerations (Post-MVP)

*   **Streaks**: Detailed stats in user progress (current streak, max streak).
*   **Archives**: API endpoint `GetPuzzlesByDate` to play past days.
*   **Leaderboards**: "Fastest to solve all 3 today".
*   **Authentication**: Server-side player identity to prevent spoofing and enable cross-device progress sync.

## Work Breakdown

1.  **Backend Core**: Create the `daily` package and the puzzle generation logic with Solver integration.
2.  **Persistence**: Implement JSON loading/saving for puzzles and user progress.
3.  **API Wiring**: Update Proto and implement RPCs.
4.  **Frontend Logic**: Create the store and API client methods.
5.  **Frontend UI**: Create the views and hook up the game board.
