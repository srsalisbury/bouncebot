# Room Configuration Modal Implementation Plan

## Goal
Create a configuration modal that lets the host user set room settings that are shared with all players:
- Enable bot solution count in header (default: off)
- Enable showing bot solutions in end game (default: off)

## Requirements
- Gear icon to open the modal (only visible to host)
- Settings synchronized to all players in room via WebSocket
- Available in waiting room and during game
- Host's preferences sticky in localStorage (for initial defaults when creating rooms)

## Architecture

Settings are stored server-side in the Room and broadcast to all players when changed.

```
Host changes setting
    ↓
UpdateRoomSettings RPC (server validates host)
    ↓
Room state updated
    ↓
BroadcastRoomSettingsChanged → all WebSocket clients
    ↓
Clients call loadRoom() to get updated settings
    ↓
UI reflects new settings
```

## Files to Modify/Create

### Server-side
| File | Changes |
|------|---------|
| `proto/bouncebot.proto` | Add `RoomSettings` message and fields to `Room` |
| `server/room/room.go` | Add settings fields to Room struct |
| `server/room/service.go` | Add `UpdateRoomSettings` method |
| `server/room/signals.go` | Add `RoomSettingsChangedEvent` |
| `server/ws/hub.go` | Add `BroadcastRoomSettingsChanged` |
| `server/api/handlers.go` | Add `UpdateRoomSettings` handler |

### Client-side
| File | Changes |
|------|---------|
| `src/components/SettingsModal.vue` | **NEW** - Modal component |
| `src/stores/roomStore.ts` | Add localStorage for host defaults |
| `src/views/RoomView.vue` | Add gear icon, use room settings |
| `src/composables/useRoomConnection.ts` | Handle settings_changed event |
| `src/components/GameBoard.vue` | Conditionally show solver solutions |
| `src/components/PlayerSolutionsDrawer.vue` | Conditionally show solver solutions |

## Implementation Steps

### 1. Proto Changes (`proto/bouncebot.proto`)

```protobuf
message RoomSettings {
  bool show_solver_move_count = 1;
  bool show_solver_solutions = 2;
}

message Room {
  // ... existing fields ...
  RoomSettings settings = 14;
}

message UpdateRoomSettingsRequest {
  string room_id = 1;
  string player_id = 2;  // Must be host
  RoomSettings settings = 3;
}

message UpdateRoomSettingsResponse {
  bool success = 1;
  string error = 2;
}
```

### 2. Server Room Changes

**room.go:**
```go
type RoomSettings struct {
    ShowSolverMoveCount bool
    ShowSolverSolutions bool
}

type Room struct {
    // ... existing fields ...
    Settings RoomSettings
}
```

**signals.go:**
```go
type RoomSettingsChangedEvent struct {
    baseEvent
}
```

**service.go:**
```go
func (s *RoomService) UpdateRoomSettings(roomID, playerID string, settings RoomSettings) error {
    room, unlock := s.repo.GetWithLock(roomID)
    if room == nil {
        unlock()
        return errors.New("room not found")
    }

    // Validate player is host (first player)
    if len(room.Players) == 0 || room.Players[0].ID != playerID {
        unlock()
        return errors.New("only host can change settings")
    }

    room.Settings = settings
    unlock()

    s.broadcaster.BroadcastRoomSettingsChanged(roomID)
    return nil
}
```

### 3. Client SettingsModal.vue

```vue
<template>
  <Teleport to="body">
    <div v-if="show" class="modal-backdrop" @click.self="emit('close')">
      <div class="settings-modal">
        <button class="close-btn" @click="emit('close')">×</button>
        <h2>Room Settings</h2>
        <div class="setting-item">
          <label class="toggle-label">
            <input type="checkbox" :checked="settings.showSolverMoveCount"
                   @change="updateSetting('showSolverMoveCount', $event)" />
            <span>Show solver move count in header</span>
          </label>
        </div>
        <div class="setting-item">
          <label class="toggle-label">
            <input type="checkbox" :checked="settings.showSolverSolutions"
                   @change="updateSetting('showSolverSolutions', $event)" />
            <span>Show solver solutions after game</span>
          </label>
        </div>
      </div>
    </div>
  </Teleport>
</template>
```

### 4. Client Integration

**useRoomConnection.ts:**
- Handle `settings_changed` WebSocket event → call `loadRoom()`

**RoomView.vue:**
- Add gear icon (visible only when `isRoomCreator`)
- Read settings from `room.settings`
- Conditionally show solver move count based on `room.settings.showSolverMoveCount`
- Pass `room.settings.showSolverSolutions` to GameBoard

**roomStore.ts:**
- Store host's preferred defaults in localStorage
- When host creates room, initialize with their defaults

## Props Flow

```
Room (from server)
  └── settings: { showSolverMoveCount, showSolverSolutions }

RoomView.vue
  ├── room.settings.showSolverMoveCount → controls header solver badge
  ├── room.settings.showSolverSolutions → passed to GameBoard
  │
  └── GameBoard.vue (prop: showSolverSolutions)
        └── PlayerSolutionsDrawer.vue (prop: showSolverSolutions)
```

## Verification

1. Start server and client
2. Create room as host - see gear icon
3. Open second browser, join same room - NO gear icon
4. Host opens settings, toggles solver move count ON
5. Both browsers should see solver move count in header during game
6. Host toggles solver solutions ON
7. Both browsers should see solver solutions in end game
8. Host refreshes page - settings should persist (from room state)
9. Non-host player cannot modify settings
