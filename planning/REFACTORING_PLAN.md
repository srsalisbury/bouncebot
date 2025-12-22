# Gemini CLI: Ongoing Refactoring Plan

This document outlines the current refactoring tasks being undertaken by the Gemini CLI agent. The agent will be working through these items sequentially, updating the status as progress is made.

## Backend Refactoring (Go)

- [ ] **Address Critical Bugs and Configuration:**
    - [x] Implement Player Disconnection Events with Grace Period:
        - [x] **Backend:**
            - [x] Add a `PlayerID` to the `Client` struct in `server/ws/hub.go` to track players.
            - [x] Add a `Status` field (e.g., `Connected`, `Disconnected`) to the `Player` struct in `server/session/session.go`.
            - [x] When a player disconnects, set their `Status` to `Disconnected` and start a grace period timer (e.g., 30 seconds).
            - [x] On reconnection, if a player with the same ID exists and is `Disconnected`, update their `Status` to `Connected` and cancel the timer.
            - [x] If the grace period timer expires, remove the player from the session and broadcast a "player left" event to all clients.
        - [x] **Frontend:**
            - [x] Handle the "player left" event in the websocket service.
            - [x] Update the Pinia store to remove the player from the game state.
            - [x] Ensure the UI correctly reflects the player's removal.
    - [x] Externalize Configuration:
        - [x] Scan the Go codebase (starting with `server/main.go`) to identify all hardcoded configuration values (e.g., CORS policy, server port).
        - [x] Choose and implement a configuration library (e.g., `godotenv`) to load settings from a `.env` file.
        - [x] Create a `Config` struct in the `server` directory to hold all application configuration in a type-safe manner.
        - [x] Implement logic in `server/main.go` to load configuration from the `.env` file at startup.
        - [x] Create a `.env.example` file to document available configuration options.
        - [x] Replace all identified hardcoded values with the corresponding fields from the `Config` struct.
        - [x] Update `server/Dockerfile` and `docker-compose.yml` to ensure configuration is passed correctly when running in Docker.
- [ ] **Refactor Core Game Logic:**
    - [ ] Refactor `ValidateMove` and Improve Test Coverage:
        - [ ] Analyze the existing `ValidateMove` and `checkPathAlongAxis` functions in `model/game.go` to fully understand the current logic.
        - [ ] Decompose the monolithic `ValidateMove` into smaller, single-purpose functions (e.g., `isMoveInBounds`, `isPathClear`, `isValidTarget`).
        - [ ] Rewrite `ValidateMove` to be a simple orchestrator that calls the new, smaller validation functions.
        - [ ] As part of the refactoring, write comprehensive unit tests for each new function.
        - [ ] Expand the existing tests for `ValidateMove` in `model/game_test.go` to cover more edge cases and scenarios.

## Frontend Refactoring (Vue.js)

- [ ] **State Management and Component Review:**
    - [ ] Decompose `gameStore` into smaller, more focused stores:
        - [ ] Analyze `client/vue1/src/stores/gameStore.ts` to categorize state into `Core Game State`, `UI State`, and `Animation Logic`.
        - [ ] Create new Pinia stores with clear responsibilities (e.g., `uiStore.ts`, `animationStore.ts`).
        - [ ] Refactor the existing `gameStore.ts` to manage only the core game state.
        - [ ] Carefully migrate state, getters, and actions from `gameStore.ts` to the appropriate new stores.
        - [ ] Update the Vue components that currently use `gameStore.ts` to import and use the new, more specific stores.
        - [ ] Thoroughly test the application to ensure that all features are working as expected after the refactoring.
    - [ ] Decompose Large Vue Components (starting with `GameBoard.vue`):
        - [ ] Analyze `GameBoard.vue` to identify distinct responsibilities and areas for extraction.
        - [ ] Extract the game replay logic, state, and template into a new `GameReplay.vue` component.
        - [ ] Extract the solutions panel (for both during-game and post-game) into a `SolutionsPanel.vue` component.
        - [ ] Create smaller, focused components for individual board elements like `BoardTarget.vue`, `BoardRobot.vue`, `BoardWall.vue`, and `HistoryDot.vue`.
        - [ ] Move the keyboard handling logic from `GameBoard.vue` into a reusable composable function (e.g., `useGameBoardKeyboard.ts`).
        - [ ] Extract generic loading and error state rendering into `LoadingIndicator.vue` and `ErrorMessage.vue` components.
        - [ ] Update `GameBoard.vue` to utilize all the newly created components and composables, significantly reducing its complexity.
        - [ ] Analyze `PlayersPanel.vue` for decomposition opportunities.
        - [ ] Extract the timer logic into a reusable `useGameTimer.ts` composable.
        - [ ] Create a `PlayerListItem.vue` component to render a single player's information.
        - [ ] Refactor `PlayersPanel.vue` to use the new composable and component, simplifying its structure.
    - [ ] Audit and Address Logic Duplication:
        - [ ] Consolidate the duplicated `computeDirection` function:
            - [ ] Create a new `client/vue1/src/utils.ts` file for shared utility functions.
            - [ ] Move the robust version of `computeDirection` from `GameBoard.vue` into `utils.ts`.
            - [ ] Update `GameBoard.vue` and `gameStore.ts` to import and use the centralized `computeDirection` function.
        - [ ] Consolidate animation and replay logic:
            - [ ] As part of the `GameBoard.vue` decomposition, ensure all end-of-game replay logic is moved into the new `GameReplay.vue` component.
            - [ ] As part of the state management refactoring, move all animation-related logic (e.g., `unwindMoves`, `replayMoves`) from `gameStore.ts` into the new `animationStore.ts`.

