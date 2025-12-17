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
    - [ ] Externalize Configuration:
        - [ ] Scan the Go codebase (starting with `server/main.go`) to identify all hardcoded configuration values (e.g., CORS policy, server port).
        - [ ] Choose and implement a configuration library (e.g., `godotenv`) to load settings from a `.env` file.
        - [ ] Create a `Config` struct in the `server` directory to hold all application configuration in a type-safe manner.
        - [ ] Implement logic in `server/main.go` to load configuration from the `.env` file at startup.
        - [ ] Create a `.env.example` file to document available configuration options.
        - [ ] Replace all identified hardcoded values with the corresponding fields from the `Config` struct.
        - [ ] Update `server/Dockerfile` and `docker-compose.yml` to ensure configuration is passed correctly when running in Docker.
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
- [ ] **Dependency and CI/CD Audit:**
    - [ ] Update Dependencies:
        - [ ] **Backend (Go):**
            - [ ] Run `go list -u -m all` in the root directory to list available updates.
            - [ ] Run `go get -u` to update dependencies to their latest versions.
            - [ ] Run `go mod tidy` to clean up the `go.mod` file.
            - [ ] Run backend tests and manually test the application to ensure no breaking changes.
        - [ ] **Frontend (npm):**
            - [ ] Run `npm outdated` in the `client/vue1` directory to check for outdated packages.
            - [ ] Run `npm update` to update the dependencies.
            - [ ] Manually test the frontend to ensure no breaking changes were introduced.
    - [x] Establish CI/CD Pipeline (Docker build & publish to ghcr.io)
