# BounceBot Vue3 Client - Implementation Plan

Stepwise plan for building a Vue3 web client from scratch.
Each step is a single PR-sized change with visible progress in the browser.

## Overview

**Goal:** Build a Vue3 client that allows users to:
- View a 16x16 game grid with walls, robots, and a target
- Move robots to solve the puzzle (get target robot to target cell)
- See the minimum number of moves required

**Tech Stack:**
- Vue 3 (Composition API)
- TypeScript
- Vite (build tool)
- Connect protocol (browser-friendly RPC to Go backend)

---

## Phase 1: Minimal Visible App

### Step 1: Initialize Vue3 Project
**Goal:** Empty Vue app running in browser.

**Tasks:**
- Create Vue3 + TypeScript + Vite project
- Verify `npm run dev` shows default page
- Clean up boilerplate (remove default components)

**Visible result:** Browser shows "BounceBot" heading on blank page

---

### Step 2: Static 16x16 Grid
**Goal:** Display a basic game board grid.

**Tasks:**
- Create GameBoard component
- Render 16x16 grid of cells with CSS
- Add border to represent board edges

**Visible result:** Browser shows empty 16x16 grid

---

### Step 3: Add Hardcoded Robots
**Goal:** Display colored robot markers on the grid.

**Tasks:**
- Define robot positions in component
- Render 4 colored circles at positions
- Style robots distinctly (red, blue, green, orange)

**Visible result:** Grid shows 4 colored robots at fixed positions

---

### Step 4: Add Hardcoded Walls
**Goal:** Display internal walls on the grid.

**Tasks:**
- Define wall positions (horizontal and vertical)
- Render wall segments between cells
- Style walls visibly (yellow/gold bars)

**Visible result:** Grid shows robots and internal walls

---

### Step 5: Add Target Marker
**Goal:** Show where the target robot needs to go.

**Tasks:**
- Define target position and which robot is the target
- Render target marker (dashed border, matching robot color)
- Add pulsing animation to draw attention

**Visible result:** Grid shows target location for one of the robots

---

## Phase 2: Interactivity

### Step 6: Robot Selection
**Goal:** Player can click to select a robot.

**Tasks:**
- Add click handler to robots
- Track selected robot in state
- Highlight selected robot visually

**Visible result:** Clicking a robot shows selection highlight

---

### Step 7: Keyboard Movement (No Physics)
**Goal:** Move selected robot with arrow keys (teleport for now).

**Tasks:**
- Add keyboard event listener
- Move selected robot one cell in direction
- Prevent moving off board edges

**Visible result:** Arrow keys move selected robot one cell at a time

---

### Step 8: Sliding Movement Physics
**Goal:** Robots slide until hitting obstacles (Ricochet Robots rules).

**Tasks:**
- Implement slide logic: move until hitting edge, wall, or robot
- Robot stops at the obstacle, not inside it
- Update movement to use sliding

**Visible result:** Robots slide across board until blocked

---

### Step 9: Move Counter & History
**Goal:** Track and display moves with full history.

**Tasks:**
- Track each move with robot ID, direction, and color
- Display move count above history panel
- Display move history list to the right of the board
- Each move shown as colored robot dot with direction arrow

**Visible result:** Move panel shows count and scrollable history of all moves

---

## Phase 3: State Management

### Step 10: Pinia Store Setup
**Goal:** Centralize game state management.

**Tasks:**
- Install and configure Pinia
- Move game state to store (board, robots, target, moves)
- Refactor component to use store

**Visible result:** Same functionality, cleaner code structure

---

### Step 11: Win Detection
**Goal:** Detect and celebrate when puzzle is solved.

**Tasks:**
- Check if target robot is at target position after each move
- Show win message/dialog
- Display final move count

**Visible result:** Solving puzzle shows "You won in X moves!"

---

### Step 12: Undo Move
**Goal:** Player can undo their last move.

**Tasks:**
- Store position history for each move
- Add Undo button
- Restore previous state on undo

**Visible result:** Undo button reverses last move

---

### Step 13: Reset Puzzle
**Goal:** Restart current puzzle from beginning.

**Tasks:**
- Store initial game state
- Add Reset button
- Restore to initial state, clear moves

**Visible result:** Reset button returns to starting positions

---

## Phase 4: Server Integration

### Step 14: Connect Server (Go Backend)
**Goal:** Go server speaks Connect protocol for browser access.

**Tasks:** (Discuss implementation approach)
- Add Connect handler to Go server
- Add CORS support
- Serve on HTTP port (e.g., 8080)

**Visible result:** Server accepts HTTP requests from browser

---

### Step 15: Proto/Connect Client Setup
**Goal:** Vue client can call server RPCs.

**Tasks:**
- Add Connect and Buf dependencies
- Generate TypeScript from proto
- Create Connect client service

**Visible result:** Client code can make RPC calls

---

### Step 16: Fetch Game from Server
**Goal:** Load game data from server instead of hardcoded.

**Tasks:**
- Call MakeGame RPC on app load
- Parse response into game state
- Handle loading and error states

**Visible result:** Game board populated from server response

---

### Step 17: New Game Button
**Goal:** Player can request a new puzzle.

**Tasks:**
- Add "New Game" button
- Call MakeGame RPC on click
- Reset state with new game data

**Visible result:** Button fetches fresh puzzle from server

---

### Step 18: Server Solution Validation
**Goal:** Validate solution with server.

**Tasks:**
- Add "Check Solution" button (optional - can auto-check on win)
- Call CheckSolution RPC
- Display server validation result

**Visible result:** Server confirms if solution is valid

---

## Phase 5: Polish

### Step 19: Move Animations
**Goal:** Smooth robot movement instead of teleporting.

**Tasks:**
- Add CSS transitions for position changes
- Animate slides over short duration
- Maintain game feel (quick but visible)

**Visible result:** Robots animate when moving

---

### Step 20: Better Styling
**Goal:** Professional appearance.

**Tasks:**
- Add Vuetify or custom styling
- Improve layout and spacing
- Add responsive design

**Visible result:** Polished, attractive UI

---

### Step 21: Error Handling
**Goal:** Graceful handling of failures.

**Tasks:**
- Display server connection errors
- Add retry mechanism
- Handle edge cases

**Visible result:** Clear error messages when things go wrong

---

## Future Phases (Deferred)

### Multiplayer Competition
- Timer and scoring
- Real-time competition rooms
- Leaderboards

### Enhanced Features
- Puzzle difficulty selection
- Optimal solution hints

---

## Recommended Implementation Order

For fastest visible progress:
1. Steps 1-5 (static display) - See the game board
2. Steps 6-9 (interaction) - Play the game locally
3. Steps 11-13 (game features) - Complete single-player experience
4. Steps 10 (Pinia) - Code quality improvement
5. Steps 14-18 (server) - Real backend integration
6. Steps 19-21 (polish) - Production quality

---

## Reference

The implementation in `/Users/mike/dev/bouncebot/cmd/vue1/` can be used as reference for:
- Connect client setup (`src/services/connectClient.ts`)
- Game logic patterns (`src/stores/gameStore.ts`)
- Board rendering approach (`src/components/GameBoard.vue`)
- TypeScript types (`src/types/game.ts`)
