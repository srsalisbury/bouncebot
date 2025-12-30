package room

import (
	"fmt"
	"time"
)

// Player represents a player in a room.
type Player struct {
	ID             string
	Name           string
	Status         PlayerStatus
	DisconnectedAt time.Time
}

// PlayerStatus represents the connection status of a player.
type PlayerStatus string

const (
	// PlayerStatusConnected means the player is actively connected.
	PlayerStatusConnected PlayerStatus = "connected"
	// PlayerStatusDisconnected means the player has disconnected but is within the grace period.
	PlayerStatusDisconnected PlayerStatus = "disconnected"
)

// DisconnectPlayer marks a player as disconnected and starts a timer to remove them.
func (store *Store) DisconnectPlayer(roomID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	idx := room.FindPlayerIndex(playerID)
	if idx == -1 {
		// Player might have been removed already, which is fine
		return nil
	}
	player := &room.Players[idx]

	player.Status = PlayerStatusDisconnected
	player.DisconnectedAt = time.Now()

	// Clean up any existing timer for this player
	if oldTimer, ok := store.timers[playerID]; ok {
		oldTimer.Stop()
		delete(store.timers, playerID)
	}

	timer := time.AfterFunc(store.disconnectGracePeriod, func() {
		store.RemovePlayer(roomID, playerID)
	})
	store.timers[playerID] = timer

	return nil
}

// ReconnectPlayer marks a player as connected and cancels their removal timer.
func (store *Store) ReconnectPlayer(roomID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	idx := room.FindPlayerIndex(playerID)
	if idx == -1 {
		return fmt.Errorf("player not found: %s", playerID)
	}
	player := &room.Players[idx]

	if player.Status == PlayerStatusDisconnected {
		player.Status = PlayerStatusConnected
		if timer, ok := store.timers[playerID]; ok {
			timer.Stop()
			delete(store.timers, playerID)
		}
	}

	return nil
}

// RemovePlayer removes a player from a room if they are still disconnected.
func (store *Store) RemovePlayer(roomID, playerID string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return // room already gone
	}

	playerIndex := room.FindPlayerIndex(playerID)
	if playerIndex == -1 {
		return
	}

	// Only remove if still disconnected
	if room.Players[playerIndex].Status != PlayerStatusDisconnected {
		return
	}

	room.Players = append(room.Players[:playerIndex], room.Players[playerIndex+1:]...)
	delete(store.timers, playerID)

	// Remove from FinishedSolving
	for i, id := range room.FinishedSolving {
		if id == playerID {
			room.FinishedSolving = removeStringAt(room.FinishedSolving, i)
			break
		}
	}

	// Remove from ReadyForNext
	for i, id := range room.ReadyForNext {
		if id == playerID {
			room.ReadyForNext = removeStringAt(room.ReadyForNext, i)
			break
		}
	}

	// Remove from Solutions
	for i, sol := range room.Solutions {
		if sol.PlayerID == playerID {
			room.Solutions = append(room.Solutions[:i], room.Solutions[i+1:]...)
			break
		}
	}

	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerLeft(roomID, playerID)
	}

	// Check if remaining players now satisfy conditions
	if len(room.Players) > 0 {
		// If game is active and all remaining players are finished, end the game
		if room.CurrentGame != nil && len(room.FinishedSolving) == len(room.Players) {
			store.endGame(room)
		}
		// If all remaining players are ready for next, start next game
		if len(room.ReadyForNext) == len(room.Players) {
			store.startNextGame(room)
		}
	}
}
