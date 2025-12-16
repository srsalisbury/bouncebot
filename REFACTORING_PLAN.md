# Refactoring Plan

## Backend Refactoring (Go)

- [ ] **Address Critical Bugs and Configuration:**
    - [ ] Implement Player Disconnection Events
    - [ ] Externalize Configuration
- [ ] **Refactor Core Game Logic:**
    - [ ] Simplify `ValidateMove`
    - [ ] Improve Test Coverage for `ValidateMove`

## Frontend Refactoring (Vue.js)

- [ ] **State Management and Component Review:**
    - [ ] Analyze State Management in Pinia stores
    - [ ] Review and Decompose Large Vue Components
    - [ ] Audit and Address Logic Duplication

## Project-Wide Improvements

- [ ] **Dependency and CI/CD Audit:**
    - [ ] Update Dependencies
    - [x] Establish CI/CD Pipeline (Docker build & publish to ghcr.io)
