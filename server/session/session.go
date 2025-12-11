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

// Session represents a multiplayer game session.
type Session struct {
	ID            string
	Players       []Player
	CreatedAt     time.Time
	CurrentGame   *model.Game
	GameStartedAt *time.Time
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

	session := &pb.Session{
		Id:        s.ID,
		Players:   players,
		CreatedAt: timestamppb.New(s.CreatedAt),
	}

	if s.CurrentGame != nil {
		session.CurrentGame = s.CurrentGame.ToProto()
	}

	if s.GameStartedAt != nil {
		session.GameStartedAt = timestamppb.New(*s.GameStartedAt)
	}

	return session
}

// Store manages sessions in memory.
type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewStore creates a new session store.
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*Session),
	}
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

	return session, nil
}
