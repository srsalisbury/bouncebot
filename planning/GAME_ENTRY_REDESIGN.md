# Game Entry Redesign Plan

## Overview

Simplify the game entry experience to make it easy to start playing and easy for new users to understand what the game is about. The primary action becomes single-player mode, with multiplayer as a secondary option.

## Current State

- **Home screen**: Name input + Create Room / Join Room buttons
- **Flow**: Enter name → Create or Join room → Wait in lobby → Start game
- **No single-player mode**: Must create a room even to play alone
- **Help page**: Exists but lacks overview/introduction

---

## Design

### Main Entry Screen

```
┌─────────────────────────────┐
│         BounceBot           │
│  A Ricochet Robots puzzle   │
│                             │
│    [ Play Solo ]  (primary) │
│                             │
│  [ Play with Friends ]      │
│                             │
│  ┌─────────────────────┐    │
│  │ [ Create Room ]     │    │
│  │ Name: [________]    │    │
│  │                     │    │
│  │ [ Join Room ]       │    │
│  │ Room: [____]        │    │
│  │ Name: [________]    │    │
│  └─────────────────────┘    │
│                             │
│      How to Play            │
│                             │
│  [ Return to game ]         │
└─────────────────────────────┘
```

- **Play Solo**: Primary button, immediately starts single-player game
- **Play with Friends**: Secondary button, expands to show Create/Join options
- **How to Play**: Link to help page
- **Return to game**: Appears when there's an active session (single or multiplayer)

### Single-Player Mode

- No name input required (use default like "Player")
- Room created behind the scenes (user never sees room ID)
- Game auto-starts immediately (no lobby screen)
- No leaderboard or solutions list displayed
- No solution review phase after solving
- "Next Puzzle" button advances to next game
- Stats tracked: puzzles solved, puzzles attempted
- Stats shown in dropdown from header area
- Session resumable on page reload (localStorage)

### Multiplayer Mode

- **Create Room**: Name input + button, creates room and enters lobby
- **Join Room**: Room ID + name inputs + button, joins existing room
- Same gameplay experience as current implementation

### Leave Game

- 'x' keybinding anywhere in game/room view
- Confirmation dialog: "Leave this game?"
- Clears session and returns to home screen

### Help Page Updates

- Add "What is BounceBot?" overview section at top
- Add basic gameplay instructions (goal, how robots move, how to solve)
- Keep existing detailed reference content (controls, keybindings)

---

## Implementation Plan

### Phase 1: Quick Wins

**Step 1: Update "Return to game" text**
- File: `client/vue1/src/views/HomeView.vue`
- Change "Return to Room {ID}" to "Return to game"

**Step 2: Add leave game keybinding**
- File: `client/vue1/src/views/RoomView.vue`
- Add 'x' keybinding
- Show confirmation dialog
- Clear session (roomStore.clear()) and navigate to home

---

### Phase 2: Single-Player Foundation

**Step 3: Add single-player mode tracking**
- File: `client/vue1/src/stores/roomStore.ts`
- Add `isSinglePlayer` flag (persisted to localStorage)
- Add `puzzlesSolved` and `puzzlesAttempted` counters
- Add functions to update stats

**Step 4: Single-player game flow**
- File: `client/vue1/src/views/HomeView.vue`
- Add `startSoloGame()` function
- Auto-create room with default name "Player"
- Auto-start game immediately after room creation
- Set `isSinglePlayer = true` in roomStore

**Step 5: Simplified single-player UI**
- File: `client/vue1/src/views/RoomView.vue`
- File: `client/vue1/src/components/GameBoard.vue`
- Hide leaderboard when `isSinglePlayer`
- Hide solutions list when `isSinglePlayer`
- Replace solution review phase with "Next Puzzle" button
- Increment stats on puzzle completion

**Step 6: Stats dropdown in header**
- File: `client/vue1/src/components/GameBoard.vue` (or new component)
- Add dropdown menu to header area
- Show "Puzzles solved: X" and "Puzzles attempted: Y"
- Only visible when `isSinglePlayer`

---

### Phase 3: Home Screen Redesign

**Step 7: Redesign HomeView**
- File: `client/vue1/src/views/HomeView.vue`
- "Play Solo" as primary button (green, prominent)
- "Play with Friends" as secondary button
- Expandable section with Create Room / Join Room
- Create Room: name input + button
- Join Room: room ID + name inputs + button
- "How to Play" link
- "Return to game" button (when applicable)

---

### Phase 4: Help Page

**Step 8: Update help page**
- File: `client/vue1/src/views/HelpView.vue`
- Add "What is BounceBot?" section explaining the game concept
- Add "How to Play" section with basic instructions:
  - Goal: Move robots to get target robot to target square
  - Robots slide until they hit a wall or another robot
  - Find the shortest solution
- Keep existing keybinding/control reference

---

## Files to Modify

| File | Changes |
|------|---------|
| `src/views/HomeView.vue` | New layout, Play Solo button, expanded multiplayer section |
| `src/views/RoomView.vue` | Leave game keybinding, single-player UI hiding |
| `src/components/GameBoard.vue` | Stats dropdown, single-player mode adjustments |
| `src/stores/roomStore.ts` | isSinglePlayer flag, stats tracking |
| `src/views/HelpView.vue` | Overview and basic instructions sections |

---

## Testing Considerations

- Test single-player flow from start to multiple puzzles
- Test page reload resumability for both single and multiplayer
- Test leave game keybinding and confirmation
- Test "Return to game" for both session types
- Verify multiplayer flow still works correctly
- Test on mobile (touch) and desktop (keyboard)
