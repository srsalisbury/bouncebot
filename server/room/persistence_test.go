package room

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/server/config"
)

func TestLoad_NonExistentFile(t *testing.T) {
	store := NewStore()
	err := store.Load("/nonexistent/path/rooms.json")
	if err != nil {
		t.Errorf("Load should not error on non-existent file, got: %v", err)
	}
	if len(store.rooms) != 0 {
		t.Errorf("Expected empty rooms, got %d", len(store.rooms))
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	// Create empty file
	if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	store := NewStore()
	err := store.Load(filename)
	if err != nil {
		t.Errorf("Load should not error on empty file, got: %v", err)
	}
	if len(store.rooms) != 0 {
		t.Errorf("Expected empty rooms, got %d", len(store.rooms))
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	// Create file with invalid JSON
	if err := os.WriteFile(filename, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	store := NewStore()
	err := store.Load(filename)
	if err == nil {
		t.Error("Load should error on invalid JSON")
	}
}

func TestLoad_ValidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	// Create valid room data
	pd := persistedData{
		Rooms: map[string]*Room{
			"TEST": {
				ID:        "TEST",
				Players:   []Player{{ID: "player1", Name: "Alice"}},
				CreatedAt: time.Now(),
				Wins:      map[string]int{"player1": 3},
			},
		},
		SavedAt: time.Now(),
		Version: 1,
	}
	data, _ := json.Marshal(pd)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	store := NewStore()
	err := store.Load(filename)
	if err != nil {
		t.Errorf("Load should not error on valid JSON, got: %v", err)
	}
	if len(store.rooms) != 1 {
		t.Errorf("Expected 1 room, got %d", len(store.rooms))
	}
	if store.rooms["TEST"] == nil {
		t.Error("Expected room 'TEST' to exist")
	}
	if store.rooms["TEST"].Players[0].Name != "Alice" {
		t.Errorf("Expected player name 'Alice', got '%s'", store.rooms["TEST"].Players[0].Name)
	}
	if store.rooms["TEST"].Wins["player1"] != 3 {
		t.Errorf("Expected 3 wins, got %d", store.rooms["TEST"].Wins["player1"])
	}
}

func TestLoad_InitializesNilWins(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	// Create room data with nil Wins map
	pd := persistedData{
		Rooms: map[string]*Room{
			"TEST": {
				ID:        "TEST",
				Players:   []Player{{ID: "player1", Name: "Alice"}},
				CreatedAt: time.Now(),
				Wins:      nil, // Explicitly nil
			},
		},
		SavedAt: time.Now(),
		Version: 1,
	}
	data, _ := json.Marshal(pd)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	store := NewStore()
	err := store.Load(filename)
	if err != nil {
		t.Errorf("Load should not error, got: %v", err)
	}
	if store.rooms["TEST"].Wins == nil {
		t.Error("Expected Wins map to be initialized, got nil")
	}
}

func TestSave_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	store := NewStore()
	store.rooms["TEST"] = &Room{
		ID:        "TEST",
		Players:   []Player{{ID: "player1", Name: "Bob"}},
		CreatedAt: time.Now(),
		Wins:      map[string]int{},
	}

	err := store.Save(filename)
	if err != nil {
		t.Errorf("Save should not error, got: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Expected file to be created")
	}

	// Verify content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var pd persistedData
	if err := json.Unmarshal(data, &pd); err != nil {
		t.Fatalf("Failed to parse saved JSON: %v", err)
	}

	if len(pd.Rooms) != 1 {
		t.Errorf("Expected 1 room in saved data, got %d", len(pd.Rooms))
	}
	if pd.Rooms["TEST"].Players[0].Name != "Bob" {
		t.Errorf("Expected player name 'Bob', got '%s'", pd.Rooms["TEST"].Players[0].Name)
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")
	tmpFile := filename + ".tmp"

	store := NewStore()
	store.rooms["TEST"] = &Room{
		ID:        "TEST",
		Players:   []Player{{ID: "player1", Name: "Charlie"}},
		CreatedAt: time.Now(),
		Wins:      map[string]int{},
	}

	err := store.Save(filename)
	if err != nil {
		t.Errorf("Save should not error, got: %v", err)
	}

	// Verify temp file was cleaned up
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("Expected temp file to be removed after save")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	// Create store with rooms
	store1 := NewStore()
	store1.rooms["ABCD"] = &Room{
		ID:          "ABCD",
		Players:     []Player{{ID: "p1", Name: "Player1"}, {ID: "p2", Name: "Player2"}},
		CreatedAt:   time.Now(),
		Wins:        map[string]int{"p1": 5, "p2": 3},
		GamesPlayed: 8,
	}
	store1.rooms["EFGH"] = &Room{
		ID:        "EFGH",
		Players:   []Player{{ID: "p3", Name: "Player3"}},
		CreatedAt: time.Now(),
		Wins:      map[string]int{},
	}

	// Save
	if err := store1.Save(filename); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load into new store
	store2 := NewStore()
	if err := store2.Load(filename); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify
	if len(store2.rooms) != 2 {
		t.Errorf("Expected 2 rooms, got %d", len(store2.rooms))
	}

	sess := store2.rooms["ABCD"]
	if sess == nil {
		t.Fatal("Room ABCD not found")
	}
	if len(sess.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(sess.Players))
	}
	if sess.Wins["p1"] != 5 {
		t.Errorf("Expected p1 to have 5 wins, got %d", sess.Wins["p1"])
	}
	if sess.GamesPlayed != 8 {
		t.Errorf("Expected 8 games played, got %d", sess.GamesPlayed)
	}
}

func TestStartAutoSave_SavesOnStop(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	store := NewStore()
	store.rooms["TEST"] = &Room{
		ID:        "TEST",
		Players:   []Player{{ID: "player1", Name: "Dana"}},
		CreatedAt: time.Now(),
		Wins:      map[string]int{},
	}

	// Start auto-save and immediately stop it
	stop := store.StartAutoSave(filename, config.DefaultConfig().AutoSaveInterval)
	close(stop)

	// Give it a moment to complete the final save
	time.Sleep(50 * time.Millisecond)

	// Verify file was saved
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Expected file to be saved on stop")
	}

	// Verify content
	store2 := NewStore()
	if err := store2.Load(filename); err != nil {
		t.Fatalf("Failed to load saved data: %v", err)
	}
	if store2.rooms["TEST"] == nil {
		t.Error("Expected room TEST to be persisted")
	}
}

