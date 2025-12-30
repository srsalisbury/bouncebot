package room

import (
	"testing"
)

func TestStartGame(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	roomID := room.ID

	room, err := store.StartGame(roomID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.CurrentGame == nil {
		t.Error("expected game to be set after StartGame")
	}
	if room.GameStartedAt == nil {
		t.Error("expected game_started_at to be set")
	}
}

func TestStartGame_NotFound(t *testing.T) {
	store := NewStore()

	_, err := store.StartGame("nonexistent", false)
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestStartGame_FixedBoard(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	roomID := room.ID

	// Start with fixed board
	room, err := store.StartGame(roomID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.CurrentGame == nil {
		t.Error("expected game to be set after StartGame")
	}

	// Fixed board should have target at position (5, 13) for robot 0
	target := room.CurrentGame.Target
	if target.Id != 0 {
		t.Errorf("expected fixed board target robot ID 0, got %d", target.Id)
	}
	if target.Pos.X != 5 || target.Pos.Y != 13 {
		t.Errorf("expected fixed board target at (5, 13), got (%d, %d)", target.Pos.X, target.Pos.Y)
	}
}

func TestStartGame_Multiple(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	roomID := room.ID

	// Start first game
	room, _ = store.StartGame(roomID, false)
	firstGameStartedAt := room.GameStartedAt

	// Start second game (simulates "next game")
	room, err := store.StartGame(roomID, false)
	if err != nil {
		t.Fatalf("unexpected error starting second game: %v", err)
	}

	if room.CurrentGame == nil {
		t.Error("expected game to be set after second StartGame")
	}
	if room.GameStartedAt == firstGameStartedAt {
		t.Error("expected game_started_at to be updated for new game")
	}
}
