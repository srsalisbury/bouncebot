# Smart Solution Time Formatting Plan

## Goal
Improve the display of solution times in the game UI. Currently, times are shown as `m:ss`. When multiple players have similar times (e.g., `1:05.12` and `1:05.89`), they both appear as `1:05`, creating ambiguity about who was faster. The goal is to show tenths (`1:05.1`) or hundredths (`1:05.12`) **only** when necessary to distinguish between tied times.

## Implementation Steps

### Step 1: Create Shared Time Utility
Create `client/web/src/services/timeUtils.ts` with the following functions:

1.  **`formatDuration(seconds: number, precision: number = 0): string`**
    *   Formats a duration in seconds to a string.
    *   `precision = 0`: `m:ss` (e.g., `1:05`)
    *   `precision = 1`: `m:ss.f` (e.g., `1:05.1`)
    *   `precision = 2`: `m:ss.ff` (e.g., `1:05.12`)

2.  **`calculateDurationSeconds(start: Timestamp, end: Timestamp): number`**
    *   Helper to compute the difference between two Protobuf Timestamps in seconds (float).

3.  **`getFormattedTimes(items: {id: string, seconds: number}[]): Map<string, string>`**
    *   Takes a list of items with IDs and raw durations.
    *   Returns a Map where keys are IDs and values are the formatted time strings.
    *   **Logic:**
        *   Group all items by their formatted string at `precision = 0`.
        *   For groups with > 1 item (collisions), re-format those items at `precision = 1`.
        *   If collisions persist, re-format those specific items at `precision = 2`.
        *   "Perfect ties" (identical times even at precision 2) will naturally share the same formatted string.

### Step 2: Refactor Components
Update the following components to use the new service:

*   **`client/web/src/components/PlayersPanel.vue`**
*   **`client/web/src/components/GameBoardPlayerSolutions.vue`**
*   **`client/web/src/components/PlayerSolutionsDrawer.vue`**

**Refactoring Logic for Each Component:**
1.  Import `getFormattedTimes` and `calculateDurationSeconds` from `timeUtils.ts`.
2.  Create a computed property (e.g., `solutionTimes`) that maps the relevant solutions (players, solvers, top 3, etc.) to a simple list of objects: `{ id: string, seconds: number }`.
3.  Create a computed property `formattedTimesMap` that calls `getFormattedTimes(solutionTimes)`.
4.  Update the template to display the time by looking it up in the map: `formattedTimesMap.get(id)`.

## Verification

1.  **Unique Times:** Verify that times with no close competitors display as `m:ss`.
2.  **Tenths Resolution:** Verify that times colliding at `m:ss` but distinct at `m:ss.f` display with tenths (e.g., `1:05.1` vs `1:05.2`).
3.  **Hundredths Resolution:** Verify that times colliding at `m:ss.f` but distinct at `m:ss.ff` display with hundredths (e.g., `1:05.12` vs `1:05.13`).
4.  **Perfect Ties:** Verify that times that are effectively identical (to the hundredth) display the same string (e.g., both `1:05.12`) and look consistent.
