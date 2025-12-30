package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/server/config"
	"github.com/srsalisbury/bouncebot/server/room"
)

// mockClient creates a test client with a buffered send channel.
func mockClient(hub *Hub, roomID, playerID string) *Client {
	return &Client{
		hub:      hub,
		conn:     nil, // No actual WebSocket connection for tests
		roomID:   roomID,
		playerID: playerID,
		send:     make(chan []byte, 256),
	}
}

func TestHubRegisterUnregister(t *testing.T) {
	store := room.NewRoomService()
	cfg := &config.Config{}
	hub := NewHub(store, cfg)

	client1 := mockClient(hub, "ROOM1", "player1")
	client2 := mockClient(hub, "ROOM1", "player2")

	// Register first client
	hub.register(client1)

	hub.mu.RLock()
	if len(hub.rooms["ROOM1"]) != 1 {
		t.Errorf("expected 1 client in ROOM1, got %d", len(hub.rooms["ROOM1"]))
	}
	hub.mu.RUnlock()

	// Register second client in same room
	hub.register(client2)

	hub.mu.RLock()
	if len(hub.rooms["ROOM1"]) != 2 {
		t.Errorf("expected 2 clients in ROOM1, got %d", len(hub.rooms["ROOM1"]))
	}
	hub.mu.RUnlock()

	// Unregister first client
	hub.unregister(client1)

	hub.mu.RLock()
	if len(hub.rooms["ROOM1"]) != 1 {
		t.Errorf("expected 1 client in ROOM1 after unregister, got %d", len(hub.rooms["ROOM1"]))
	}
	hub.mu.RUnlock()

	// Unregister second client - room should be removed
	hub.unregister(client2)

	hub.mu.RLock()
	if _, exists := hub.rooms["ROOM1"]; exists {
		t.Error("expected ROOM1 to be removed when empty")
	}
	hub.mu.RUnlock()
}

