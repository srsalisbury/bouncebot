// Package ws provides WebSocket functionality for real-time session updates.
package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/srsalisbury/bouncebot/server/session"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Allow any localhost port
		if strings.HasPrefix(origin, "http://localhost:") || origin == "http://localhost" {
			return true
		}
		// Allow guido.local for local network dev
		if strings.HasPrefix(origin, "http://guido.local:") || origin == "http://guido.local" {
			return true
		}
		return false
	},
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
	// Game data is sent via session refresh
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
	WinnerID   string                `json:"winnerId"`
	WinnerName string                `json:"winnerName"`
	Moves      []session.MovePayload `json:"moves"`
}

// Client represents a WebSocket client connection.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	sessionID string
	send      chan []byte
}

// Hub manages WebSocket connections for all sessions.
type Hub struct {
	mu       sync.RWMutex
	sessions map[string]map[*Client]bool // sessionID -> clients
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		sessions: make(map[string]map[*Client]bool),
	}
}

// register adds a client to a session.
func (h *Hub) register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.sessions[client.sessionID] == nil {
		h.sessions[client.sessionID] = make(map[*Client]bool)
	}
	h.sessions[client.sessionID][client] = true
	log.Printf("WebSocket: client connected to session %s (total: %d)", client.sessionID, len(h.sessions[client.sessionID]))
}

// unregister removes a client from a session.
func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.sessions[client.sessionID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)
			log.Printf("WebSocket: client disconnected from session %s (remaining: %d)", client.sessionID, len(clients))
			if len(clients) == 0 {
				delete(h.sessions, client.sessionID)
			}
		}
	}
}

// BroadcastPlayerJoined broadcasts a player_joined event to all clients in a session.
func (h *Hub) BroadcastPlayerJoined(sessionID, playerID, playerName string) {
	h.Broadcast(sessionID, Event{
		Type: "player_joined",
		Payload: PlayerJoinedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
		},
	})
}

// BroadcastGameStarted broadcasts a game_started event to all clients in a session.
func (h *Hub) BroadcastGameStarted(sessionID string) {
	h.Broadcast(sessionID, Event{
		Type:    "game_started",
		Payload: GameStartedPayload{},
	})
}

// BroadcastPlayerSolved broadcasts a player_solved event to all clients in a session.
func (h *Hub) BroadcastPlayerSolved(sessionID, playerID string, moveCount int) {
	h.Broadcast(sessionID, Event{
		Type: "player_solved",
		Payload: PlayerSolvedPayload{
			PlayerID:  playerID,
			MoveCount: moveCount,
		},
	})
}

// BroadcastSolutionRetracted broadcasts a solution_retracted event to all clients in a session.
func (h *Hub) BroadcastSolutionRetracted(sessionID, playerID string) {
	h.Broadcast(sessionID, Event{
		Type: "solution_retracted",
		Payload: SolutionRetractedPayload{
			PlayerID: playerID,
		},
	})
}

// BroadcastPlayerFinishedSolving broadcasts a player_finished_solving event to all clients in a session.
func (h *Hub) BroadcastPlayerFinishedSolving(sessionID, playerID string) {
	h.Broadcast(sessionID, Event{
		Type: "player_finished_solving",
		Payload: PlayerFinishedSolvingPayload{
			PlayerID: playerID,
		},
	})
}

// BroadcastPlayerReadyForNext broadcasts a player_ready_for_next event to all clients in a session.
func (h *Hub) BroadcastPlayerReadyForNext(sessionID, playerID string) {
	h.Broadcast(sessionID, Event{
		Type: "player_ready_for_next",
		Payload: PlayerReadyForNextPayload{
			PlayerID: playerID,
		},
	})
}

// BroadcastGameEnded broadcasts a game_ended event to all clients in a session.
func (h *Hub) BroadcastGameEnded(sessionID, winnerID, winnerName string, moves []session.MovePayload) {
	h.Broadcast(sessionID, Event{
		Type: "game_ended",
		Payload: GameEndedPayload{
			WinnerID:   winnerID,
			WinnerName: winnerName,
			Moves:      moves,
		},
	})
}

// Broadcast sends an event to all clients in a session.
func (h *Hub) Broadcast(sessionID string, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("WebSocket: failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	clients := h.sessions[sessionID]
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
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "sessionId required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket: upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:       h,
		conn:      conn,
		sessionID: sessionID,
		send:      make(chan []byte, 256),
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