## Project-Wide Improvements

- [ ] **Standardize `Move` Representation:**
    - [ ] **Backend:**
        - [ ] Update `proto/bouncebot.proto` to define a `Direction` enum and a `Move` message that includes `robot_id`, `direction`, `from_pos`, and `to_pos`.
        - [ ] Update the `PlayerSolution` message to use `repeated Move`.
        - [ ] Run `compile_protos.sh` to regenerate the Go and TypeScript code.
        - [ ] Update the Go backend (`model/game.go`, `server/session/session.go`) to use the new `Move` struct.
    - [ ] **Frontend:**
        - [ ] Create a `client/vue1/src/types.ts` file to define a canonical `Move` type.
        - [ ] Update all components and stores (e.g., `gameStore.ts`, `GameBoard.vue`) to import and use the canonical `Move` type.
        - [ ] If the backend changes are not made, create a data transformation layer to convert server data into the canonical frontend `Move` type.
- [x] **Dependency and CI/CD Audit:**
    - [x] Update Dependencies:
        - [x] **Backend (Go):**
            - [x] Run `go list -u -m all` in the root directory to list available updates.
            - [x] Run `go get -u` to update dependencies to their latest versions.
            - [x] Run `go mod tidy` to clean up the `go.mod` file.
            - [x] Run backend tests and manually test the application to ensure no breaking changes.
        - [x] **Frontend (npm):**
            - [x] Run `npm outdated` in the `client/vue1` directory to check for outdated packages.
            - [x] Run `npm update` to update the dependencies.
            - [x] Manually test the frontend to ensure no breaking changes were introduced.
    - [x] Establish CI/CD Pipeline (Docker build & publish to ghcr.io)

## Proposed Additional Refactorings (New Findings)

- [ ] **Unify Physics Engine & Logic:**
    - [ ] **Backend (Go):**
        - [ ] Implement a `ComputeMove(botId BotId, dir Direction) (Position, error)` function in `model/game.go`. This essentially "runs" the physics for one move.
        - [ ] Refactor `ValidateMove` to use `ComputeMove` internally (i.e., calculate where it *would* go, and check if it matches the requested end position).
    - [ ] **Cross-Language Testing:**
        - [ ] Create a shared JSON test fixture (e.g., `tests/physics_cases.json`) containing board setups, moves, and expected outcomes.
        - [ ] Write a Go test that executes these cases against `model/game.go`.
        - [ ] Write a TypeScript test (Vitest) that executes these cases against `client/vue1/src/gamePhysics.ts`.
        - [ ] This ensures identical behavior between client (prediction) and server (validation).

- [ ] **Decouple Frontend Animation from State:**
    - [ ] **Problem:** `gameStore.ts` currently contains hardcoded `setTimeout` delays (150ms) to manage animations. This mixes UI concerns with business logic and makes the store hard to test.
    - [ ] **Solution:** Refactor `unwindMoves` and `replayMoves` into a "Playback Controller" composable (e.g., `useGamePlayback.ts`).
        - [ ] The store should only hold the *target* state and the *current* state.
        - [ ] The composable handles the timing and dispatching of intermediate actions to the store.
        - [ ] Alternatively, use CSS transitions/Vue `<TransitionGroup>` more effectively and let the store just update the "current" positions instantly, or use a "command queue" that the UI consumes at its own pace.

- [ ] **Strengthen Type Safety:**
    - [ ] **Frontend:** Introduce Opaque Types (Branded Types) for `RobotId`, `RoomId`, and `PlayerId` to prevent accidental mixing of these integer values.
    - [ ] **Backend:** Ensure `BotId` and `PlayerId` are consistently used instead of `int` or `int32` in all interfaces.

- [ ] **API & Protobuf Optimization:**
    - [ ] **State vs Stateless:** Currently `CheckSolution` requires the client to send the full `Game` object back. This is bandwidth-heavy.
        - [ ] If `CheckSolution` is only used for local validation feedback, the client already has the logic (`gamePhysics.ts`).
        - [ ] If it's used for server verification, it should ideally rely on a `GameID` if the server is holding state (which it is for Rooms).
    - [ ] **Proposal:** Deprecate stateless `CheckSolution` in favor of client-side validation + server-side `SubmitSolution` (which is already stateful).
