package room

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// ErrRoomNotFound is returned when a room doesn't exist.
var ErrRoomNotFound = errors.New("room not found")

// GameStartCallback is called when a new game starts.
type GameStartCallback func(room *Room)

// RoomService is the facade that orchestrates all room operations.
// It interprets signals and coordinates components.
type RoomService struct {
	repo        RoomRepository
	playerMgr   PlayerManager
	gameMgr     GameLifecycle
	solutionMgr SolutionManager
	persistence PersistenceManager
	timerMgr    TimerManager

	broadcaster           EventBroadcaster
	disconnectGracePeriod time.Duration
	onGameStart           GameStartCallback
}

// NewRoomService creates a new RoomService with all components.
func NewRoomService() *RoomService {
	solutionMgr := NewSolutionManager()
	return &RoomService{
		repo:                  NewRoomRepository(),
		playerMgr:             NewPlayerManager(),
		gameMgr:               NewGameLifecycle(solutionMgr),
		solutionMgr:           solutionMgr,
		persistence:           NewPersistenceManager(),
		timerMgr:              NewTimerManager(),
		disconnectGracePeriod: 30 * time.Second,
	}
}

// SetBroadcaster sets the event broadcaster.
func (s *RoomService) SetBroadcaster(b EventBroadcaster) {
	s.broadcaster = b
}

// SetDisconnectGracePeriod sets the grace period for player disconnection.
func (s *RoomService) SetDisconnectGracePeriod(d time.Duration) {
	s.disconnectGracePeriod = d
}

// SetOnGameStart sets the callback for when a new game starts.
func (s *RoomService) SetOnGameStart(cb GameStartCallback) {
	s.onGameStart = cb
}

// ---- Public API (backward compatible with old Store) ----

// Create creates a new room with the given player.
// If isSinglePlayer is true, no other players can join.
func (s *RoomService) Create(playerName string, isSinglePlayer bool) *Room {
	return s.repo.Create(playerName, isSinglePlayer)
}

// Join adds a player to an existing room.
// Returns the room, the new player's ID, and any error.
func (s *RoomService) Join(roomID, playerName string) (*Room, string, error) {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return nil, "", fmt.Errorf("room not found: %s", roomID)
	}

	// Reject joins to single player rooms
	if room.IsSinglePlayer {
		unlock()
		return nil, "", fmt.Errorf("cannot join single player room")
	}

	playerID, signals, err := s.playerMgr.AddPlayer(room, playerName)
	unlock()

	if err != nil {
		return nil, "", err
	}

	s.processSignals(signals)
	return room, playerID, nil
}

// Get retrieves a room by ID.
func (s *RoomService) Get(roomID string) (*Room, error) {
	room := s.repo.Get(roomID)
	if room == nil {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}
	return room, nil
}

// StartGame starts a new game in the room.
func (s *RoomService) StartGame(roomID string) (*Room, error) {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	signals, err := s.gameMgr.StartGame(room)
	// Make a copy for callback (room might be modified after unlock)
	roomCopy := *room
	unlock()

	if err != nil {
		return nil, err
	}

	s.processSignals(signals)

	// Notify about game start (for solver etc)
	if s.onGameStart != nil && roomCopy.CurrentGame != nil {
		s.onGameStart(&roomCopy)
	}

	return room, nil
}

// SubmitSolution records a player's solution.
func (s *RoomService) SubmitSolution(roomID, playerID string, moves []model.BotPosition) (*PlayerSolution, error) {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	solution, signals, err := s.solutionMgr.SubmitSolution(room, playerID, moves)
	unlock()

	if err != nil {
		return nil, err
	}

	s.processSignals(signals)
	return solution, nil
}

// MarkFinishedSolving marks a player as finished solving.
func (s *RoomService) MarkFinishedSolving(roomID, playerID string) error {
	return s.withRoomLock(roomID, func(room *Room) ([]Signal, error) {
		return s.gameMgr.MarkFinishedSolving(room, playerID)
	})
}

