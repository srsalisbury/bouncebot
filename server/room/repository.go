package room

import (
	crypto_rand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"
)

// RoomRepository provides thread-safe CRUD operations for rooms.
// Uses per-room locking for better concurrency.
type RoomRepository interface {
	// Create creates a new room with the given player.
	// If isSinglePlayer is true, no other players can join.
	// Returns the room, player ID, and session token.
	Create(playerName string, isSinglePlayer bool) (*Room, string, string)

	// Get retrieves a room by ID. Returns nil if not found.
	Get(roomID string) *Room

	// GetWithLock retrieves a room and holds a lock on it.
	// Caller MUST call the returned unlock function when done.
	// Returns (nil, no-op func) if room not found.
	GetWithLock(roomID string) (*Room, func())

	// Delete removes a room by ID.
	Delete(roomID string)

	// All returns a copy of all rooms (for persistence).
	All() map[string]*Room

	// Replace replaces all rooms (for loading from persistence).
	Replace(rooms map[string]*Room)

	// Count returns the number of rooms.
	Count() int
}

// roomRepository is the concrete implementation of RoomRepository.
type roomRepository struct {
	mu    sync.RWMutex
	rooms map[string]*Room
	locks map[string]*sync.Mutex // per-room locks
}

// NewRoomRepository creates a new RoomRepository.
func NewRoomRepository() RoomRepository {
	return &roomRepository{
		rooms: make(map[string]*Room),
		locks: make(map[string]*sync.Mutex),
	}
}

// roomIDChars is the character set for room IDs (letters only, no digits)
const roomIDChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// generateRoomID creates a random 4-character room ID.
func generateRoomID() string {
	result := make([]byte, 4)
	for i := range result {
		result[i] = roomIDChars[rand.IntN(len(roomIDChars))]
	}
	return string(result)
}

// generatePlayerID creates a random player ID.
func generatePlayerID() string {
	return fmt.Sprintf("%016x", rand.Uint64())
}

// generateSessionToken creates a cryptographically secure session token.
func generateSessionToken() string {
	bytes := make([]byte, 32)
	if _, err := crypto_rand.Read(bytes); err != nil {
		// Fallback to less secure but still random token
		return fmt.Sprintf("%016x%016x", rand.Uint64(), rand.Uint64())
	}
	return hex.EncodeToString(bytes)
}

func (r *roomRepository) Create(playerName string, isSinglePlayer bool) (*Room, string, string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate unique room ID (retry if collision)
	roomID := generateRoomID()
	for r.rooms[roomID] != nil {
		roomID = generateRoomID()
	}

	playerID := generatePlayerID()
	sessionToken := generateSessionToken()
	now := time.Now()

	room := &Room{
		ID: roomID,
		Players: []Player{
			{ID: playerID, SessionToken: sessionToken, Name: playerName, Status: PlayerStatusConnected},
		},
		CreatedAt:      now,
		LastActivityAt: now,
		Wins:           make(map[string]int),
		IsSinglePlayer: isSinglePlayer,
		Settings:       RoomSettings{MinSolutionLength: MinMinSolutionLength},
	}

	r.rooms[roomID] = room
	r.locks[roomID] = &sync.Mutex{}
	return room, playerID, sessionToken
}

func (r *roomRepository) Get(roomID string) *Room {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedID := strings.ToUpper(roomID)
	return r.rooms[normalizedID]
}

func (r *roomRepository) GetWithLock(roomID string) (*Room, func()) {
	normalizedID := strings.ToUpper(roomID)

	r.mu.RLock()
	room := r.rooms[normalizedID]
	lock := r.locks[normalizedID]
	r.mu.RUnlock()

	if room == nil || lock == nil {
		return nil, func() {}
	}

	lock.Lock()
	return room, lock.Unlock
}

func (r *roomRepository) Delete(roomID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalizedID := strings.ToUpper(roomID)
	delete(r.rooms, normalizedID)
	delete(r.locks, normalizedID)
}

func (r *roomRepository) All() map[string]*Room {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a shallow copy to avoid concurrent map access issues
	result := make(map[string]*Room, len(r.rooms))
	for k, v := range r.rooms {
		result[k] = v
	}
	return result
}

func (r *roomRepository) Replace(rooms map[string]*Room) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rooms = rooms
	if r.rooms == nil {
		r.rooms = make(map[string]*Room)
	}

	// Create locks for all rooms
	r.locks = make(map[string]*sync.Mutex, len(r.rooms))
	for id := range r.rooms {
		r.locks[id] = &sync.Mutex{}
	}
}

func (r *roomRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.rooms)
}
