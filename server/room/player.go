package room

import "time"

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