// MarkReadyForNext marks a player as ready for the next game.
func (s *RoomService) MarkReadyForNext(roomID, playerID string) error {
	return s.withRoomLock(roomID, func(room *Room) ([]Signal, error) {
		return s.gameMgr.MarkReadyForNext(room, playerID)
	})
}

// DisconnectPlayer marks a player as disconnected.
func (s *RoomService) DisconnectPlayer(roomID, playerID string) error {
	return s.withRoomLock(roomID, func(room *Room) ([]Signal, error) {
		return s.playerMgr.DisconnectPlayer(room, playerID)
	})
}

// ReconnectPlayer marks a player as connected.
func (s *RoomService) ReconnectPlayer(roomID, playerID string) error {
	return s.withRoomLock(roomID, func(room *Room) ([]Signal, error) {
		return s.playerMgr.ReconnectPlayer(room, playerID)
	})
}

// RemovePlayer removes a player from a room.
// If the room becomes empty (no players and no pending players), the room is deleted.
func (s *RoomService) RemovePlayer(roomID, playerID string) {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return
	}

	signals := s.playerMgr.RemovePlayer(room, playerID)

	// Check if room is now empty and should be garbage collected
	roomEmpty := len(room.Players) == 0 && len(room.PendingPlayers) == 0
	unlock()

	s.processSignals(signals)

	// Delete the room if it's empty
	if roomEmpty {
		s.repo.Delete(roomID)
		log.Printf("Room %s garbage collected (no players remaining)", roomID)
	}
}

// UpdateRoomSettings updates the room settings. Only the host (first player) can change settings.
func (s *RoomService) UpdateRoomSettings(roomID, playerID string, settings RoomSettings) error {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return fmt.Errorf("room not found: %s", roomID)
	}

	// Validate player is host (first player)
	if len(room.Players) == 0 || room.Players[0].ID != playerID {
		unlock()
		return fmt.Errorf("only host can change settings")
	}

	room.Settings = settings
	unlock()

	// Broadcast the change to all clients
	if s.broadcaster != nil {
		s.broadcaster.BroadcastRoomSettingsChanged(roomID)
	}

	return nil
}

// BootPlayer removes a player from the room. Only the host (first player) can boot players.
// The host can boot themselves, in which case the next player becomes host.
func (s *RoomService) BootPlayer(roomID, hostPlayerID, targetPlayerID string) error {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return fmt.Errorf("room not found: %s", roomID)
	}

	// Validate caller is host (first player)
	if len(room.Players) == 0 || room.Players[0].ID != hostPlayerID {
		unlock()
		return fmt.Errorf("only host can boot players")
	}

	// Find target player
	targetIdx := -1
	for i, p := range room.Players {
		if p.ID == targetPlayerID {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		unlock()
		return fmt.Errorf("target player not found: %s", targetPlayerID)
	}

	// Use ForceRemovePlayer to remove regardless of connection status
	signals := s.playerMgr.ForceRemovePlayer(room, targetPlayerID)

	// Check if room is now empty and should be garbage collected
	roomEmpty := len(room.Players) == 0 && len(room.PendingPlayers) == 0
	unlock()

	s.processSignals(signals)

	// Delete the room if it's empty
	if roomEmpty {
		s.repo.Delete(roomID)
		log.Printf("Room %s garbage collected (no players remaining)", roomID)
	}

	return nil
}

// SetSolverResult stores a solver result in a room (upserts by solver name).
func (s *RoomService) SetSolverResult(roomID string, result *SolverResult) {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return
	}
	if room.SolverResults == nil {
		room.SolverResults = make(map[string]*SolverResult)
	}
	room.SolverResults[result.SolverName] = result
	unlock()
}

// ---- Test Helpers ----

// rooms returns the internal rooms map (for testing only).
func (s *RoomService) rooms() map[string]*Room {
	return s.repo.All()
}

// setRoom directly sets a room (for testing only).
func (s *RoomService) setRoom(id string, room *Room) {
	rooms := s.repo.All()
	rooms[id] = room
	s.repo.Replace(rooms)
}

// hasTimer returns true if a timer exists for the given player (for testing only).
func (s *RoomService) hasTimer(playerID string) bool {
	return s.timerMgr.HasTimer(playerID)
}
