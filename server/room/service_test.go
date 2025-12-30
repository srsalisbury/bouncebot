package room

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/server/config"
)

// Integration tests for RoomService - tests the full component composition

func TestService_CreateAndGet(t *testing.T) {
	svc := NewRoomService()

	room := svc.Create("Alice")
	if room.ID == "" {
		t.Error("expected room ID to be set")
	}
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(room.Players))
	}
	if room.Players[0].Name != "Alice" {
		t.Errorf("expected player name 'Alice', got '%s'", room.Players[0].Name)
	}

	retrieved, err := svc.Get(room.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != room.ID {
		t.Errorf("expected room ID '%s', got '%s'", room.ID, retrieved.ID)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	svc := NewRoomService()

	_, err := svc.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestService_Join(t *testing.T) {
	svc := NewRoomService()

	room := svc.Create("Alice")
	room, err := svc.Join(room.ID, "Bob")
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

func TestService_Join_NotFound(t *testing.T) {
	svc := NewRoomService()

	_, err := svc.Join("nonexistent", "Bob")
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestService_StartGame(t *testing.T) {
	svc := NewRoomService()

	room := svc.Create("Alice")
	room, err := svc.StartGame(room.ID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.CurrentGame == nil {
		t.Error("expected game to be set after StartGame")
	}
	if room.GameStartedAt == nil {
		t.Error("expected GameStartedAt to be set")
	}
}

func TestService_SubmitSolution_ValidSolution(t *testing.T) {
	svc := NewRoomService()
	mock := &mockBroadcaster{}
	svc.SetBroadcaster(mock)

	room := svc.Create("Alice")
	svc.StartGame(room.ID, true) // Use fixed board
	aliceID := room.Players[0].ID

	solution, err := svc.SubmitSolution(room.ID, aliceID, validSolution())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if solution.PlayerID != aliceID {
		t.Errorf("expected player ID %s, got %s", aliceID, solution.PlayerID)
	}
	if solution.MoveCount() != 7 {
		t.Errorf("expected 7 moves, got %d", solution.MoveCount())
	}

	// Check room has solution
	room, _ = svc.Get(room.ID)
	if len(room.Solutions) != 1 {
		t.Errorf("expected 1 solution, got %d", len(room.Solutions))
	}

	// Check broadcast was called
	if !mock.playerSolvedCalled {
		t.Error("expected BroadcastPlayerSolved to be called")
	}
}

func TestService_RetractSolution(t *testing.T) {
	svc := NewRoomService()
	mock := &mockBroadcaster{}
	svc.SetBroadcaster(mock)

	room := svc.Create("Alice")
	svc.StartGame(room.ID, true)
	aliceID := room.Players[0].ID

	svc.SubmitSolution(room.ID, aliceID, validSolution())

	err := svc.RetractSolution(room.ID, aliceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, _ = svc.Get(room.ID)
	if len(room.Solutions) != 0 {
		t.Errorf("expected 0 solutions after retraction, got %d", len(room.Solutions))
	}

	if !mock.solutionRetractedCalled {
		t.Error("expected BroadcastSolutionRetracted to be called")
	}
}

func TestService_DisconnectAndReconnect(t *testing.T) {
	svc := NewRoomService()

	room := svc.Create("Alice")
	aliceID := room.Players[0].ID

	// Disconnect
	err := svc.DisconnectPlayer(room.ID, aliceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, _ = svc.Get(room.ID)
	if room.Players[0].Status != PlayerStatusDisconnected {
		t.Errorf("expected player status 'disconnected', got '%s'", room.Players[0].Status)
	}

	// Verify timer was started
	if !svc.hasTimer(aliceID) {
		t.Error("expected timer to be created for disconnected player")
	}

	// Reconnect
	err = svc.ReconnectPlayer(room.ID, aliceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, _ = svc.Get(room.ID)
	if room.Players[0].Status != PlayerStatusConnected {
		t.Errorf("expected player status 'connected', got '%s'", room.Players[0].Status)
	}

	// Verify timer was cancelled
	if svc.hasTimer(aliceID) {
		t.Error("expected timer to be cancelled after reconnect")
	}
}

func TestService_RemovePlayer(t *testing.T) {
	svc := NewRoomService()

	room := svc.Create("Alice")
	svc.Join(room.ID, "Bob")
	aliceID := room.Players[0].ID

	// Must disconnect first
	svc.DisconnectPlayer(room.ID, aliceID)
	svc.RemovePlayer(room.ID, aliceID)

	room, _ = svc.Get(room.ID)
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player after removal, got %d", len(room.Players))
	}
	if room.Players[0].Name != "Bob" {
		t.Errorf("expected remaining player to be 'Bob', got '%s'", room.Players[0].Name)
	}
}

func TestService_MarkFinishedSolving_TriggersGameEnd(t *testing.T) {
	svc := NewRoomService()
	mock := &mockBroadcaster{}
	svc.SetBroadcaster(mock)

	room := svc.Create("Alice")
	svc.Join(room.ID, "Bob")
	svc.StartGame(room.ID, true)

	aliceID := room.Players[0].ID
	bobID := room.Players[1].ID

	// Alice finishes
	svc.MarkFinishedSolving(room.ID, aliceID)
	if mock.gameEndedCalled {
		t.Error("game should not have ended yet")
	}

	// Bob finishes - should trigger game end
	svc.MarkFinishedSolving(room.ID, bobID)
	if !mock.gameEndedCalled {
		t.Error("expected game to end when all players finished")
	}
}

func TestService_MarkReadyForNext_StartsNextGame(t *testing.T) {
	svc := NewRoomService()
	mock := &mockBroadcaster{}
	svc.SetBroadcaster(mock)

	room := svc.Create("Alice")
	svc.Join(room.ID, "Bob")

	aliceID := room.Players[0].ID
	bobID := room.Players[1].ID

	// Alice is ready
	svc.MarkReadyForNext(room.ID, aliceID)
	mock.gameStartedCalled = false

	// Bob is ready - should start next game
	svc.MarkReadyForNext(room.ID, bobID)
	if !mock.gameStartedCalled {
		t.Error("expected next game to start when all players ready")
	}
}

func TestService_RemovePlayer_TriggersGameEnd(t *testing.T) {
	svc := NewRoomService()
	mock := &mockBroadcaster{}
	svc.SetBroadcaster(mock)

	room := svc.Create("Alice")
	svc.Join(room.ID, "Bob")
	svc.StartGame(room.ID, false)

	aliceID := room.Players[0].ID
	bobID := room.Players[1].ID

	// Alice marks finished
	svc.MarkFinishedSolving(room.ID, aliceID)

	// Bob disconnects and is removed
	svc.DisconnectPlayer(room.ID, bobID)
	svc.RemovePlayer(room.ID, bobID)

	// Game should end (Alice is the only player and she's finished)
	if !mock.gameEndedCalled {
		t.Error("expected game to end when last unfinished player was removed")
	}
}

func TestService_Persistence_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	svc1 := NewRoomService()
	room := svc1.Create("Alice")
	svc1.Join(room.ID, "Bob")

	// Save
	if err := svc1.Save(filename); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load into new service
	svc2 := NewRoomService()
	if err := svc2.Load(filename); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	loaded, err := svc2.Get(room.ID)
	if err != nil {
		t.Fatalf("failed to get room after load: %v", err)
	}
	if len(loaded.Players) != 2 {
		t.Errorf("expected 2 players after load, got %d", len(loaded.Players))
	}
}

func TestService_StartAutoSave_SavesOnStop(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "rooms.json")

	svc := NewRoomService()
	svc.Create("Alice")

	// Start auto-save and immediately stop
	stop := svc.StartAutoSave(filename, config.DefaultConfig().AutoSaveInterval)
	close(stop)

	// Give it a moment
	time.Sleep(50 * time.Millisecond)

	// Load into new service to verify
	svc2 := NewRoomService()
	if err := svc2.Load(filename); err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if len(svc2.rooms()) != 1 {
		t.Errorf("expected 1 room to be saved, got %d", len(svc2.rooms()))
	}
}

func TestService_CleanupStaleRooms(t *testing.T) {
	svc := NewRoomService()

	// Create a stale room
	svc.setRoom("STALE", &Room{
		ID:             "STALE",
		LastActivityAt: time.Now().Add(-48 * time.Hour),
		Wins:           map[string]int{},
	})

	// Create a recent room
	svc.setRoom("RECENT", &Room{
		ID:             "RECENT",
		LastActivityAt: time.Now().Add(-1 * time.Hour),
		Wins:           map[string]int{},
	})

	removed := svc.CleanupStaleRooms(24 * time.Hour)

	if removed != 1 {
		t.Errorf("expected 1 room removed, got %d", removed)
	}
	if len(svc.rooms()) != 1 {
		t.Errorf("expected 1 room remaining, got %d", len(svc.rooms()))
	}
	if _, err := svc.Get("STALE"); err == nil {
		t.Error("expected STALE room to be removed")
	}
	if _, err := svc.Get("RECENT"); err != nil {
		t.Error("expected RECENT room to remain")
	}
}

func TestService_ToProto(t *testing.T) {
	svc := NewRoomService()

	room := svc.Create("Alice")
	svc.Join(room.ID, "Bob")
	svc.StartGame(room.ID, false)

	room, _ = svc.Get(room.ID)
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
