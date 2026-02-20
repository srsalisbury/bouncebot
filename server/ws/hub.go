// Package ws provides WebSocket functionality for real-time room updates.
package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/srsalisbury/bouncebot/server/config"
	"github.com/srsalisbury/bouncebot/server/room"
)

const (
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	maxMessageSize = 4096
	writeWait      = 10 * time.Second
)

// OriginChecker is an interface for checking if origins are allowed.
type OriginChecker interface {
	IsOriginAllowedForRequest(origin, requestHost string) bool
}

// Event represents a WebSocket event sent to clients.
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// PlayerJoinedPayload is the payload for player_joined events.
type PlayerJoinedPayload struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

// PlayerLeftPayload is the payload for player_left events.
type PlayerLeftPayload struct {
	PlayerID string `json:"playerId"`
}

// GameStartedPayload is the payload for game_started events.
type GameStartedPayload struct {
	// Game data is sent via room refresh
}

// PlayerSolvedPayload is the payload for player_solved events.
type PlayerSolvedPayload struct {
	PlayerID  string `json:"playerId"`
	MoveCount int    `json:"moveCount"`
}

// PlayerFinishedSolvingPayload is the payload for player_finished_solving events.
type PlayerFinishedSolvingPayload struct {
	PlayerID string `json:"playerId"`
}

// PlayerReadyForNextPayload is the payload for player_ready_for_next events.
type PlayerReadyForNextPayload struct {
	PlayerID string `json:"playerId"`
}

// GameEndedPayload is the payload for game_ended events.
type GameEndedPayload struct {
	WinnerID   string             `json:"winnerId"`
	WinnerName string             `json:"winnerName"`
	Moves      []room.MovePayload `json:"moves"`
}

// SolverResultPayload is the payload for solver_complete events.
type SolverResultPayload struct {
	SolverName string             `json:"solverName"`
	Moves      []room.MovePayload `json:"moves"`
	Error      string             `json:"error,omitempty"`
	Completed  bool               `json:"completed"`
}

// SettingsChangedPayload is the payload for settings_changed events.
type SettingsChangedPayload struct {
	// Settings are fetched via room refresh
}

// Client represents a WebSocket client connection.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	roomID    string
	playerID  string
	send      chan []byte
	done      chan struct{}
	closeOnce sync.Once
}

// Hub manages WebSocket connections for all rooms.
type Hub struct {
	mu            sync.RWMutex
	rooms         map[string]map[*Client]bool // roomID -> clients
	activeClients map[string]*Client          // "roomID:playerID" -> current client
	store         *room.RoomService
	config        *config.Config
	upgrader      websocket.Upgrader
}

// NewHub creates a new WebSocket hub.
func NewHub(store *room.RoomService, cfg *config.Config) *Hub {
	h := &Hub{
		rooms:         make(map[string]map[*Client]bool),
		activeClients: make(map[string]*Client),
		store:         store,
		config:        cfg,
	}
	h.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return cfg.IsOriginAllowedForRequest(origin, r.Host)
		},
	}
	return h
}

// register adds a client to a room.
func (h *Hub) register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.roomID] == nil {
		h.rooms[client.roomID] = make(map[*Client]bool)
	}
	h.rooms[client.roomID][client] = true

	if client.playerID != "" {
		key := fmt.Sprintf("%s:%s", client.roomID, client.playerID)
		h.activeClients[key] = client
	}

	log.Printf("WebSocket: client connected to room %s (total: %d)", client.roomID, len(h.rooms[client.roomID]))
}

// unregister removes a client from a room.
func (h *Hub) unregister(client *Client) {
	shouldDisconnect := false

	h.mu.Lock()
	if clients, ok := h.rooms[client.roomID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			log.Printf("WebSocket: client disconnected from room %s (remaining: %d)", client.roomID, len(clients))
			if len(clients) == 0 {
				delete(h.rooms, client.roomID)
			}
		}
	}
	if client.playerID != "" {
		key := fmt.Sprintf("%s:%s", client.roomID, client.playerID)
		if h.activeClients[key] == client {
			delete(h.activeClients, key)
			shouldDisconnect = true
		}
	}
	h.mu.Unlock()

	if shouldDisconnect {
		h.store.DisconnectPlayer(client.roomID, client.playerID)
	}

	client.closeOnce.Do(func() {
		close(client.done)
	})
}

