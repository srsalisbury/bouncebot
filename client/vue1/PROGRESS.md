# BounceBot Vue3 Client - Progress Tracker

Tracks completed steps from IMPLEMENTATION_PLAN.md.

## Completed Steps

### Step 1: Initialize Vue3 Project
**PR:** https://github.com/srsalisbury/bouncebot/pull/10
**Status:** Complete

**What was done:**
- Scaffolded Vue3 + TypeScript + Vite project in `client/vue1/`
- Removed boilerplate (HelloWorld component, default assets)
- Created minimal App.vue with "BounceBot" heading
- Simplified global styles (dark theme)

**Files added:**
- `index.html` - HTML entry point
- `src/main.ts` - Vue app initialization
- `src/App.vue` - Root component
- `src/style.css` - Global styles
- `vite.config.ts` - Vite configuration
- `tsconfig.json`, `tsconfig.app.json`, `tsconfig.node.json` - TypeScript config
- `package.json`, `package-lock.json` - Dependencies

---

### Step 2: Static 16x16 Grid
**PR:** https://github.com/srsalisbury/bouncebot/pull/12
**Status:** Complete

**What was done:**
- Created GameBoard component with CSS grid layout
- Rendered 16x16 cells with light gray background (#dddddd)
- Added border to represent board edges
- Imported GameBoard into App.vue

**Files added:**
- `src/components/GameBoard.vue` - Game board component

**Files modified:**
- `src/App.vue` - Import and render GameBoard

---

### Step 3: Add Hardcoded Robots
**Status:** Complete

**What was done:**
- Added 4 robots with hardcoded positions
- Styled robots as colored circles (red, blue, green, yellow)
- Numbered robots 1-4 for identification
- Factored out robot colors as named constants (ROBOT_COLORS)

**Files modified:**
- `src/components/GameBoard.vue` - Added robot rendering

---

### Step 4: Add Hardcoded Walls
**Status:** Complete

**What was done:**
- Added vertical and horizontal walls with hardcoded positions
- Styled walls as brown bars (WALL_COLOR constant)
- Made board border match wall color/thickness
- Made grid lines thinner (0.5px) to distinguish from walls

**Files modified:**
- `src/components/GameBoard.vue` - Added wall rendering

---

### Step 5: Add Target Marker
**Status:** Complete

**What was done:**
- Added target with hardcoded position and robot ID
- Styled as solid rounded rectangle with circular hole (robot-sized)
- Used CSS mask to create the hole effect
- Added black number in center matching target robot ID
- Target color matches the target robot's color

**Files modified:**
- `src/components/GameBoard.vue` - Added target rendering

---

### Step 6: Robot Selection
**Status:** Complete

**What was done:**
- Added click handler to select/deselect robots
- Track selected robot ID in reactive state
- Visual highlight: white border with black outline, scale effect
- Hover effect on robots
- Keyboard support: press 1-4 to select robots by number

**Files modified:**
- `src/components/GameBoard.vue` - Added selection state and interaction

---

### Step 7: Keyboard Movement (No Physics)
**Status:** Complete

**What was done:**
- Made robots array reactive for movement updates
- Added arrow key and WASD support for movement
- Move selected robot one cell in pressed direction
- Prevent moving off board edges
- preventDefault on arrow keys to avoid page scrolling

**Files modified:**
- `src/components/GameBoard.vue` - Added movement logic

---

### Step 8: Sliding Movement Physics
**Status:** Complete

**What was done:**
- Added hasWall() function to check walls and board edges
- Added isOccupied() function to check for other robots
- Added calculateDestination() to compute slide endpoint
- Robots now slide until hitting wall, edge, or another robot
- Replaced one-cell movement with sliding physics

**Files modified:**
- `src/components/GameBoard.vue` - Added sliding physics logic

---

### Step 9: Move Counter & History
**Status:** Complete

**What was done:**
- Added Move type tracking robotId, direction, and color
- Added moves ref array to store move history
- Created move panel displayed to the right of the board
- Shows move count at top ("Moves: N")
- Lists each move as colored robot dot with direction arrow
- Added DIRECTION_ARROWS constant for arrow symbols (↑↓←→)
- Used flexbox layout for game-container to position board and panel side by side

**Files modified:**
- `src/components/GameBoard.vue` - Added move tracking and history panel

---

### Step 10: Pinia Store Setup
**Status:** Complete

**What was done:**
- Installed and configured Pinia for state management
- Created `src/stores/gameStore.ts` with centralized game state
- Moved all game state to store: robots, walls, target, moves, selectedRobotId
- Moved game logic to store: selectRobot, moveRobot, hasWall, isOccupied, calculateDestination
- Exported reusable types: Direction, Robot, Wall, Target, Move
- Refactored GameBoard.vue to use store instead of local state

**Files added:**
- `src/stores/gameStore.ts` - Pinia store for game state

**Files modified:**
- `src/main.ts` - Configure Pinia
- `src/components/GameBoard.vue` - Use store instead of local state
- `package.json` - Added pinia dependency

---

### Refactor: Implicit Robot Colors
**Status:** Complete

**What was done:**
- Changed ROBOT_COLORS from named object to array of 10 colors
- Added getRobotColor(robotId) function to derive color from ID
- Removed color field from Robot and Move types
- Removed targetRobot computed property (no longer needed)
- Updated GameBoard.vue to use getRobotColor() everywhere

**Files modified:**
- `src/stores/gameStore.ts` - Color palette and helper function
- `src/components/GameBoard.vue` - Use getRobotColor()

---

### Step 11: Solved Detection
**Status:** Complete

**What was done:**
- Added isSolved computed property to check if target robot is at target position
- Display "Solved" label in green next to move count when puzzle is solved
- Prevent further moves after puzzle is solved
- Added walls at (0,7) and (7,7) to make test puzzle solvable

**Files modified:**
- `src/stores/gameStore.ts` - Added isSolved computed, block moves when solved, added walls
- `src/components/GameBoard.vue` - Added solved label display

---

### Step 12: Undo Move
**Status:** Complete

**What was done:**
- Added fromX/fromY to Move type to track previous position
- Added undoMove() action to restore robot to previous position
- Keyboard shortcuts: z, u, or Escape to undo
- Auto-select the robot that was just undone for easy re-move

**Files modified:**
- `src/stores/gameStore.ts` - Added position tracking and undoMove action
- `src/components/GameBoard.vue` - Added undo keyboard shortcuts

---

### Step 13: Reset Puzzle
**Status:** Complete

**What was done:**
- Store initial robot positions for reset
- Added resetPuzzle() action to restore robots, clear moves, deselect
- Keyboard shortcut: Shift+R to reset (uppercase to prevent accidental reset)

**Files modified:**
- `src/stores/gameStore.ts` - Added initialRobots and resetPuzzle action
- `src/components/GameBoard.vue` - Added reset keyboard shortcut

---

## In Progress

_None currently_

---

## Up Next

- Step 14: Connect Server (Go Backend)
