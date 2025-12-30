package room

import (
	"fmt"
	"log"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

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

// processSignals interprets and executes signals.
// This is where the orchestration happens.
func (s *RoomService) processSignals(signals []Signal) {
	for _, sig := range signals {
		switch signal := sig.(type) {
		case BroadcastSignal:
			s.processBroadcast(signal.Event)

		case EndGameSignal:
			room, unlock := s.repo.GetWithLock(signal.RoomID)
			if room != nil {
				newSignals := s.gameMgr.EndGame(room)
				unlock()
				s.processSignals(newSignals)
			} else {
				unlock()
			}

		case StartNextGameSignal:
			room, unlock := s.repo.GetWithLock(signal.RoomID)
			if room != nil {
				newSignals := s.gameMgr.StartNextGame(room)
				unlock()
				s.processSignals(newSignals)
			} else {
				unlock()
			}

		case StartTimerSignal:
			s.timerMgr.StartTimer(
				signal.RoomID,
				signal.PlayerID,
				s.disconnectGracePeriod,
				s.onTimerFired,
			)

		case CancelTimerSignal:
			s.timerMgr.CancelTimer(signal.PlayerID)
		}
	}
}

func (s *RoomService) processBroadcast(event BroadcastEvent) {
	if s.broadcaster == nil {
		return
	}

	switch e := event.(type) {
	case PlayerJoinedEvent:
		s.broadcaster.BroadcastPlayerJoined(e.RoomID, e.PlayerID, e.PlayerName)
	case PlayerLeftEvent:
		s.broadcaster.BroadcastPlayerLeft(e.RoomID, e.PlayerID)
	case GameStartedEvent:
		s.broadcaster.BroadcastGameStarted(e.RoomID)
	case PlayerFinishedSolvingEvent:
		s.broadcaster.BroadcastPlayerFinishedSolving(e.RoomID, e.PlayerID)
	case PlayerReadyForNextEvent:
		s.broadcaster.BroadcastPlayerReadyForNext(e.RoomID, e.PlayerID)
	case PlayerSolvedEvent:
		s.broadcaster.BroadcastPlayerSolved(e.RoomID, e.PlayerID, e.MoveCount)
	case SolutionRetractedEvent:
		s.broadcaster.BroadcastSolutionRetracted(e.RoomID, e.PlayerID)
	case GameEndedEvent:
		s.broadcaster.BroadcastGameEnded(e.RoomID, e.WinnerID, e.WinnerName, e.Moves)
	}
}

func (s *RoomService) onTimerFired(roomID, playerID string) {
	s.RemovePlayer(roomID, playerID)
}

// ---- Public API (backward compatible with old Store) ----

// Create creates a new room with the given player.
func (s *RoomService) Create(playerName string) *Room {
	return s.repo.Create(playerName)
}

// Join adds a player to an existing room.
func (s *RoomService) Join(roomID, playerName string) (*Room, error) {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	signals, err := s.playerMgr.AddPlayer(room, playerName)
	unlock()

	if err != nil {
		return nil, err
	}

	s.processSignals(signals)
	return room, nil
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
func (s *RoomService) StartGame(roomID string, useFixedBoard bool) (*Room, error) {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	signals, err := s.gameMgr.StartGame(room, useFixedBoard)
	unlock()

	if err != nil {
		return nil, err
	}

	s.processSignals(signals)
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

// RetractSolution removes a player's current solution.
func (s *RoomService) RetractSolution(roomID, playerID string) error {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return fmt.Errorf("room not found: %s", roomID)
	}

	signals, err := s.solutionMgr.RetractSolution(room, playerID)
	unlock()

	if err != nil {
		return err
	}

	s.processSignals(signals)
	return nil
}

// MarkFinishedSolving marks a player as finished solving.
func (s *RoomService) MarkFinishedSolving(roomID, playerID string) error {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return fmt.Errorf("room not found: %s", roomID)
	}

	signals, err := s.gameMgr.MarkFinishedSolving(room, playerID)
	unlock()

	if err != nil {
		return err
	}

	s.processSignals(signals)
	return nil
}

// MarkReadyForNext marks a player as ready for the next game.
func (s *RoomService) MarkReadyForNext(roomID, playerID string) error {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return fmt.Errorf("room not found: %s", roomID)
	}

	signals, err := s.gameMgr.MarkReadyForNext(room, playerID)
	unlock()

	if err != nil {
		return err
	}

	s.processSignals(signals)
	return nil
}

// DisconnectPlayer marks a player as disconnected.
func (s *RoomService) DisconnectPlayer(roomID, playerID string) error {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return fmt.Errorf("room not found: %s", roomID)
	}

	signals, err := s.playerMgr.DisconnectPlayer(room, playerID)
	unlock()

	if err != nil {
		return err
	}

	s.processSignals(signals)
	return nil
}

// ReconnectPlayer marks a player as connected.
func (s *RoomService) ReconnectPlayer(roomID, playerID string) error {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return fmt.Errorf("room not found: %s", roomID)
	}

	signals, err := s.playerMgr.ReconnectPlayer(room, playerID)
	unlock()

	if err != nil {
		return err
	}

	s.processSignals(signals)
	return nil
}

// RemovePlayer removes a player from a room.
func (s *RoomService) RemovePlayer(roomID, playerID string) {
	room, unlock := s.repo.GetWithLock(roomID)
	if room == nil {
		unlock()
		return
	}

	signals := s.playerMgr.RemovePlayer(room, playerID)
	unlock()

	s.processSignals(signals)
}

// ---- Persistence Methods ----

// Load loads rooms from the data file.
func (s *RoomService) Load(filename string) error {
	rooms, err := s.persistence.Load(filename)
	if err != nil {
		return err
	}
	s.repo.Replace(rooms)
	return nil
}

// Save saves all rooms to the data file.
func (s *RoomService) Save(filename string) error {
	return s.persistence.Save(filename, s.repo.All())
}

// StartAutoSave starts a goroutine that periodically saves rooms.
func (s *RoomService) StartAutoSave(filename string, interval time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.Save(filename); err != nil {
					log.Printf("Auto-save failed: %v", err)
				}
			case <-stop:
				// Final save before stopping
				if err := s.Save(filename); err != nil {
					log.Printf("Final save failed: %v", err)
				}
				return
			}
		}
	}()

	return stop
}

// CleanupStaleRooms removes rooms that have been inactive for longer than maxAge.
func (s *RoomService) CleanupStaleRooms(maxAge time.Duration) int {
	stale := s.persistence.FindStaleRooms(s.repo.All(), maxAge)
	for _, id := range stale {
		s.repo.Delete(id)
	}

	if len(stale) > 0 {
		log.Printf("Cleaned up %d stale rooms (inactive for >%v)", len(stale), maxAge)
	}

	return len(stale)
}

// StartCleanup starts a goroutine that periodically removes stale rooms.
func (s *RoomService) StartCleanup(interval, maxAge time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.CleanupStaleRooms(maxAge)
			case <-stop:
				return
			}
		}
	}()

	return stop
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
