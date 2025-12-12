// Package session provides multiplayer game session management.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Player represents a player in a session.
type Player struct {
	ID   string
	Name string
}

// PlayerSolution represents a player's solution to the current game.
type PlayerSolution struct {
	PlayerID   string
	PlayerName string
	MoveCount  int
	SolvedAt   time.Time
}

// PlayerSolutionHistory tracks all solutions a player has found (for restoring after retraction).
type PlayerSolutionHistory struct {
	PlayerID   string
	PlayerName string
	Solutions  []struct {
		MoveCount int
		SolvedAt  time.Time
	}
}

// Session represents a multiplayer game session.
type Session struct {
	ID              string
	Players         []Player
	CreatedAt       time.Time
	CurrentGame     *model.Game
	GameStartedAt   *time.Time
	Solutions       []PlayerSolution          // Current best solution per player
	SolutionHistory []PlayerSolutionHistory   // All solutions per player (for retraction)
}

// ToProto converts a Session to its protobuf representation.
func (s *Session) ToProto() *pb.Session {
	players := make([]*pb.Player, len(s.Players))
	for i, p := range s.Players {
		players[i] = &pb.Player{
			Id:   p.ID,
			Name: p.Name,
		}
	}

	solutions := make([]*pb.PlayerSolution, len(s.Solutions))
	for i, sol := range s.Solutions {
		solutions[i] = &pb.PlayerSolution{
			PlayerId:   sol.PlayerID,
			PlayerName: sol.PlayerName,
			MoveCount:  int32(sol.MoveCount),
			SolvedAt:   timestamppb.New(sol.SolvedAt),
		}
	}

	session := &pb.Session{
		Id:        s.ID,
		Players:   players,
		CreatedAt: timestamppb.New(s.CreatedAt),
		Solutions: solutions,
	}

	if s.CurrentGame != nil {
		session.CurrentGame = s.CurrentGame.ToProto()
	}

	if s.GameStartedAt != nil {
		session.GameStartedAt = timestamppb.New(*s.GameStartedAt)
	}

	return session
}

// EventBroadcaster is an interface for broadcasting session events.
type EventBroadcaster interface {
	BroadcastPlayerJoined(sessionID, playerID, playerName string)
	BroadcastGameStarted(sessionID string)
	BroadcastPlayerSolved(sessionID, playerID string, moveCount int)
	BroadcastSolutionRetracted(sessionID, playerID string)
}

// Store manages sessions in memory.
type Store struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	broadcaster EventBroadcaster
}

// NewStore creates a new session store.
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*Session),
	}
}

// SetBroadcaster sets the event broadcaster for the store.
func (store *Store) SetBroadcaster(b EventBroadcaster) {
	store.broadcaster = b
}

// generateID creates a random session or player ID.
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Create creates a new session with the given player.
func (store *Store) Create(playerName string) *Session {
	store.mu.Lock()
	defer store.mu.Unlock()

	sessionID := generateID()
	playerID := generateID()

	session := &Session{
		ID: sessionID,
		Players: []Player{
			{ID: playerID, Name: playerName},
		},
		CreatedAt: time.Now(),
	}

	store.sessions[sessionID] = session
	return session
}

// Join adds a player to an existing session.
func (store *Store) Join(sessionID, playerName string) (*Session, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	playerID := generateID()
	session.Players = append(session.Players, Player{
		ID:   playerID,
		Name: playerName,
	})

	// Broadcast player joined event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerJoined(sessionID, playerID, playerName)
	}

	return session, nil
}

// Get retrieves a session by ID.
func (store *Store) Get(sessionID string) (*Session, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// StartGame starts a new game in the session.
func (store *Store) StartGame(sessionID string) (*Session, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Generate a new game
	game := model.Game1()
	now := time.Now()

	session.CurrentGame = game
	session.GameStartedAt = &now
	session.Solutions = nil        // Clear solutions for new game
	session.SolutionHistory = nil  // Clear history for new game

	// Broadcast game started event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastGameStarted(sessionID)
	}

	return session, nil
}

