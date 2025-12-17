package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_NonExistentFile(t *testing.T) {
	store := NewStore()
	err := store.Load("/nonexistent/path/sessions.json")
	if err != nil {
		t.Errorf("Load should not error on non-existent file, got: %v", err)
	}
	if len(store.sessions) != 0 {
		t.Errorf("Expected empty sessions, got %d", len(store.sessions))
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "sessions.json")

	// Create empty file
	if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	store := NewStore()
	err := store.Load(filename)
	if err != nil {
		t.Errorf("Load should not error on empty file, got: %v", err)
	}
	if len(store.sessions) != 0 {
		t.Errorf("Expected empty sessions, got %d", len(store.sessions))
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "sessions.json")

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
	filename := filepath.Join(tmpDir, "sessions.json")

	// Create valid session data
	pd := persistedData{
		Sessions: map[string]*Session{
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
	if len(store.sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(store.sessions))
	}
	if store.sessions["TEST"] == nil {
		t.Error("Expected session 'TEST' to exist")
	}
	if store.sessions["TEST"].Players[0].Name != "Alice" {
		t.Errorf("Expected player name 'Alice', got '%s'", store.sessions["TEST"].Players[0].Name)
	}
	if store.sessions["TEST"].Wins["player1"] != 3 {
		t.Errorf("Expected 3 wins, got %d", store.sessions["TEST"].Wins["player1"])
	}
}

func TestLoad_InitializesNilWins(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "sessions.json")

	// Create session data with nil Wins map
	pd := persistedData{
		Sessions: map[string]*Session{
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
	if store.sessions["TEST"].Wins == nil {
		t.Error("Expected Wins map to be initialized, got nil")
	}
}

func TestSave_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "sessions.json")

	store := NewStore()
	store.sessions["TEST"] = &Session{
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

	if len(pd.Sessions) != 1 {
		t.Errorf("Expected 1 session in saved data, got %d", len(pd.Sessions))
	}
	if pd.Sessions["TEST"].Players[0].Name != "Bob" {
		t.Errorf("Expected player name 'Bob', got '%s'", pd.Sessions["TEST"].Players[0].Name)
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "sessions.json")
	tmpFile := filename + ".tmp"

	store := NewStore()
	store.sessions["TEST"] = &Session{
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
	filename := filepath.Join(tmpDir, "sessions.json")

	// Create store with sessions
	store1 := NewStore()
	store1.sessions["ABCD"] = &Session{
		ID:          "ABCD",
		Players:     []Player{{ID: "p1", Name: "Player1"}, {ID: "p2", Name: "Player2"}},
		CreatedAt:   time.Now(),
		Wins:        map[string]int{"p1": 5, "p2": 3},
		GamesPlayed: 8,
	}
	store1.sessions["EFGH"] = &Session{
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
	if len(store2.sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(store2.sessions))
	}

	sess := store2.sessions["ABCD"]
	if sess == nil {
		t.Fatal("Session ABCD not found")
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
	filename := filepath.Join(tmpDir, "sessions.json")

	store := NewStore()
	store.sessions["TEST"] = &Session{
		ID:        "TEST",
		Players:   []Player{{ID: "player1", Name: "Dana"}},
		CreatedAt: time.Now(),
		Wins:      map[string]int{},
	}

	// Start auto-save and immediately stop it
	stop := store.StartAutoSave(filename, DefaultAutoSaveInterval)
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
	if store2.sessions["TEST"] == nil {
		t.Error("Expected session TEST to be persisted")
	}
}

func TestCleanupStaleSessions_RemovesOldSessions(t *testing.T) {
	store := NewStore()
	now := time.Now()

	// Create a stale session (2 days old)
	store.sessions["STALE"] = &Session{
		ID:             "STALE",
		Players:        []Player{{ID: "p1", Name: "Old"}},
		CreatedAt:      now.Add(-48 * time.Hour),
		LastActivityAt: now.Add(-48 * time.Hour),
		Wins:           map[string]int{},
	}

	// Create a recent session (1 hour old)
	store.sessions["RECENT"] = &Session{
		ID:             "RECENT",
		Players:        []Player{{ID: "p2", Name: "New"}},
		CreatedAt:      now.Add(-1 * time.Hour),
		LastActivityAt: now.Add(-1 * time.Hour),
		Wins:           map[string]int{},
	}

	// Cleanup sessions older than 24 hours
	removed := store.CleanupStaleSessions(24 * time.Hour)

	if removed != 1 {
		t.Errorf("Expected 1 session removed, got %d", removed)
	}
	if len(store.sessions) != 1 {
		t.Errorf("Expected 1 session remaining, got %d", len(store.sessions))
	}
	if store.sessions["STALE"] != nil {
		t.Error("Expected STALE session to be removed")
	}
	if store.sessions["RECENT"] == nil {
		t.Error("Expected RECENT session to remain")
	}
}

func TestCleanupStaleSessions_KeepsAllRecentSessions(t *testing.T) {
	store := NewStore()
	now := time.Now()

	// Create two recent sessions
	store.sessions["A"] = &Session{
		ID:             "A",
		LastActivityAt: now.Add(-1 * time.Hour),
		Wins:           map[string]int{},
	}
	store.sessions["B"] = &Session{
		ID:             "B",
		LastActivityAt: now.Add(-12 * time.Hour),
		Wins:           map[string]int{},
	}

	removed := store.CleanupStaleSessions(24 * time.Hour)

	if removed != 0 {
		t.Errorf("Expected 0 sessions removed, got %d", removed)
	}
	if len(store.sessions) != 2 {
		t.Errorf("Expected 2 sessions remaining, got %d", len(store.sessions))
	}
}

func TestCleanupStaleSessions_RemovesAllStaleSessions(t *testing.T) {
	store := NewStore()
	now := time.Now()

	// Create three stale sessions
	store.sessions["A"] = &Session{
		ID:             "A",
		LastActivityAt: now.Add(-25 * time.Hour),
		Wins:           map[string]int{},
	}
	store.sessions["B"] = &Session{
		ID:             "B",
		LastActivityAt: now.Add(-48 * time.Hour),
		Wins:           map[string]int{},
	}
	store.sessions["C"] = &Session{
		ID:             "C",
		LastActivityAt: now.Add(-72 * time.Hour),
		Wins:           map[string]int{},
	}

	removed := store.CleanupStaleSessions(24 * time.Hour)

	if removed != 3 {
		t.Errorf("Expected 3 sessions removed, got %d", removed)
	}
	if len(store.sessions) != 0 {
		t.Errorf("Expected 0 sessions remaining, got %d", len(store.sessions))
	}
}

func TestLoad_InitializesZeroLastActivityAt(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "sessions.json")

	createdAt := time.Now().Add(-1 * time.Hour)

	// Create session data with zero LastActivityAt (simulates old data)
	pd := persistedData{
		Sessions: map[string]*Session{
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

	sess := store.sessions["TEST"]
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
		t.Error("LastActivityAt should equal CreatedAt for new sessions")
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
