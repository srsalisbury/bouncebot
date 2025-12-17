package session

import (
	"testing"
	"time"
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

func TestDisconnectPlayer(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	playerID := session.Players[0].ID

	err := store.DisconnectPlayer(session.ID, playerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	session, _ = store.Get(session.ID)
	if session.Players[0].Status != PlayerStatusDisconnected {
		t.Errorf("expected player status 'disconnected', got '%s'", session.Players[0].Status)
	}
	if session.Players[0].DisconnectedAt.IsZero() {
		t.Error("expected DisconnectedAt to be set")
	}

	// Verify timer was created
	if _, ok := store.timers[playerID]; !ok {
		t.Error("expected timer to be created for disconnected player")
	}
}

func TestDisconnectPlayer_SessionNotFound(t *testing.T) {
	store := NewStore()

	err := store.DisconnectPlayer("nonexistent", "player1")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestReconnectPlayer(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	playerID := session.Players[0].ID

	// Disconnect first
	store.DisconnectPlayer(session.ID, playerID)

	// Then reconnect
	err := store.ReconnectPlayer(session.ID, playerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	session, _ = store.Get(session.ID)
	if session.Players[0].Status != PlayerStatusConnected {
		t.Errorf("expected player status 'connected', got '%s'", session.Players[0].Status)
	}

	// Verify timer was cancelled
	if _, ok := store.timers[playerID]; ok {
		t.Error("expected timer to be cancelled after reconnect")
	}
}

func TestReconnectPlayer_SessionNotFound(t *testing.T) {
	store := NewStore()

	err := store.ReconnectPlayer("nonexistent", "player1")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestReconnectPlayer_PlayerNotFound(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")

	err := store.ReconnectPlayer(session.ID, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent player")
	}
}

func TestRemovePlayer(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	store.Join(session.ID, "Bob")
	playerID := session.Players[0].ID

	// Must disconnect first (RemovePlayer only removes disconnected players)
	store.DisconnectPlayer(session.ID, playerID)

	store.RemovePlayer(session.ID, playerID)

	session, _ = store.Get(session.ID)
	if len(session.Players) != 1 {
		t.Errorf("expected 1 player after removal, got %d", len(session.Players))
	}
	if session.Players[0].Name != "Bob" {
		t.Errorf("expected remaining player to be 'Bob', got '%s'", session.Players[0].Name)
	}
}

func TestRemovePlayer_OnlyRemovesDisconnected(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	store.Join(session.ID, "Bob")
	playerID := session.Players[0].ID

	// Try to remove without disconnecting first
	store.RemovePlayer(session.ID, playerID)

	session, _ = store.Get(session.ID)
	if len(session.Players) != 2 {
		t.Errorf("expected 2 players (connected player should not be removed), got %d", len(session.Players))
	}
}

func TestRemovePlayer_TimerFires(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	store.Join(session.ID, "Bob")
	playerID := session.Players[0].ID

	// Disconnect - this starts a 30s timer, but we'll call RemovePlayer directly
	store.DisconnectPlayer(session.ID, playerID)

	// Simulate timer firing by calling RemovePlayer
	store.RemovePlayer(session.ID, playerID)

	session, _ = store.Get(session.ID)
	if len(session.Players) != 1 {
		t.Errorf("expected 1 player after timer-triggered removal, got %d", len(session.Players))
	}

	// Timer should be cleaned up
	if _, ok := store.timers[playerID]; ok {
		t.Error("expected timer to be cleaned up after removal")
	}
}

func TestDisconnectReconnectCycle(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	playerID := session.Players[0].ID

	// Disconnect
	store.DisconnectPlayer(session.ID, playerID)
	session, _ = store.Get(session.ID)
	if session.Players[0].Status != PlayerStatusDisconnected {
		t.Error("expected disconnected status")
	}

	// Reconnect
	store.ReconnectPlayer(session.ID, playerID)
	session, _ = store.Get(session.ID)
	if session.Players[0].Status != PlayerStatusConnected {
		t.Error("expected connected status after reconnect")
	}

	// Disconnect again
	store.DisconnectPlayer(session.ID, playerID)
	session, _ = store.Get(session.ID)
	if session.Players[0].Status != PlayerStatusDisconnected {
		t.Error("expected disconnected status after second disconnect")
	}

	// Player count should still be 1
	if len(session.Players) != 1 {
		t.Errorf("expected 1 player throughout cycle, got %d", len(session.Players))
	}
}

func TestPlayerStatus_InitiallyConnected(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	if session.Players[0].Status != PlayerStatusConnected {
		t.Errorf("expected new player status 'connected', got '%s'", session.Players[0].Status)
	}

	store.Join(session.ID, "Bob")
	session, _ = store.Get(session.ID)
	if session.Players[1].Status != PlayerStatusConnected {
		t.Errorf("expected joined player status 'connected', got '%s'", session.Players[1].Status)
	}
}

func TestDisconnectPlayer_ReplacesExistingTimer(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	playerID := session.Players[0].ID

	// Disconnect twice - second should replace first timer
	store.DisconnectPlayer(session.ID, playerID)
	firstTimer := store.timers[playerID]

	// Small delay to ensure different timer
	time.Sleep(10 * time.Millisecond)

	store.DisconnectPlayer(session.ID, playerID)
	secondTimer := store.timers[playerID]

	if firstTimer == secondTimer {
		t.Error("expected second disconnect to create new timer")
	}
}

func TestRemovePlayer_CleansUpFinishedSolving(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	store.Join(session.ID, "Bob")
	store.StartGame(session.ID, false)

	aliceID := session.Players[0].ID

	// Alice marks finished
	store.MarkFinishedSolving(session.ID, aliceID)

	session, _ = store.Get(session.ID)
	if len(session.FinishedSolving) != 1 {
		t.Fatalf("expected 1 finished player, got %d", len(session.FinishedSolving))
	}

	// Alice disconnects and is removed
	store.DisconnectPlayer(session.ID, aliceID)
	store.RemovePlayer(session.ID, aliceID)

	session, _ = store.Get(session.ID)
	if len(session.FinishedSolving) != 0 {
		t.Errorf("expected 0 finished players after removal, got %d", len(session.FinishedSolving))
	}
}

func TestRemovePlayer_CleansUpReadyForNext(t *testing.T) {
	store := NewStore()

	session := store.Create("Alice")
	store.Join(session.ID, "Bob")

	aliceID := session.Players[0].ID

	// Alice marks ready
	store.MarkReadyForNext(session.ID, aliceID)

	session, _ = store.Get(session.ID)
	if len(session.ReadyForNext) != 1 {
		t.Fatalf("expected 1 ready player, got %d", len(session.ReadyForNext))
	}

	// Alice disconnects and is removed
	store.DisconnectPlayer(session.ID, aliceID)
	store.RemovePlayer(session.ID, aliceID)

	session, _ = store.Get(session.ID)
	if len(session.ReadyForNext) != 0 {
		t.Errorf("expected 0 ready players after removal, got %d", len(session.ReadyForNext))
	}
}

// mockBroadcaster implements EventBroadcaster for testing
type mockBroadcaster struct {
	gameEndedCalled   bool
	gameStartedCalled bool
}

func (m *mockBroadcaster) BroadcastPlayerJoined(sessionID, playerID, playerName string) {}
func (m *mockBroadcaster) BroadcastPlayerLeft(sessionID, playerID string)               {}
func (m *mockBroadcaster) BroadcastGameStarted(sessionID string)                         { m.gameStartedCalled = true }
func (m *mockBroadcaster) BroadcastPlayerFinishedSolving(sessionID, playerID string)     {}
func (m *mockBroadcaster) BroadcastPlayerReadyForNext(sessionID, playerID string)        {}
func (m *mockBroadcaster) BroadcastPlayerSolved(sessionID, playerID string, moveCount int) {
}
func (m *mockBroadcaster) BroadcastSolutionRetracted(sessionID, playerID string) {}
func (m *mockBroadcaster) BroadcastGameEnded(sessionID, winnerID, winnerName string, moves []MovePayload) {
	m.gameEndedCalled = true
}

func TestRemovePlayer_TriggersGameEnd(t *testing.T) {
	store := NewStore()
	mock := &mockBroadcaster{}
	store.SetBroadcaster(mock)

	session := store.Create("Alice")
	store.Join(session.ID, "Bob")
	store.StartGame(session.ID, false)

	aliceID := session.Players[0].ID
	bobID := session.Players[1].ID

	// Alice marks finished
	store.MarkFinishedSolving(session.ID, aliceID)

	// Game should not have ended yet (Bob hasn't finished)
	if mock.gameEndedCalled {
		t.Error("game should not have ended yet")
	}

	// Bob disconnects and is removed
	store.DisconnectPlayer(session.ID, bobID)
	store.RemovePlayer(session.ID, bobID)

	// Now game should end (Alice is the only player and she's finished)
	if !mock.gameEndedCalled {
		t.Error("expected game to end when last unfinished player was removed")
	}
}

func TestRemovePlayer_TriggersNextGame(t *testing.T) {
	store := NewStore()
	mock := &mockBroadcaster{}
	store.SetBroadcaster(mock)

	session := store.Create("Alice")
	store.Join(session.ID, "Bob")

	aliceID := session.Players[0].ID
	bobID := session.Players[1].ID

	// Alice marks ready for next
	store.MarkReadyForNext(session.ID, aliceID)

	// Reset the flag (it might have been called during setup)
	mock.gameStartedCalled = false

	// Bob disconnects and is removed
	store.DisconnectPlayer(session.ID, bobID)
	store.RemovePlayer(session.ID, bobID)

	// Now next game should start (Alice is the only player and she's ready)
	if !mock.gameStartedCalled {
		t.Error("expected next game to start when last unready player was removed")
	}
}
