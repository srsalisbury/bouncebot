package room

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistenceManager_Load_NonExistentFile(t *testing.T) {
	pm := NewPersistenceManager()

	rooms, err := pm.Load("/nonexistent/path/rooms.json")
	if err != nil {
		t.Errorf("Load should not error on non-existent file, got: %v", err)
	}
	if len(rooms) != 0 {
		t.Errorf("expected empty rooms map, got %d", len(rooms))
	}
}

func TestPersistenceManager_Load_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	pm := NewPersistenceManager()

	rooms, err := pm.Load(filename)
	if err != nil {
		t.Errorf("Load should not error on empty file, got: %v", err)
	}
	if len(rooms) != 0 {
		t.Errorf("expected empty rooms map, got %d", len(rooms))
	}
}

func TestPersistenceManager_Load_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	if err := os.WriteFile(filename, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	pm := NewPersistenceManager()

	_, err := pm.Load(filename)
	if err == nil {
		t.Error("Load should error on invalid JSON")
	}
}

func TestPersistenceManager_Load_ValidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

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
		t.Fatalf("failed to create test file: %v", err)
	}

	pm := NewPersistenceManager()

	rooms, err := pm.Load(filename)
	if err != nil {
		t.Errorf("Load should not error on valid JSON, got: %v", err)
	}
	if len(rooms) != 1 {
		t.Errorf("expected 1 room, got %d", len(rooms))
	}
	if rooms["TEST"] == nil {
		t.Error("expected room 'TEST' to exist")
	}
	if rooms["TEST"].Players[0].Name != "Alice" {
		t.Errorf("expected player name 'Alice', got '%s'", rooms["TEST"].Players[0].Name)
	}
	if rooms["TEST"].Wins["player1"] != 3 {
		t.Errorf("expected 3 wins, got %d", rooms["TEST"].Wins["player1"])
	}
}

func TestPersistenceManager_Load_InitializesNilWins(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	pd := persistedData{
		Rooms: map[string]*Room{
			"TEST": {
				ID:        "TEST",
				Players:   []Player{{ID: "player1", Name: "Alice"}},
				CreatedAt: time.Now(),
				Wins:      nil,
			},
		},
		SavedAt: time.Now(),
		Version: 1,
	}
	data, _ := json.Marshal(pd)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	pm := NewPersistenceManager()

	rooms, err := pm.Load(filename)
	if err != nil {
		t.Errorf("Load should not error, got: %v", err)
	}
	if rooms["TEST"].Wins == nil {
		t.Error("expected Wins map to be initialized, got nil")
	}
}

func TestPersistenceManager_Load_InitializesZeroLastActivityAt(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	createdAt := time.Now().Add(-1 * time.Hour)

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
		t.Fatalf("failed to create test file: %v", err)
	}

	pm := NewPersistenceManager()

	rooms, err := pm.Load(filename)
	if err != nil {
		t.Errorf("Load should not error, got: %v", err)
	}
	if rooms["TEST"].LastActivityAt.IsZero() {
		t.Error("expected LastActivityAt to be initialized")
	}
	if !rooms["TEST"].LastActivityAt.Equal(createdAt) {
		t.Errorf("expected LastActivityAt to equal CreatedAt, got %v vs %v",
			rooms["TEST"].LastActivityAt, createdAt)
	}
}

func TestPersistenceManager_Save_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	pm := NewPersistenceManager()

	rooms := map[string]*Room{
		"TEST": {
			ID:        "TEST",
			Players:   []Player{{ID: "player1", Name: "Bob"}},
			CreatedAt: time.Now(),
			Wins:      map[string]int{},
		},
	}

	err := pm.Save(filename, rooms)
	if err != nil {
		t.Errorf("Save should not error, got: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}

	// Verify content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	var pd persistedData
	if err := json.Unmarshal(data, &pd); err != nil {
		t.Fatalf("failed to parse saved JSON: %v", err)
	}

	if len(pd.Rooms) != 1 {
		t.Errorf("expected 1 room in saved data, got %d", len(pd.Rooms))
	}
	if pd.Rooms["TEST"].Players[0].Name != "Bob" {
		t.Errorf("expected player name 'Bob', got '%s'", pd.Rooms["TEST"].Players[0].Name)
	}
}

func TestPersistenceManager_Save_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")
	tmpFile := filename + ".tmp"

	pm := NewPersistenceManager()

	rooms := map[string]*Room{
		"TEST": {
			ID:        "TEST",
			Players:   []Player{{ID: "player1", Name: "Charlie"}},
			CreatedAt: time.Now(),
			Wins:      map[string]int{},
		},
	}

	err := pm.Save(filename, rooms)
	if err != nil {
		t.Errorf("Save should not error, got: %v", err)
	}

	// Verify temp file was cleaned up
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("expected temp file to be removed after save")
	}
}