// BroadcastPlayerJoined broadcasts a player_joined event to all clients in a room.
func (h *Hub) BroadcastPlayerJoined(roomID, playerID, playerName string) {
	h.Broadcast(roomID, Event{
		Type: "player_joined",
		Payload: PlayerJoinedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
		},
	})
}

// BroadcastPlayerLeft broadcasts a player_left event to all clients in a room.
func (h *Hub) BroadcastPlayerLeft(roomID, playerID string) {
	h.Broadcast(roomID, Event{
		Type: "player_left",
		Payload: PlayerLeftPayload{
			PlayerID: playerID,
		},
	})
}

// BroadcastGameStarted broadcasts a game_started event to all clients in a room.
func (h *Hub) BroadcastGameStarted(roomID string) {
	h.Broadcast(roomID, Event{
		Type:    "game_started",
		Payload: GameStartedPayload{},
	})
}

// BroadcastPlayerSolved broadcasts a player_solved event to all clients in a room.
func (h *Hub) BroadcastPlayerSolved(roomID, playerID string, moveCount int) {
	h.Broadcast(roomID, Event{
		Type: "player_solved",
		Payload: PlayerSolvedPayload{
			PlayerID:  playerID,
			MoveCount: moveCount,
		},
	})
}

// BroadcastPlayerFinishedSolving broadcasts a player_finished_solving event to all clients in a room.
func (h *Hub) BroadcastPlayerFinishedSolving(roomID, playerID string) {
	h.Broadcast(roomID, Event{
		Type: "player_finished_solving",
		Payload: PlayerFinishedSolvingPayload{
			PlayerID: playerID,
		},
	})
}

// BroadcastPlayerReadyForNext broadcasts a player_ready_for_next event to all clients in a room.
func (h *Hub) BroadcastPlayerReadyForNext(roomID, playerID string) {
	h.Broadcast(roomID, Event{
		Type: "player_ready_for_next",
		Payload: PlayerReadyForNextPayload{
			PlayerID: playerID,
		},
	})
}

// BroadcastGameEnded broadcasts a game_ended event to all clients in a room.
func (h *Hub) BroadcastGameEnded(roomID, winnerID, winnerName string, moves []room.MovePayload) {
	h.Broadcast(roomID, Event{
		Type: "game_ended",
		Payload: GameEndedPayload{
			WinnerID:   winnerID,
			WinnerName: winnerName,
			Moves:      moves,
		},
	})
}

// BroadcastSolverComplete broadcasts a solver_complete event to all clients in a room.
func (h *Hub) BroadcastSolverComplete(roomID string, payload SolverResultPayload) {
	h.Broadcast(roomID, Event{
		Type:    "solver_complete",
		Payload: payload,
	})
}

// BroadcastRoomSettingsChanged broadcasts a settings_changed event to all clients in a room.
func (h *Hub) BroadcastRoomSettingsChanged(roomID string) {
	h.Broadcast(roomID, Event{
		Type:    "settings_changed",
		Payload: SettingsChangedPayload{},
	})
}

// Broadcast sends an event to all clients in a room.
func (h *Hub) Broadcast(roomID string, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("WebSocket: failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	clients := make([]*Client, 0, len(h.rooms[roomID]))
	for client := range h.rooms[roomID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	var failed []*Client
	for _, client := range clients {
		select {
		case client.send <- data:
		case <-client.done:
			// Client is shutting down, skip
		default:
			// Client's send buffer is full
			failed = append(failed, client)
		}
	}

	for _, client := range failed {
		h.unregister(client)
	}
}

// HandleWebSocket handles WebSocket connections.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomId")
	if roomID == "" {
		http.Error(w, "roomId required", http.StatusBadRequest)
		return
	}
	sessionToken := r.URL.Query().Get("sessionToken")
	if sessionToken == "" {
		http.Error(w, "sessionToken required", http.StatusBadRequest)
		return
	}

	// Validate session and reconnect if needed (all under a single room lock)
	playerID, err := h.store.ValidateAndReconnect(roomID, sessionToken)
	if err != nil {
		if errors.Is(err, room.ErrRoomNotFound) {
			http.Error(w, "room not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket: upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		roomID:   roomID,
		playerID: playerID,
		send:     make(chan []byte, 256),
		done:     make(chan struct{}),
	}

	h.register(client)

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket connection.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket: read error: %v", err)
			}
			break
		}
		// Currently we don't expect any client messages, just keep connection alive
	}
}

// writePump writes messages to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket: write error: %v", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}
