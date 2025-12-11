package main

import (
	"testing"
)

func TestCreateSession(t *testing.T) {
	store := NewSessionStore()

	session := store.CreateSession("Alice")

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

func TestJoinSession(t *testing.T) {
	store := NewSessionStore()

	session := store.CreateSession("Alice")
	sessionID := session.ID

	session, err := store.JoinSession(sessionID, "Bob")
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

func TestJoinSession_NotFound(t *testing.T) {
	store := NewSessionStore()

	_, err := store.JoinSession("nonexistent", "Bob")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestGetSession(t *testing.T) {
	store := NewSessionStore()

	created := store.CreateSession("Alice")

	session, err := store.GetSession(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.ID != created.ID {
		t.Errorf("expected session ID '%s', got '%s'", created.ID, session.ID)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	store := NewSessionStore()

	_, err := store.GetSession("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestStartGame(t *testing.T) {
	store := NewSessionStore()

	session := store.CreateSession("Alice")
	sessionID := session.ID

	session, err := store.StartGame(sessionID)
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
	store := NewSessionStore()

	_, err := store.StartGame("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestStartGame_Multiple(t *testing.T) {
	store := NewSessionStore()

	session := store.CreateSession("Alice")
	sessionID := session.ID

	// Start first game
	session, _ = store.StartGame(sessionID)
	firstGameStartedAt := session.GameStartedAt

	// Start second game (simulates "next game")
	session, err := store.StartGame(sessionID)
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
	store := NewSessionStore()

	session := store.CreateSession("Alice")
	store.JoinSession(session.ID, "Bob")
	store.StartGame(session.ID)

	session, _ = store.GetSession(session.ID)
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