// SubmitSolution records a player's solution for the current game.
func (store *Store) SubmitSolution(sessionID, playerID string, moveCount int) (*PlayerSolution, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if session.CurrentGame == nil {
		return nil, fmt.Errorf("no game in progress")
	}

	// Find player name from ID
	var playerName string
	for _, p := range session.Players {
		if p.ID == playerID {
			playerName = p.Name
			break
		}
	}
	if playerName == "" {
		return nil, fmt.Errorf("player not found: %s", playerID)
	}

	now := time.Now()

	// Add to solution history
	store.addToHistory(session, playerID, playerName, moveCount, now)

	// Check if player already submitted a solution for this game
	for i := range session.Solutions {
		if session.Solutions[i].PlayerID == playerID {
			// Update if better solution
			if moveCount < session.Solutions[i].MoveCount {
				session.Solutions[i].MoveCount = moveCount
				session.Solutions[i].SolvedAt = now
				// Broadcast updated solution
				if store.broadcaster != nil {
					store.broadcaster.BroadcastPlayerSolved(sessionID, playerID, moveCount)
				}
			}
			// Return existing (possibly updated) solution
			return &session.Solutions[i], nil
		}
	}

	solution := PlayerSolution{
		PlayerID:   playerID,
		PlayerName: playerName,
		MoveCount:  moveCount,
		SolvedAt:   now,
	}
	session.Solutions = append(session.Solutions, solution)

	// Broadcast player solved event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerSolved(sessionID, playerID, moveCount)
	}

	return &solution, nil
}

// addToHistory adds a solution to the player's history (if not already present with same move count).
func (store *Store) addToHistory(session *Session, playerID, playerName string, moveCount int, solvedAt time.Time) {
	// Find or create history entry for this player
	var history *PlayerSolutionHistory
	for i := range session.SolutionHistory {
		if session.SolutionHistory[i].PlayerID == playerID {
			history = &session.SolutionHistory[i]
			break
		}
	}
	if history == nil {
		session.SolutionHistory = append(session.SolutionHistory, PlayerSolutionHistory{
			PlayerID:   playerID,
			PlayerName: playerName,
		})
		history = &session.SolutionHistory[len(session.SolutionHistory)-1]
	}

	// Check if we already have this move count in history
	for _, sol := range history.Solutions {
		if sol.MoveCount == moveCount {
			return // Already have this solution
		}
	}

	// Add to history
	history.Solutions = append(history.Solutions, struct {
		MoveCount int
		SolvedAt  time.Time
	}{
		MoveCount: moveCount,
		SolvedAt:  solvedAt,
	})
}

// RetractSolution removes a player's current best solution and restores the previous one from history.
func (store *Store) RetractSolution(sessionID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.CurrentGame == nil {
		return fmt.Errorf("no game in progress")
	}

	// Find the player's current solution
	var currentMoveCount int
	var solutionIndex int = -1
	for i, sol := range session.Solutions {
		if sol.PlayerID == playerID {
			currentMoveCount = sol.MoveCount
			solutionIndex = i
			break
		}
	}

	if solutionIndex == -1 {
		return fmt.Errorf("no solution found for player: %s", playerID)
	}

	// Find the player's history and remove the current solution from it
	var history *PlayerSolutionHistory
	for i := range session.SolutionHistory {
		if session.SolutionHistory[i].PlayerID == playerID {
			history = &session.SolutionHistory[i]
			break
		}
	}

	// Remove current move count from history
	if history != nil {
		for i, sol := range history.Solutions {
			if sol.MoveCount == currentMoveCount {
				history.Solutions = append(history.Solutions[:i], history.Solutions[i+1:]...)
				break
			}
		}

		// Find the next best solution in history (smallest move count remaining)
		if len(history.Solutions) > 0 {
			bestIdx := 0
			for i, sol := range history.Solutions {
				if sol.MoveCount < history.Solutions[bestIdx].MoveCount {
					bestIdx = i
				}
			}
			// Restore the previous best solution
			session.Solutions[solutionIndex].MoveCount = history.Solutions[bestIdx].MoveCount
			session.Solutions[solutionIndex].SolvedAt = history.Solutions[bestIdx].SolvedAt

			// Broadcast the restored solution
			if store.broadcaster != nil {
				store.broadcaster.BroadcastPlayerSolved(sessionID, playerID, history.Solutions[bestIdx].MoveCount)
			}
			return nil
		}
	}

	// No previous solution to restore - remove completely
	session.Solutions = append(session.Solutions[:solutionIndex], session.Solutions[solutionIndex+1:]...)

	// Broadcast solution retracted event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastSolutionRetracted(sessionID, playerID)
	}

	return nil
}
