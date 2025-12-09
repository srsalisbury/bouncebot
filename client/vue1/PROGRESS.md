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

### Step 14: Connect Server (Go Backend)
**Status:** Complete

**What was done:**
- Fixed proto go_package to use correct module path (github.com not github.org)
- Updated compile_protos.sh to generate Connect code
- Generated Connect handler in proto/protoconnect/
- Added Connect and CORS dependencies to go.mod
- Rewrote server/main.go to use Connect instead of gRPC
- Added CORS support for browser access from localhost:5173
- Changed default port from 50055 to 8080
- Added h2c (HTTP/2 cleartext) support so gRPC clients still work
- Updated cmd/clienttest to use new default port

**Files added:**
- `proto/protoconnect/bouncebot.connect.go` - Generated Connect handlers

**Files modified:**
- `proto/bouncebot.proto` - Fixed go_package path
- `proto/compile_protos.sh` - Added Connect code generation
- `server/main.go` - Rewrote for Connect with CORS and h2c
- `go.mod` - Added connectrpc.com/connect, github.com/rs/cors, h2c
- `cmd/clienttest/main.go` - Updated default port to 8080

---

### Step 15: Proto/Connect Client Setup
**Status:** Complete

**What was done:**
- Installed @connectrpc/connect, @connectrpc/connect-web, @bufbuild/protobuf
- Installed @bufbuild/buf, @bufbuild/protoc-gen-es as dev dependencies
- Created buf.gen.yaml for TypeScript code generation
- Generated TypeScript types from proto in src/gen/bouncebot_pb.ts
- Created Connect client service in src/services/connectClient.ts
- Added npm run generate script

**Files added:**
- `buf.gen.yaml` - Buf code generation config
- `src/gen/bouncebot_pb.ts` - Generated TypeScript types
- `src/services/connectClient.ts` - Connect client setup

**Files modified:**
- `package.json` - Added dependencies and generate script

---

### Step 16: Fetch Game from Server
**Status:** Complete

**What was done:**
- Added loadGame() action to store that calls MakeGame RPC
- Added applyGame() helper to parse server response into game state
- Added isLoading and error state for loading/error handling
- Removed hardcoded robot, wall, and target data
- Call loadGame() on component mount
- Display loading message while fetching
- Display error message with retry button on failure

**Files modified:**
- `src/stores/gameStore.ts` - Added loadGame, applyGame, isLoading, error
- `src/components/GameBoard.vue` - Added loading/error UI, call loadGame on mount

---

### Step 17: New Game Button
**Status:** Complete

**What was done:**
- Added "New Game" button to move panel
- Button calls loadGame() to fetch a fresh puzzle from server

**Files modified:**
- `src/components/GameBoard.vue` - Added New Game button with styling

---

### Step 18: Server Solution Validation
**Status:** Complete

**What was done:**
- Added toX/toY to Move type to track destination positions
- Store initial Game object from server for validation requests
- Added checkSolution() action that calls CheckSolution RPC
- Added isValidating and validationResult state
- Auto-validate with server when puzzle is solved (via watcher)
- Display validation status: "Validating...", success message, or error

**Files modified:**
- `src/stores/gameStore.ts` - Added checkSolution, validation state, toX/toY tracking
- `src/components/GameBoard.vue` - Added watcher for auto-validation, validation UI

---

### Step 19: Move Animations
**Status:** Complete

**What was done:**
- Added CSS transition for left/top properties on robots
- Robots now animate smoothly when sliding (150ms ease-out)

**Files modified:**
- `src/components/GameBoard.vue` - Added position transitions to robot class

---

### Step 20: Better Styling
**Status:** Complete

**What was done:**
- Styled "New Game" button with green theme, hover/active states
- Improved move panel layout with consistent spacing
- Styled error retry button to match theme
- Added keyboard hints below board showing controls
- Restructured layout with game-content wrapper for better organization
- Improved App.vue title spacing

**Files modified:**
- `src/App.vue` - Improved layout spacing
- `src/components/GameBoard.vue` - Button styles, keyboard hints, layout restructure

---

### Step 21: Render Robot History
**Status:** Complete

**What was done:**
- Made initialRobots reactive and exposed from store
- Added larger dots at each robot's starting position
- Added smaller dots at each move destination (using committedMoves)
- Dots use robot colors with 80% opacity
- Proper z-index ordering (robots above dots)
- Added committedMoves array with 150ms delay to sync dots with animation
- Dots cleared on undo/reset

**Files modified:**
- `src/stores/gameStore.ts` - Made initialRobots reactive, added committedMoves with delay
- `src/components/GameBoard.vue` - Added history dot rendering and styling

---

### Step 22: Distinguish Walls Adjacent to Target
**Status:** Complete

**What was done:**
- Changed wall color from brown (#8b4513) to dark gray (#2a2a2a)
- Dark gray contrasts better with all robot colors

**Files modified:**
- `src/components/GameBoard.vue` - Changed WALL_COLOR constant

---

## In Progress

_None currently_

---

## Up Next

- Step 23: Error Handling
