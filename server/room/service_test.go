package room

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/model"
	"github.com/srsalisbury/bouncebot/server/config"
)

// Integration tests for RoomService - tests the full component composition

func TestService_CreateAndGet(t *testing.T) {
	svc := NewRoomService()

	room, _, _ := svc.Create("Alice", false)
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

	room, _, _ := svc.Create("Alice", false)
	proto, _, _, err := svc.Join(room.ID, "Bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(proto.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(proto.Players))
	}
	if proto.Players[1].Name != "Bob" {
		t.Errorf("expected second player name 'Bob', got '%s'", proto.Players[1].Name)
	}
}

func TestService_Join_NotFound(t *testing.T) {
	svc := NewRoomService()

	_, _, _, err := svc.Join("nonexistent", "Bob")
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestService_StartGame(t *testing.T) {
	svc := NewRoomService()

	room, _, _ := svc.Create("Alice", false)
	proto, err := svc.StartGame(room.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proto.CurrentGame == nil {
		t.Error("expected game to be set after StartGame")
	}
	if proto.GameStartedAt == nil {
		t.Error("expected GameStartedAt to be set")
	}
}

func TestService_StartGame_ThreadsMinSolutionLengthToSolveFunc(t *testing.T) {
	svc := NewRoomService()

	var solveCalls int
	var mu sync.Mutex
	svc.SetBoardSolveFunc(func(*model.Game) ([]model.BotPosition, bool) {
		mu.Lock()
		solveCalls++
		mu.Unlock()
		return make([]model.BotPosition, 8), true
	})

	room, _, aliceToken := svc.Create("Alice", false)
	if err := svc.UpdateRoomSettings(room.ID, aliceToken, RoomSettings{MinSolutionLength: 8}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	proto, err := svc.StartGame(room.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proto.CurrentGame == nil {
		t.Error("expected game to be set after StartGame")
	}

	mu.Lock()
	defer mu.Unlock()
	if solveCalls == 0 {
		t.Error("expected the configured BoardSolveFunc to be invoked when MinSolutionLength > 1")
	}
}

func TestService_StartGame_DoesNotHoldLockDuringSearch(t *testing.T) {
	svc := NewRoomService()

	unblock := make(chan struct{})
	svc.SetBoardSolveFunc(func(*model.Game) ([]model.BotPosition, bool) {
		<-unblock // block until the test says to proceed
		return make([]model.BotPosition, 1), true
	})

	room, _, aliceToken := svc.Create("Alice", false)
	if err := svc.UpdateRoomSettings(room.ID, aliceToken, RoomSettings{MinSolutionLength: 8}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		svc.StartGame(room.ID)
		close(done)
	}()

	// Give StartGame a moment to enter the (blocked) search.
	time.Sleep(20 * time.Millisecond)

	// If the room lock were still held during the search, this would block
	// until StartGame finishes. It must return promptly instead.
	readDone := make(chan struct{})
	go func() {
		svc.GetProto(room.ID)
		close(readDone)
	}()

	select {
	case <-readDone:
		// good: read completed without waiting on the search
	case <-time.After(2 * time.Second):
		t.Fatal("GetProto blocked, implying the room lock is held during board search")
	}

	close(unblock)
	<-done
}

func TestService_SubmitSolution_ValidSolution(t *testing.T) {
	svc := NewRoomService()
	mock := &mockBroadcaster{}
	svc.SetBroadcaster(mock)

	room, _, aliceToken := svc.Create("Alice", false)
	svc.StartGame(room.ID)
	// Use fixed Game1 board so validSolution() works
	room.CurrentGame = model.Game1()
	aliceID := room.Players[0].ID

	solution, err := svc.SubmitSolution(room.ID, aliceToken, validSolution())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if solution.PlayerID != aliceID {
		t.Errorf("expected player ID %s, got %s", aliceID, solution.PlayerID)
	}
	if solution.MoveCount() != 10 {
		t.Errorf("expected 10 moves, got %d", solution.MoveCount())
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

func TestService_DisconnectAndReconnect(t *testing.T) {
	svc := NewRoomService()

	room, _, _ := svc.Create("Alice", false)
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

	room, _, _ := svc.Create("Alice", false)
	_, _, _, _ = svc.Join(room.ID, "Bob")
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

func TestService_RemovePlayer_DeletesEmptyRoom(t *testing.T) {
	svc := NewRoomService()

	room, _, _ := svc.Create("Alice", false)
	roomID := room.ID
	aliceID := room.Players[0].ID

	// Must disconnect first
	svc.DisconnectPlayer(roomID, aliceID)
	svc.RemovePlayer(roomID, aliceID)

	// Room should be garbage collected
	_, err := svc.Get(roomID)
	if err == nil {
		t.Error("expected room to be deleted after last player removed")
	}
}

func TestService_RemovePlayer_KeepsRoomWithPendingPlayers(t *testing.T) {
	svc := NewRoomService()

	room, _, _ := svc.Create("Alice", false)
	roomID := room.ID
	aliceID := room.Players[0].ID

	// Start a game so new players become pending
	svc.StartGame(roomID)

	// Bob joins as a pending player
	_, _, _, _ = svc.Join(roomID, "Bob")

	// Alice disconnects and is removed
	svc.DisconnectPlayer(roomID, aliceID)
	svc.RemovePlayer(roomID, aliceID)

	// Room should still exist because Bob is pending
	room, err := svc.Get(roomID)
	if err != nil {
		t.Errorf("expected room to still exist with pending players, got error: %v", err)
	}
	if len(room.PendingPlayers) != 1 {
		t.Errorf("expected 1 pending player, got %d", len(room.PendingPlayers))
	}
}

func TestService_MarkFinishedSolving_TriggersGameEnd(t *testing.T) {
	svc := NewRoomService()
	mock := &mockBroadcaster{}
	svc.SetBroadcaster(mock)

	room, _, aliceToken := svc.Create("Alice", false)
	_, _, bobToken, _ := svc.Join(room.ID, "Bob")
	svc.StartGame(room.ID)

	// Alice finishes
	svc.MarkFinishedSolving(room.ID, aliceToken)
	if mock.gameEndedCalled {
		t.Error("game should not have ended yet")
	}

	// Bob finishes - should trigger game end
	svc.MarkFinishedSolving(room.ID, bobToken)
	if !mock.gameEndedCalled {
		t.Error("expected game to end when all players finished")
	}
}

func TestService_MarkReadyForNext_StartsNextGame(t *testing.T) {
	svc := NewRoomService()
	mock := &mockBroadcaster{}
	svc.SetBroadcaster(mock)

	room, _, aliceToken := svc.Create("Alice", false)
	_, _, bobToken, _ := svc.Join(room.ID, "Bob")

	// Alice is ready
	svc.MarkReadyForNext(room.ID, aliceToken)
	mock.gameStartedCalled = false

	// Bob is ready - should start next game
	svc.MarkReadyForNext(room.ID, bobToken)
	if !mock.gameStartedCalled {
		t.Error("expected next game to start when all players ready")
	}
}

func TestService_RemovePlayer_TriggersGameEnd(t *testing.T) {
	svc := NewRoomService()
	mock := &mockBroadcaster{}
	svc.SetBroadcaster(mock)

	room, _, aliceToken := svc.Create("Alice", false)
	_, bobID, _, _ := svc.Join(room.ID, "Bob")
	svc.StartGame(room.ID)

	// Alice marks finished
	svc.MarkFinishedSolving(room.ID, aliceToken)

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
	room, _, _ := svc1.Create("Alice", false)
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
	_, _, _ = svc.Create("Alice", false)

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

	room, _, _ := svc.Create("Alice", false)
	_, _, _, _ = svc.Join(room.ID, "Bob")
	svc.StartGame(room.ID)

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

func TestService_BootPlayer_HostCanBoot(t *testing.T) {
	svc := NewRoomService()
	mock := &mockBroadcaster{}
	svc.SetBroadcaster(mock)

	room, _, aliceToken := svc.Create("Alice", false)
	_, bobID, _, _ := svc.Join(room.ID, "Bob")
	aliceID := room.Players[0].ID

	err := svc.BootPlayer(room.ID, aliceToken, bobID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, _ = svc.Get(room.ID)
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player after boot, got %d", len(room.Players))
	}
	if room.Players[0].ID != aliceID {
		t.Errorf("expected Alice to remain, got %s", room.Players[0].Name)
	}

	// Verify player_left was broadcast
	if !mock.playerLeftCalled {
		t.Error("expected BroadcastPlayerLeft to be called")
	}
}

func TestService_BootPlayer_NonHostCannotBoot(t *testing.T) {
	svc := NewRoomService()

	room, _, _ := svc.Create("Alice", false)
	_, _, bobToken, _ := svc.Join(room.ID, "Bob")
	aliceID := room.Players[0].ID

	// Bob (non-host) tries to boot Alice
	err := svc.BootPlayer(room.ID, bobToken, aliceID)
	if err == nil {
		t.Error("expected error when non-host tries to boot")
	}

	// Verify no one was removed
	room, _ = svc.Get(room.ID)
	if len(room.Players) != 2 {
		t.Errorf("expected 2 players (no one booted), got %d", len(room.Players))
	}
}

func TestService_BootPlayer_TargetNotFound(t *testing.T) {
	svc := NewRoomService()

	room, _, aliceToken := svc.Create("Alice", false)

	err := svc.BootPlayer(room.ID, aliceToken, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent target player")
	}
}

func TestService_BootPlayer_RoomNotFound(t *testing.T) {
	svc := NewRoomService()

	err := svc.BootPlayer("nonexistent", "host", "target")
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestService_BootPlayer_HostCanBootSelf(t *testing.T) {
	svc := NewRoomService()

	room, aliceID, aliceToken := svc.Create("Alice", false)
	_, bobID, _, _ := svc.Join(room.ID, "Bob")

	// Alice (host) boots herself
	err := svc.BootPlayer(room.ID, aliceToken, aliceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, _ = svc.Get(room.ID)
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player after self-boot, got %d", len(room.Players))
	}
	// Bob should now be the host (first player)
	if room.Players[0].ID != bobID {
		t.Errorf("expected Bob to be new host, got %s", room.Players[0].Name)
	}
}

func TestService_BootPlayer_LastPlayerDeletesRoom(t *testing.T) {
	svc := NewRoomService()

	room, aliceID, aliceToken := svc.Create("Alice", false)
	roomID := room.ID

	// Alice boots herself (the only player)
	err := svc.BootPlayer(roomID, aliceToken, aliceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Room should be garbage collected
	_, err = svc.Get(roomID)
	if err == nil {
		t.Error("expected room to be deleted after last player booted")
	}
}
