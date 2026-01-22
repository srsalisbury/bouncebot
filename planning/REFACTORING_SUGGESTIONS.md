# Code Refactoring Suggestions

Based on a review of the codebase (specifically `client/vue1` and `server/room`), the following refactorings are suggested to improve maintainability, reduce duplication, and enhance consistency.

## Client-Side (Vue)

### 1. Extract Solution Display Logic into a Composable
**Problem:** `GameBoardPlayerSolutions.vue` and `PlayerSolutionsDrawer.vue` share significant identical logic for managing solution sorting, indexing, and display visibility.
**Duplicated Logic:**
*   `isCurrentPlayerInTopThree`: Boolean check.
*   `solverStartIndex` & `currentPlayerStartIndex`: Index calculations.
*   `hasTopThreeSolutions` & `showCurrentPlayerSolution`: Visibility checks.
*   `formatSolveTime`: Time formatting helper.

**Suggestion:** Create `client/vue1/src/composables/useSolutionDisplay.ts` to encapsulate these computed properties and helpers.

### 2. Centralize Move Type Definitions
**Problem:** Multiple components define the same local interface:
```typescript
interface MoveWithDirection {
  robotId: number
  direction: Direction
}
```
**Suggestion:** Move this to `client/vue1/src/constants.ts` or a new `types/game.ts` file.

### 3. Unified "Solution List" Data Structure
**Problem:** Components receive `topThreeSolutions`, `solverSolutions`, `playerSolutions`, etc., separately and reconstruct the display order internally. This leads to complex template logic with multiple `v-if` blocks for dividers and ordering.
**Suggestion:** Create a selector in `gameStore` or a computed property in the parent view that returns a single, flat list of "DisplayableSolutions" decorated with their type (`top`, `solver`, `current`, `other`). Components can then iterate over this single list.

### 4. CSS Variable Standardization
**Problem:** Components use hardcoded hex values (e.g., `#43a047`, `#2a2a2a`) and duplicate media queries for dark/light mode.
**Suggestion:** Extract common colors into CSS variables in `client/vue1/src/style.css`. Define the theme once and apply it across components.

## Server-Side (Go)

### 5. Room & Game Lifecycle Separation
**Observation:** The `Room` struct in `server/room/room.go` contains some helper methods (e.g., `ClearGameState`, `FindPlayerIndex`).
**Suggestion:** Ensure strict separation of concerns. `Room` should ideally be a pure data structure (anemic model). Logic for state mutation or complex lookups could be moved entirely into the respective managers (`GameLifecycleManager`, `PlayerManager`) to improve testability and keep the data model clean.
