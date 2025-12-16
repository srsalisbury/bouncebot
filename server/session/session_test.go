package session

import (
	"testing"
)

func TestCreate(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")

	if session.ID == "" {
		t.Error("expected session ID to be set")
	}
	if len(session.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(session.Players))
	}
	if session.Players[0].Name != "Alice" {
		t.Errorf("expected player name 'Alice', got '%s'", session.Players[0].Name)
	}
	if session.Players[0].ID == "" {
		t.Error("expected player ID to be set")
	}
	if session.CurrentGame != nil {
		t.Error("expected no game initially")
	}
	if session.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
}

func TestCreate_UniqueIDs(t *testing.T) {
	store := NewStore()

	// Create multiple sessions and verify all IDs are unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		session := store.Create("Player")
		if ids[session.ID] {
			t.Errorf("duplicate session ID generated: %s", session.ID)
		}
		ids[session.ID] = true
	}

	if len(store.sessions) != 100 {
		t.Errorf("expected 100 sessions, got %d", len(store.sessions))
	}
}

func TestJoin(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	sessionID := session.ID

	session, err := store.Join(sessionID, "Bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(session.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(session.Players))
	}
	if session.Players[1].Name != "Bob" {
		t.Errorf("expected second player name 'Bob', got '%s'", session.Players[1].Name)
	}
}

func TestJoin_NotFound(t *testing.T) {
	store := NewStore()

	_, err := store.Join("nonexistent", "Bob")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestGet(t *testing.T) {
	store := NewStore()

	created := store.Create("Alice")

	session, err := store.Get(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.ID != created.ID {
		t.Errorf("expected session ID '%s', got '%s'", created.ID, session.ID)
	}
}

func TestGet_NotFound(t *testing.T) {
	store := NewStore()

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestStartGame(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	sessionID := session.ID

	session, err := store.StartGame(sessionID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.CurrentGame == nil {
		t.Error("expected game to be set after StartGame")
	}
	if session.GameStartedAt == nil {
		t.Error("expected game_started_at to be set")
	}
}

func TestStartGame_NotFound(t *testing.T) {
	store := NewStore()

	_, err := store.StartGame("nonexistent", false)
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestStartGame_FixedBoard(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	sessionID := session.ID

	// Start with fixed board
	session, err := store.StartGame(sessionID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.CurrentGame == nil {
		t.Error("expected game to be set after StartGame")
	}

	// Fixed board should have target at position (5, 13) for robot 0
	target := session.CurrentGame.Target
	if target.Id != 0 {
		t.Errorf("expected fixed board target robot ID 0, got %d", target.Id)
	}
	if target.Pos.X != 5 || target.Pos.Y != 13 {
		t.Errorf("expected fixed board target at (5, 13), got (%d, %d)", target.Pos.X, target.Pos.Y)
	}
}

func TestStartGame_Multiple(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	sessionID := session.ID

	// Start first game
	session, _ = store.StartGame(sessionID, false)
	firstGameStartedAt := session.GameStartedAt

	// Start second game (simulates "next game")
	session, err := store.StartGame(sessionID, false)
	if err != nil {
		t.Fatalf("unexpected error starting second game: %v", err)
	}

	if session.CurrentGame == nil {
		t.Error("expected game to be set after second StartGame")
	}
	if session.GameStartedAt == firstGameStartedAt {
		t.Error("expected game_started_at to be updated for new game")
	}
}

func TestSessionToProto(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	store.Join(session.ID, "Bob")
	store.StartGame(session.ID, false)

	session, _ = store.Get(session.ID)
	proto := session.ToProto()

	if proto.Id != session.ID {
		t.Errorf("expected proto ID '%s', got '%s'", session.ID, proto.Id)
	}
	if len(proto.Players) != 2 {
		t.Errorf("expected 2 players in proto, got %d", len(proto.Players))
	}
	if proto.CurrentGame == nil {
		t.Error("expected current_game in proto")
	}
	if proto.GameStartedAt == nil {
		t.Error("expected game_started_at in proto")
	}
}
