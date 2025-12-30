package room

import (
	"fmt"
	"time"
)

// PlayerManager handles player connection state transitions.
// Does NOT manage timers directly - returns signals for timer operations.
type PlayerManager interface {
	// AddPlayer adds a player to a room.
	// Returns signals or error.
	AddPlayer(room *Room, playerName string) ([]Signal, error)

	// DisconnectPlayer marks a player as disconnected.
	// Returns signals or error.
	DisconnectPlayer(room *Room, playerID string) ([]Signal, error)

	// ReconnectPlayer marks a player as connected.
	// Returns signals or error.
	ReconnectPlayer(room *Room, playerID string) ([]Signal, error)

	// RemovePlayer removes a disconnected player from the room.
	// Returns signals indicating state changes (including potential game transitions).
	RemovePlayer(room *Room, playerID string) []Signal
}

// playerManager is the concrete implementation of PlayerManager.
type playerManager struct{}

// NewPlayerManager creates a new PlayerManager.
func NewPlayerManager() PlayerManager {
	return &playerManager{}
}

func (pm *playerManager) AddPlayer(room *Room, playerName string) ([]Signal, error) {
	playerID := generatePlayerID()
	room.Players = append(room.Players, Player{
		ID:     playerID,
		Name:   playerName,
		Status: PlayerStatusConnected,
	})
	room.LastActivityAt = time.Now()

	signals := []Signal{
		BroadcastSignal{Event: PlayerJoinedEvent{
			RoomID:     room.ID,
			PlayerID:   playerID,
			PlayerName: playerName,
		}},
	}

	return signals, nil
}

func (pm *playerManager) DisconnectPlayer(room *Room, playerID string) ([]Signal, error) {
	idx := room.FindPlayerIndex(playerID)
	if idx == -1 {
		// Player might have been removed already, which is fine
		return nil, nil
	}

	player := &room.Players[idx]
	player.Status = PlayerStatusDisconnected
	player.DisconnectedAt = time.Now()

	signals := []Signal{
		StartTimerSignal{RoomID: room.ID, PlayerID: playerID},
	}

	return signals, nil
}

func (pm *playerManager) ReconnectPlayer(room *Room, playerID string) ([]Signal, error) {
	idx := room.FindPlayerIndex(playerID)
	if idx == -1 {
		return nil, fmt.Errorf("player not found: %s", playerID)
	}

	player := &room.Players[idx]
	if player.Status == PlayerStatusDisconnected {
		player.Status = PlayerStatusConnected

		signals := []Signal{
			CancelTimerSignal{PlayerID: playerID},
		}
		return signals, nil
	}

	return nil, nil
}

func (pm *playerManager) RemovePlayer(room *Room, playerID string) []Signal {
	idx := room.FindPlayerIndex(playerID)
	if idx == -1 {
		return nil
	}

	// Only remove if disconnected
	if room.Players[idx].Status != PlayerStatusDisconnected {
		return nil
	}

	// Track game state BEFORE removal
	hasGame := room.CurrentGame != nil

	// Remove player
	room.Players = append(room.Players[:idx], room.Players[idx+1:]...)

	// Clean up from FinishedSolving
	for i, id := range room.FinishedSolving {
		if id == playerID {
			room.FinishedSolving = removeStringAt(room.FinishedSolving, i)
			break
		}
	}

	// Clean up from ReadyForNext
	for i, id := range room.ReadyForNext {
		if id == playerID {
			room.ReadyForNext = removeStringAt(room.ReadyForNext, i)
			break
		}
	}

	// Clean up from Solutions
	for i, sol := range room.Solutions {
		if sol.PlayerID == playerID {
			room.Solutions = append(room.Solutions[:i], room.Solutions[i+1:]...)
			break
		}
	}

	signals := []Signal{
		CancelTimerSignal{PlayerID: playerID},
		BroadcastSignal{Event: PlayerLeftEvent{RoomID: room.ID, PlayerID: playerID}},
	}

	// Check if removal triggers game state changes
	if len(room.Players) > 0 {
		// If game is active and all remaining players are finished, signal end game
		if hasGame && len(room.FinishedSolving) == len(room.Players) {
			signals = append(signals, EndGameSignal{RoomID: room.ID})
		}
		// If all remaining players are ready for next, signal start next game
		if len(room.ReadyForNext) == len(room.Players) {
			signals = append(signals, StartNextGameSignal{RoomID: room.ID})
		}
	}

	return signals
}
