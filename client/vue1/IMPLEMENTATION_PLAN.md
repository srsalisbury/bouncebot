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

### Step 21: Render robot history
**Goal:** Make it easier for the user to understand each robot's past movement.

**Tasks:**
- Leave a robot-colored dot behind showing each robot's starting position and a smaller dot at each robot's move-end position.
- Make sure to remove the dots on undo.

**Visible result:** Dots are left in robot's path.

---

### Step 22: Distinguish walls adjacent to target
**Goal:** Cleaner visuals.

**Tasks:**
- Figure out a way to better distinguish between the target rendering and the walls adjacent to it. Currently they both blend into each other when the colors are both dark.

**Visible result:** Visually distinct target box and walls.

---

### Step 23: Error Handling
**Goal:** Graceful handling of failures.

**Tasks:**
- Display server connection errors
- Add retry mechanism
- Handle edge cases

**Visible result:** Clear error messages when things go wrong

---

## Future Steps

### Step 24: How to Play Popup
**Goal:** Help new players understand the game.

**Tasks:**
- Add "How to Play" button or icon
- Create modal/popup explaining game rules
- Cover: goal, robot movement physics, controls
- Dismissable with click outside or close button

**Visible result:** New players can learn how to play

---

### Step 25: Keyboard Hints Toggle
**Goal:** Reduce visual clutter for experienced players.

**Tasks:**
- Add toggle button or key to show/hide keyboard hints
- Remember preference in localStorage
- Default to showing hints

**Visible result:** Cleaner UI for experienced players

---

### Step 26: Multiple Solutions Tracking
**Goal:** Track and compare different solution attempts.

**Tasks:**
- Store multiple solutions per puzzle
- Allow switching between saved solutions
- Show move count comparison
- Clear solutions on new puzzle

**Visible result:** Can save and compare different approaches

---

### Step 27: Update Solution Checking
**Goal:** Improve how solutions are validated.

**Tasks:**
- Review current CheckSolution RPC usage
- Determine improvements needed (TBD based on requirements)

**Visible result:** More robust solution validation

---

### Step 28: Documentation
**Goal:** Help developers build and run the project.

**Tasks:**
- Document Go server build and run instructions
- Document Vue client setup and dev server
- Document proto compilation steps
- Add to root README.md

**Visible result:** Clear setup instructions for contributors

---

### Step 29: Multiplayer Support
**Goal:** Allow multiple players to compete on the same puzzle.

**Visible result:** Competitive multiplayer experience

---

#### Step 29.1: Session Model & API
**Goal:** Server can create and manage game sessions.

**Tasks:**
- Add Session message to proto (id, players, created_at, current_game)
- Add Player message (id, name)
- Add CreateSession RPC (player_name) → returns session (no game yet)
- Add JoinSession RPC (session_id, player_name) → returns session
- Add GetSession RPC (session_id) → returns current session state
- Add StartGame RPC (session_id) → generates puzzle, returns session with game
- Server: Add session.go model with in-memory session store
- Server: Implement session RPCs
- Tests for session creation, joining, starting games

**Visible result:** Can create session, have players join, then start a game via API

---

#### Step 29.2: Client Session Flow
**Goal:** Players can create and join game sessions.

**Tasks:**
- Generate TypeScript types from updated proto
- Add vue-router for navigation
- Create HomeView: "Create Session" button, "Join Session" input
- Create SessionView: waiting room before game starts
- Create GameView: existing game board (session-aware)
- Store session state in Pinia (sessionId, players, currentPlayer, currentGame)
- Routes: / → HomeView, /session/:sessionId → SessionView or GameView
- Create session on "Create", redirect to /session/:sessionId
- Join session on "Join" or direct URL access
- Display shareable link for session
- "Start Game" button generates puzzle for everyone

**Visible result:** Two browsers can join same session, then start a game together

---

#### Step 29.3: Player Display
**Goal:** See who's in the session.

**Tasks:**
- Add players panel showing connected players
- Poll GetSession every 3s to refresh player list
- Show player names with colored indicators
- Highlight current player
- Show "Waiting for players..." if alone

**Visible result:** See other players join in real-time (via polling)

---

#### Step 29.4: WebSocket Infrastructure
**Goal:** Real-time updates without polling.

**Tasks:**
- Server: Add /ws endpoint using gorilla/websocket
- Server: Session broadcasts events to connected clients
- Server: Event types: player_joined, player_left, player_solved, game_started
- Client: Connect to WebSocket on session join
- Client: Handle incoming events, update store
- Client: Reconnect on disconnect
- Replace polling with WebSocket events
- Tests for WebSocket connection and event handling

**Visible result:** Instant player join/leave updates

---

#### Step 29.5: Solution Broadcasting
**Goal:** Players see when others solve the puzzle.

**Tasks:**
- Client: Send solve event to server when puzzle solved
- Server: Broadcast player_solved event (player_id, move_count)
- Client: Show notification "Player X solved in N moves!"
- Add solved players list to session state
- Display each player's best solution count
- Sort players by solution (best first, unsolved last)

**Visible result:** See "Alice solved in 5 moves!" when another player solves

---

#### Step 29.6: Game Timer
**Goal:** Timed competitive play.

**Tasks:**
- Server: Add started_at timestamp to current game
- Client: Display elapsed time since game started
- Show timer prominently in UI
- Timer starts when StartGame is called

**Visible result:** Timer visible to all players during game

---

#### Step 29.7: Scoring & Results
**Goal:** Determine winner and show results after each game.

**Tasks:**
- Define scoring: moves (primary), time to solve (secondary)
- Server: Track solve time for each player per game
- Client: Show game results when all players solve (or timeout)
- Display winner announcement for current game
- Track cumulative scores across games in session
- "Next Game" button starts new puzzle for same session
- Show session leaderboard (total wins/points)

**Visible result:** Results screen after each game, cumulative session scores

---

#### Step 30: Share Game Configuration
**Goal:** Share interesting puzzles with others.

**Tasks:**
- Design compact encoding for game state (walls, robots, target)
- Server/model: Encode game to short URL-safe string
- Server/model: Decode string back to game state
- Client: Add "Copy Puzzle Code" button to generate share code
- Client: Add "Import Puzzle" input/button to load a shared puzzle
- Validate imported puzzles before applying

**Visible result:** Copy a short code like "ABC123..." and share it; others can paste to play same puzzle

---

#### Step 31: Track available robot target locations
**Goal:** Panels and Boards know where it's okay to place a target location.

**Tasks:**
- Add possibleTargets Position to board/Board.
- Update parse board to handle possible targets
- Render should not render possible targets
- Rotate panel should rotate possible targets
- Add tests

**Visible result:** Nothing.

---

#### Step 32: Support dynamic generation of game boards.
**Goal:** New game can result in lots of possible configurations.

**Tasks:**
- A new game be formed by random panels, random robots, and random target.
- Panels should be some permutation of 1-4 for now
- Target can be on any possibleTarget location with a random robot chosen.
- Robots can be placed anywhere except on each other, on the target, or in the middle four cells.
- We should still be able to generate the current fixed configuration for now.
- Add tests

**Visible result:** New session generates a random game configuration.

---

## Reference

The implementation in `/Users/mike/dev/bouncebot/cmd/vue1/` can be used as reference for:
- Connect client setup (`src/services/connectClient.ts`)
- Game logic patterns (`src/stores/gameStore.ts`)
- Board rendering approach (`src/components/GameBoard.vue`)
- TypeScript types (`src/types/game.ts`)