func TestPersistenceManager_SaveAndLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	pm := NewPersistenceManager()

	original := map[string]*Room{
		"ABCD": {
			ID:          "ABCD",
			Players:     []Player{{ID: "p1", Name: "Player1"}, {ID: "p2", Name: "Player2"}},
			CreatedAt:   time.Now(),
			Wins:        map[string]int{"p1": 5, "p2": 3},
			GamesPlayed: 8,
		},
		"EFGH": {
			ID:        "EFGH",
			Players:   []Player{{ID: "p3", Name: "Player3"}},
			CreatedAt: time.Now(),
			Wins:      map[string]int{},
		},
	}

	// Save
	if err := pm.Save(filename, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := pm.Load(filename)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify
	if len(loaded) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(loaded))
	}

	room := loaded["ABCD"]
	if room == nil {
		t.Fatal("Room ABCD not found")
	}
	if len(room.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(room.Players))
	}
	if room.Wins["p1"] != 5 {
		t.Errorf("expected p1 to have 5 wins, got %d", room.Wins["p1"])
	}
	if room.GamesPlayed != 8 {
		t.Errorf("expected 8 games played, got %d", room.GamesPlayed)
	}
}

func TestPersistenceManager_FindStaleRooms_RemovesOldRooms(t *testing.T) {
	pm := NewPersistenceManager()
	now := time.Now()

	rooms := map[string]*Room{
		"STALE": {
			ID:             "STALE",
			LastActivityAt: now.Add(-48 * time.Hour),
		},
		"RECENT": {
			ID:             "RECENT",
			LastActivityAt: now.Add(-1 * time.Hour),
		},
	}

	stale := pm.FindStaleRooms(rooms, 24*time.Hour)

	if len(stale) != 1 {
		t.Errorf("expected 1 stale room, got %d", len(stale))
	}
	if stale[0] != "STALE" {
		t.Errorf("expected STALE room to be marked stale, got %s", stale[0])
	}
}

func TestPersistenceManager_FindStaleRooms_KeepsAllRecentRooms(t *testing.T) {
	pm := NewPersistenceManager()
	now := time.Now()

	rooms := map[string]*Room{
		"A": {ID: "A", LastActivityAt: now.Add(-1 * time.Hour)},
		"B": {ID: "B", LastActivityAt: now.Add(-12 * time.Hour)},
	}

	stale := pm.FindStaleRooms(rooms, 24*time.Hour)

	if len(stale) != 0 {
		t.Errorf("expected 0 stale rooms, got %d", len(stale))
	}
}

func TestPersistenceManager_FindStaleRooms_RemovesAllStaleRooms(t *testing.T) {
	pm := NewPersistenceManager()
	now := time.Now()

	rooms := map[string]*Room{
		"A": {ID: "A", LastActivityAt: now.Add(-25 * time.Hour)},
		"B": {ID: "B", LastActivityAt: now.Add(-48 * time.Hour)},
		"C": {ID: "C", LastActivityAt: now.Add(-72 * time.Hour)},
	}

	stale := pm.FindStaleRooms(rooms, 24*time.Hour)

	if len(stale) != 3 {
		t.Errorf("expected 3 stale rooms, got %d", len(stale))
	}
}

func TestPersistenceManager_FindStaleRooms_EmptyMap(t *testing.T) {
	pm := NewPersistenceManager()

	stale := pm.FindStaleRooms(map[string]*Room{}, 24*time.Hour)

	if len(stale) != 0 {
		t.Errorf("expected 0 stale rooms for empty map, got %d", len(stale))
	}
}
