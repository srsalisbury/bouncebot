// Package ws provides WebSocket functionality for real-time room updates.
package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/srsalisbury/bouncebot/server/config"
	"github.com/srsalisbury/bouncebot/server/room"
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

// SolutionRetractedPayload is the payload for solution_retracted events.
type SolutionRetractedPayload struct {
	PlayerID string `json:"playerId"`
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

// Client represents a WebSocket client connection.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	roomID   string
	playerID string
	send     chan []byte
}

// Hub manages WebSocket connections for all rooms.
type Hub struct {
	mu       sync.RWMutex
	rooms    map[string]map[*Client]bool // roomID -> clients
	store    *room.RoomService
	config   *config.Config
	upgrader websocket.Upgrader
}

// NewHub creates a new WebSocket hub.
func NewHub(store *room.RoomService, cfg *config.Config) *Hub {
	h := &Hub{
		rooms:  make(map[string]map[*Client]bool),
		store:  store,
		config: cfg,
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
	log.Printf("WebSocket: client connected to room %s (total: %d)", client.roomID, len(h.rooms[client.roomID]))
}

// unregister removes a client from a room.
func (h *Hub) unregister(client *Client) {
	if client.playerID != "" {
		h.store.DisconnectPlayer(client.roomID, client.playerID)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[client.roomID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)
			log.Printf("WebSocket: client disconnected from room %s (remaining: %d)", client.roomID, len(clients))
			if len(clients) == 0 {
				delete(h.rooms, client.roomID)
			}
		}
	}
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

// BroadcastSolutionRetracted broadcasts a solution_retracted event to all clients in a room.
func (h *Hub) BroadcastSolutionRetracted(roomID, playerID string) {
	h.Broadcast(roomID, Event{
		Type: "solution_retracted",
		Payload: SolutionRetractedPayload{
			PlayerID: playerID,
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

// Broadcast sends an event to all clients in a room.
func (h *Hub) Broadcast(roomID string, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("WebSocket: failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	clients := h.rooms[roomID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.send <- data:
		default:
			// Client's send buffer is full, close connection
			h.unregister(client)
		}
	}
}

// HandleWebSocket handles WebSocket connections.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomId")
	if roomID == "" {
		http.Error(w, "roomId required", http.StatusBadRequest)
		return
	}
	playerID := r.URL.Query().Get("playerId")
	if playerID == "" {
		http.Error(w, "playerId required", http.StatusBadRequest)
		return
	}

	// Check if player can reconnect
	rm, err := h.store.Get(roomID)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	var player room.Player
	found := false
	for _, p := range rm.Players {
		if p.ID == playerID {
			player = p
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "player not found", http.StatusForbidden)
		return
	}

	if player.Status == room.PlayerStatusDisconnected {
		if err := h.store.ReconnectPlayer(roomID, playerID); err != nil {
			log.Printf("WebSocket: failed to reconnect player %s in room %s: %v", playerID, roomID, err)
			http.Error(w, "failed to reconnect", http.StatusInternalServerError)
			return
		}
		log.Printf("WebSocket: player %s reconnected to room %s", playerID, roomID)
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
	defer c.conn.Close()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("WebSocket: write error: %v", err)
			return
		}
	}
}