func TestCleanupStaleRooms_RemovesOldRooms(t *testing.T) {
	store := NewStore()
	now := time.Now()

	// Create a stale room (2 days old)
	store.rooms["STALE"] = &Room{
		ID:             "STALE",
		Players:        []Player{{ID: "p1", Name: "Old"}},
		CreatedAt:      now.Add(-48 * time.Hour),
		LastActivityAt: now.Add(-48 * time.Hour),
		Wins:           map[string]int{},
	}

	// Create a recent room (1 hour old)
	store.rooms["RECENT"] = &Room{
		ID:             "RECENT",
		Players:        []Player{{ID: "p2", Name: "New"}},
		CreatedAt:      now.Add(-1 * time.Hour),
		LastActivityAt: now.Add(-1 * time.Hour),
		Wins:           map[string]int{},
	}

	// Cleanup rooms older than 24 hours
	removed := store.CleanupStaleRooms(24 * time.Hour)

	if removed != 1 {
		t.Errorf("Expected 1 room removed, got %d", removed)
	}
	if len(store.rooms) != 1 {
		t.Errorf("Expected 1 room remaining, got %d", len(store.rooms))
	}
	if store.rooms["STALE"] != nil {
		t.Error("Expected STALE room to be removed")
	}
	if store.rooms["RECENT"] == nil {
		t.Error("Expected RECENT room to remain")
	}
}

func TestCleanupStaleRooms_KeepsAllRecentRooms(t *testing.T) {
	store := NewStore()
	now := time.Now()

	// Create two recent rooms
	store.rooms["A"] = &Room{
		ID:             "A",
		LastActivityAt: now.Add(-1 * time.Hour),
		Wins:           map[string]int{},
	}
	store.rooms["B"] = &Room{
		ID:             "B",
		LastActivityAt: now.Add(-12 * time.Hour),
		Wins:           map[string]int{},
	}

	removed := store.CleanupStaleRooms(24 * time.Hour)

	if removed != 0 {
		t.Errorf("Expected 0 rooms removed, got %d", removed)
	}
	if len(store.rooms) != 2 {
		t.Errorf("Expected 2 rooms remaining, got %d", len(store.rooms))
	}
}