func TestHubBroadcastDelivery(t *testing.T) {
	store := room.NewRoomService()
	cfg := &config.Config{}
	hub := NewHub(store, cfg)

	client := mockClient(hub, "ROOM1", "player1")
	hub.register(client)

	// Broadcast an event
	hub.Broadcast("ROOM1", Event{
		Type:    "test_event",
		Payload: map[string]string{"key": "value"},
	})

	// Check that client received the message
	select {
	case msg := <-client.send:
		var event Event
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Type != "test_event" {
			t.Errorf("expected event type 'test_event', got '%s'", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client did not receive broadcast message")
	}

	hub.unregister(client)
}

func TestHubRoomIsolation(t *testing.T) {
	store := room.NewRoomService()
	cfg := &config.Config{}
	hub := NewHub(store, cfg)

	client1 := mockClient(hub, "ROOM1", "player1")
	client2 := mockClient(hub, "ROOM2", "player2")
	hub.register(client1)
	hub.register(client2)

	// Broadcast to ROOM1 only
	hub.Broadcast("ROOM1", Event{Type: "room1_event", Payload: nil})

	// client1 should receive the message
	select {
	case msg := <-client1.send:
		var event Event
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Type != "room1_event" {
			t.Errorf("client1 received wrong event type: %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client1 did not receive broadcast message")
	}

	// client2 should NOT receive the message
	select {
	case msg := <-client2.send:
		t.Errorf("client2 should not receive ROOM1 message, got: %s", string(msg))
	case <-time.After(50 * time.Millisecond):
		// Expected - no message for client2
	}

	hub.unregister(client1)
	hub.unregister(client2)
}

func TestBroadcastPlayerJoined(t *testing.T) {
	store := room.NewRoomService()
	cfg := &config.Config{}
	hub := NewHub(store, cfg)

	client := mockClient(hub, "ROOM1", "player1")
	hub.register(client)

	hub.BroadcastPlayerJoined("ROOM1", "newPlayer", "NewPlayerName")

	select {
	case msg := <-client.send:
		var event Event
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Type != "player_joined" {
			t.Errorf("expected event type 'player_joined', got '%s'", event.Type)
		}
		payload, ok := event.Payload.(map[string]interface{})
		if !ok {
			t.Fatalf("payload is not a map")
		}
		if payload["playerId"] != "newPlayer" {
			t.Errorf("expected playerId 'newPlayer', got '%v'", payload["playerId"])
		}
		if payload["playerName"] != "NewPlayerName" {
			t.Errorf("expected playerName 'NewPlayerName', got '%v'", payload["playerName"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client did not receive broadcast message")
	}

	hub.unregister(client)
}

func TestBroadcastPlayerSolved(t *testing.T) {
	store := room.NewRoomService()
	cfg := &config.Config{}
	hub := NewHub(store, cfg)

	client := mockClient(hub, "ROOM1", "player1")
	hub.register(client)

	hub.BroadcastPlayerSolved("ROOM1", "solver", 5)

	select {
	case msg := <-client.send:
		var event Event
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Type != "player_solved" {
			t.Errorf("expected event type 'player_solved', got '%s'", event.Type)
		}
		payload, ok := event.Payload.(map[string]interface{})
		if !ok {
			t.Fatalf("payload is not a map")
		}
		if payload["playerId"] != "solver" {
			t.Errorf("expected playerId 'solver', got '%v'", payload["playerId"])
		}
		// JSON unmarshals numbers as float64
		if payload["moveCount"].(float64) != 5 {
			t.Errorf("expected moveCount 5, got '%v'", payload["moveCount"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client did not receive broadcast message")
	}

	hub.unregister(client)
}

func TestBroadcastGameEnded(t *testing.T) {
	store := room.NewRoomService()
	cfg := &config.Config{}
	hub := NewHub(store, cfg)

	client := mockClient(hub, "ROOM1", "player1")
	hub.register(client)

	moves := []room.MovePayload{
		{RobotId: 0, X: 5, Y: 3},
		{RobotId: 1, X: 7, Y: 2},
	}
	hub.BroadcastGameEnded("ROOM1", "winner123", "WinnerName", moves)

	select {
	case msg := <-client.send:
		var event Event
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Type != "game_ended" {
			t.Errorf("expected event type 'game_ended', got '%s'", event.Type)
		}
		payload, ok := event.Payload.(map[string]interface{})
		if !ok {
			t.Fatalf("payload is not a map")
		}
		if payload["winnerId"] != "winner123" {
			t.Errorf("expected winnerId 'winner123', got '%v'", payload["winnerId"])
		}
		if payload["winnerName"] != "WinnerName" {
			t.Errorf("expected winnerName 'WinnerName', got '%v'", payload["winnerName"])
		}
		movesPayload, ok := payload["moves"].([]interface{})
		if !ok {
			t.Fatalf("moves is not a slice")
		}
		if len(movesPayload) != 2 {
			t.Errorf("expected 2 moves, got %d", len(movesPayload))
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("client did not receive broadcast message")
	}

	hub.unregister(client)
}

func TestBroadcastToEmptyRoom(t *testing.T) {
	store := room.NewRoomService()
	cfg := &config.Config{}
	hub := NewHub(store, cfg)

	// Should not panic when broadcasting to a room with no clients
	hub.Broadcast("NONEXISTENT", Event{Type: "test", Payload: nil})
}

func TestMultipleClientsInRoom(t *testing.T) {
	store := room.NewRoomService()
	cfg := &config.Config{}
	hub := NewHub(store, cfg)

	client1 := mockClient(hub, "ROOM1", "player1")
	client2 := mockClient(hub, "ROOM1", "player2")
	client3 := mockClient(hub, "ROOM1", "player3")

	hub.register(client1)
	hub.register(client2)
	hub.register(client3)

	hub.BroadcastGameStarted("ROOM1")

	// All three clients should receive the message
	clients := []*Client{client1, client2, client3}
	for i, client := range clients {
		select {
		case msg := <-client.send:
			var event Event
			if err := json.Unmarshal(msg, &event); err != nil {
				t.Fatalf("client%d: failed to unmarshal event: %v", i+1, err)
			}
			if event.Type != "game_started" {
				t.Errorf("client%d: expected event type 'game_started', got '%s'", i+1, event.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("client%d did not receive broadcast message", i+1)
		}
	}

	hub.unregister(client1)
	hub.unregister(client2)
	hub.unregister(client3)
}
