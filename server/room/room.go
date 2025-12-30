// Package room provides multiplayer game room management.
package room

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Room represents a multiplayer game room.
type Room struct {
	ID              string
	Players         []Player
	CreatedAt       time.Time
	LastActivityAt  time.Time               // Last user action timestamp (for cleanup)
	CurrentGame     *model.Game
	GameStartedAt   *time.Time
	Solutions       []PlayerSolution        // Current best solution per player
	SolutionHistory []PlayerSolutionHistory // All solutions per player (for retraction)
	Wins            map[string]int          // Wins per player ID
	GamesPlayed     int                     // Total games completed in room
	FinishedSolving []string                // Player IDs who are finished solving (triggers game end)
	ReadyForNext    []string                // Player IDs who are ready for next game
}

// GetPlayerName returns the name of the player with the given ID, or empty string if not found.
func (r *Room) GetPlayerName(playerID string) string {
	for _, p := range r.Players {
		if p.ID == playerID {
			return p.Name
		}
	}
	return ""
}

// ToProto converts a Room to its protobuf representation.
func (r *Room) ToProto() *pb.Room {
	players := make([]*pb.Player, len(r.Players))
	for i, p := range r.Players {
		players[i] = &pb.Player{
			Id:   p.ID,
			Name: p.Name,
		}
	}

	solutions := make([]*pb.PlayerSolution, len(r.Solutions))
	for i, sol := range r.Solutions {
		moves := make([]*pb.BotPos, len(sol.Moves))
		for j, move := range sol.Moves {
			moves[j] = move.ToProto()
		}
		solutions[i] = &pb.PlayerSolution{
			PlayerId: sol.PlayerID,
			SolvedAt: timestamppb.New(sol.SolvedAt),
			Moves:    moves,
		}
	}

	// Convert wins map to proto
	scores := make([]*pb.PlayerScore, 0, len(r.Wins))
	for playerID, wins := range r.Wins {
		scores = append(scores, &pb.PlayerScore{
			PlayerId: playerID,
			Wins:     int32(wins),
		})
	}

	room := &pb.Room{
		Id:              r.ID,
		Players:         players,
		CreatedAt:       timestamppb.New(r.CreatedAt),
		Solutions:       solutions,
		Scores:          scores,
		GamesPlayed:     int32(r.GamesPlayed),
		FinishedSolving: r.FinishedSolving,
		ReadyForNext:    r.ReadyForNext,
	}

	if r.CurrentGame != nil {
		room.CurrentGame = r.CurrentGame.ToProto()
	}

	if r.GameStartedAt != nil {
		room.GameStartedAt = timestamppb.New(*r.GameStartedAt)
	}

	return room
}

// MovePayload represents a single move for WebSocket broadcast.
type MovePayload struct {
	RobotId int `json:"robotId"`
	X       int `json:"x"`
	Y       int `json:"y"`
}

// EventBroadcaster is an interface for broadcasting room events.
type EventBroadcaster interface {
	BroadcastPlayerJoined(roomID, playerID, playerName string)
	BroadcastPlayerLeft(roomID, playerID string)
	BroadcastGameStarted(roomID string)
	BroadcastPlayerFinishedSolving(roomID, playerID string)
	BroadcastPlayerReadyForNext(roomID, playerID string)
	BroadcastPlayerSolved(roomID, playerID string, moveCount int)
	BroadcastSolutionRetracted(roomID, playerID string)
	BroadcastGameEnded(roomID, winnerID, winnerName string, moves []MovePayload)
}

// Store manages rooms in memory.
type Store struct {
	mu                    sync.RWMutex
	rooms                 map[string]*Room
	broadcaster           EventBroadcaster
	timers                map[string]*time.Timer // playerID -> timer
	disconnectGracePeriod time.Duration
}

// NewStore creates a new room store.
func NewStore() *Store {
	return &Store{
		rooms:                 make(map[string]*Room),
		timers:                make(map[string]*time.Timer),
		disconnectGracePeriod: 30 * time.Second, // default
	}
}

// SetBroadcaster sets the event broadcaster for the store.
func (store *Store) SetBroadcaster(b EventBroadcaster) {
	store.broadcaster = b
}

// SetDisconnectGracePeriod sets the grace period for player disconnection.
func (store *Store) SetDisconnectGracePeriod(d time.Duration) {
	store.disconnectGracePeriod = d
}

// roomIDChars is the character set for room IDs (no 0, 1, I, O to avoid confusion)
const roomIDChars = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

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

// Create creates a new room with the given player.
func (store *Store) Create(playerName string) *Room {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Generate unique room ID (retry if collision)
	roomID := generateRoomID()
	for store.rooms[roomID] != nil {
		roomID = generateRoomID()
	}

	playerID := generatePlayerID()
	now := time.Now()

	room := &Room{
		ID: roomID,
		Players: []Player{
			{ID: playerID, Name: playerName, Status: PlayerStatusConnected},
		},
		CreatedAt:      now,
		LastActivityAt: now,
		Wins:           make(map[string]int),
	}

	store.rooms[roomID] = room
	return room
}

// Join adds a player to an existing room.
func (store *Store) Join(roomID, playerName string) (*Room, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Normalize room ID to uppercase for case-insensitive matching
	normalizedID := strings.ToUpper(roomID)
	room, ok := store.rooms[normalizedID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	playerID := generatePlayerID()
	room.Players = append(room.Players, Player{
		ID:     playerID,
		Name:   playerName,
		Status: PlayerStatusConnected,
	})
	room.LastActivityAt = time.Now()

	// Broadcast player joined event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerJoined(roomID, playerID, playerName)
	}

	return room, nil
}

// Get retrieves a room by ID.
func (store *Store) Get(roomID string) (*Room, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	// Normalize room ID to uppercase for case-insensitive matching
	normalizedID := strings.ToUpper(roomID)
	room, ok := store.rooms[normalizedID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	return room, nil
}