func TestCleanupStaleRooms_RemovesAllStaleRooms(t *testing.T) {
	store := NewStore()
	now := time.Now()

	// Create three stale rooms
	store.rooms["A"] = &Room{
		ID:             "A",
		LastActivityAt: now.Add(-25 * time.Hour),
		Wins:           map[string]int{},
	}
	store.rooms["B"] = &Room{
		ID:             "B",
		LastActivityAt: now.Add(-48 * time.Hour),
		Wins:           map[string]int{},
	}
	store.rooms["C"] = &Room{
		ID:             "C",
		LastActivityAt: now.Add(-72 * time.Hour),
		Wins:           map[string]int{},
	}

	removed := store.CleanupStaleRooms(24 * time.Hour)

	if removed != 3 {
		t.Errorf("Expected 3 rooms removed, got %d", removed)
	}
	if len(store.rooms) != 0 {
		t.Errorf("Expected 0 rooms remaining, got %d", len(store.rooms))
	}
}

func TestLoad_InitializesZeroLastActivityAt(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	createdAt := time.Now().Add(-1 * time.Hour)

	// Create room data with zero LastActivityAt (simulates old data)
	pd := persistedData{
		Rooms: map[string]*Room{
			"TEST": {
				ID:             "TEST",
				Players:        []Player{{ID: "player1", Name: "Alice"}},
				CreatedAt:      createdAt,
				LastActivityAt: time.Time{}, // Zero value
				Wins:           map[string]int{},
			},
		},
		SavedAt: time.Now(),
		Version: 1,
	}
	data, _ := json.Marshal(pd)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	store := NewStore()
	err := store.Load(filename)
	if err != nil {
		t.Errorf("Load should not error, got: %v", err)
	}

	sess := store.rooms["TEST"]
	if sess.LastActivityAt.IsZero() {
		t.Error("Expected LastActivityAt to be initialized")
	}
	if !sess.LastActivityAt.Equal(createdAt) {
		t.Errorf("Expected LastActivityAt to equal CreatedAt, got %v vs %v", sess.LastActivityAt, createdAt)
	}
}

func TestCreate_SetsLastActivityAt(t *testing.T) {
	store := NewStore()
	before := time.Now()

	sess := store.Create("TestPlayer")

	after := time.Now()

	if sess.LastActivityAt.Before(before) || sess.LastActivityAt.After(after) {
		t.Errorf("LastActivityAt should be between %v and %v, got %v", before, after, sess.LastActivityAt)
	}
	if !sess.LastActivityAt.Equal(sess.CreatedAt) {
		t.Error("LastActivityAt should equal CreatedAt for new rooms")
	}
}

func TestJoin_UpdatesLastActivityAt(t *testing.T) {
	store := NewStore()
	sess := store.Create("Player1")
	originalActivity := sess.LastActivityAt

	// Small delay to ensure time difference
	time.Sleep(10 * time.Millisecond)

	sess, err := store.Join(sess.ID, "Player2")
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	if !sess.LastActivityAt.After(originalActivity) {
		t.Error("LastActivityAt should be updated after Join")
	}
}

func TestSaveAndLoad_WithActiveGame(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	// Create store with a room that has an active game
	store1 := NewStore()
	sess := store1.Create("Player1")
	store1.Join(sess.ID, "Player2")

	// Start a game
	sess, err := store1.StartGame(sess.ID, true) // use fixed board for deterministic test
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	if sess.CurrentGame == nil {
		t.Fatal("Expected CurrentGame to be set after StartGame")
	}

	originalGame := sess.CurrentGame

	// Save
	if err := store1.Save(filename); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load into new store
	store2 := NewStore()
	if err := store2.Load(filename); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify room was restored
	restoredSess := store2.rooms[sess.ID]
	if restoredSess == nil {
		t.Fatalf("Room %s not found after load", sess.ID)
	}

	// Verify game was restored
	if restoredSess.CurrentGame == nil {
		t.Fatal("Expected CurrentGame to be restored")
	}

	// Use Game.Equals for complete comparison (board including walls, bots, target)
	if !restoredSess.CurrentGame.Equals(originalGame) {
		t.Errorf("Game mismatch after persistence round-trip:\noriginal:\n%s\nrestored:\n%s",
			originalGame.String(), restoredSess.CurrentGame.String())
	}
}
