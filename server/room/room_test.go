package room

import (
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")

	if room.ID == "" {
		t.Error("expected room ID to be set")
	}
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(room.Players))
	}
	if room.Players[0].Name != "Alice" {
		t.Errorf("expected player name 'Alice', got '%s'", room.Players[0].Name)
	}
	if room.Players[0].ID == "" {
		t.Error("expected player ID to be set")
	}
	if room.CurrentGame != nil {
		t.Error("expected no game initially")
	}
	if room.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
}

func TestCreate_UniqueIDs(t *testing.T) {
	store := NewStore()

	// Create multiple rooms and verify all IDs are unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		room := store.Create("Player")
		if ids[room.ID] {
			t.Errorf("duplicate room ID generated: %s", room.ID)
		}
		ids[room.ID] = true
	}

	if len(store.rooms) != 100 {
		t.Errorf("expected 100 rooms, got %d", len(store.rooms))
	}
}

func TestJoin(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	roomID := room.ID

	room, err := store.Join(roomID, "Bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(room.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(room.Players))
	}
	if room.Players[1].Name != "Bob" {
		t.Errorf("expected second player name 'Bob', got '%s'", room.Players[1].Name)
	}
}

func TestJoin_NotFound(t *testing.T) {
	store := NewStore()

	_, err := store.Join("nonexistent", "Bob")
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestJoin_CaseInsensitive(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	roomID := room.ID

	// Join with lowercase version of room ID
	lowercaseID := strings.ToLower(roomID)
	room, err := store.Join(lowercaseID, "Bob")
	if err != nil {
		t.Fatalf("expected case-insensitive join to work: %v", err)
	}

	if len(room.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(room.Players))
	}
}

func TestGet(t *testing.T) {
	store := NewStore()

	created := store.Create("Alice")

	room, err := store.Get(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.ID != created.ID {
		t.Errorf("expected room ID '%s', got '%s'", created.ID, room.ID)
	}
}

func TestGet_NotFound(t *testing.T) {
	store := NewStore()

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestRoomToProto(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.Join(room.ID, "Bob")
	store.StartGame(room.ID, false)

	room, _ = store.Get(room.ID)
	proto := room.ToProto()

	if proto.Id != room.ID {
		t.Errorf("expected proto ID '%s', got '%s'", room.ID, proto.Id)
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
