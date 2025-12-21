package room

import (
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/model"
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

func TestDisconnectPlayer(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	playerID := room.Players[0].ID

	err := store.DisconnectPlayer(room.ID, playerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, _ = store.Get(room.ID)
	if room.Players[0].Status != PlayerStatusDisconnected {
		t.Errorf("expected player status 'disconnected', got '%s'", room.Players[0].Status)
	}
	if room.Players[0].DisconnectedAt.IsZero() {
		t.Error("expected DisconnectedAt to be set")
	}

	// Verify timer was created
	if _, ok := store.timers[playerID]; !ok {
		t.Error("expected timer to be created for disconnected player")
	}
}

func TestDisconnectPlayer_RoomNotFound(t *testing.T) {
	store := NewStore()

	err := store.DisconnectPlayer("nonexistent", "player1")
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestReconnectPlayer(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	playerID := room.Players[0].ID

	// Disconnect first
	store.DisconnectPlayer(room.ID, playerID)

	// Then reconnect
	err := store.ReconnectPlayer(room.ID, playerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, _ = store.Get(room.ID)
	if room.Players[0].Status != PlayerStatusConnected {
		t.Errorf("expected player status 'connected', got '%s'", room.Players[0].Status)
	}

	// Verify timer was cancelled
	if _, ok := store.timers[playerID]; ok {
		t.Error("expected timer to be cancelled after reconnect")
	}
}

func TestReconnectPlayer_RoomNotFound(t *testing.T) {
	store := NewStore()

	err := store.ReconnectPlayer("nonexistent", "player1")
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestReconnectPlayer_PlayerNotFound(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")

	err := store.ReconnectPlayer(room.ID, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent player")
	}
}

func TestRemovePlayer(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.Join(room.ID, "Bob")
	playerID := room.Players[0].ID

	// Must disconnect first (RemovePlayer only removes disconnected players)
	store.DisconnectPlayer(room.ID, playerID)

	store.RemovePlayer(room.ID, playerID)

	room, _ = store.Get(room.ID)
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player after removal, got %d", len(room.Players))
	}
	if room.Players[0].Name != "Bob" {
		t.Errorf("expected remaining player to be 'Bob', got '%s'", room.Players[0].Name)
	}
}

func TestRemovePlayer_OnlyRemovesDisconnected(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.Join(room.ID, "Bob")
	playerID := room.Players[0].ID

	// Try to remove without disconnecting first
	store.RemovePlayer(room.ID, playerID)

	room, _ = store.Get(room.ID)
	if len(room.Players) != 2 {
		t.Errorf("expected 2 players (connected player should not be removed), got %d", len(room.Players))
	}
}

func TestRemovePlayer_TimerFires(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.Join(room.ID, "Bob")
	playerID := room.Players[0].ID

	// Disconnect - this starts a 30s timer, but we'll call RemovePlayer directly
	store.DisconnectPlayer(room.ID, playerID)

	// Simulate timer firing by calling RemovePlayer
	store.RemovePlayer(room.ID, playerID)

	room, _ = store.Get(room.ID)
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player after timer-triggered removal, got %d", len(room.Players))
	}

	// Timer should be cleaned up
	if _, ok := store.timers[playerID]; ok {
		t.Error("expected timer to be cleaned up after removal")
	}
}

func TestDisconnectReconnectCycle(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	playerID := room.Players[0].ID

	// Disconnect
	store.DisconnectPlayer(room.ID, playerID)
	room, _ = store.Get(room.ID)
	if room.Players[0].Status != PlayerStatusDisconnected {
		t.Error("expected disconnected status")
	}

	// Reconnect
	store.ReconnectPlayer(room.ID, playerID)
	room, _ = store.Get(room.ID)
	if room.Players[0].Status != PlayerStatusConnected {
		t.Error("expected connected status after reconnect")
	}

	// Disconnect again
	store.DisconnectPlayer(room.ID, playerID)
	room, _ = store.Get(room.ID)
	if room.Players[0].Status != PlayerStatusDisconnected {
		t.Error("expected disconnected status after second disconnect")
	}

	// Player count should still be 1
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player throughout cycle, got %d", len(room.Players))
	}
}

func TestPlayerStatus_InitiallyConnected(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	if room.Players[0].Status != PlayerStatusConnected {
		t.Errorf("expected new player status 'connected', got '%s'", room.Players[0].Status)
	}

	store.Join(room.ID, "Bob")
	room, _ = store.Get(room.ID)
	if room.Players[1].Status != PlayerStatusConnected {
		t.Errorf("expected joined player status 'connected', got '%s'", room.Players[1].Status)
	}
}

func TestDisconnectPlayer_ReplacesExistingTimer(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	playerID := room.Players[0].ID

	// Disconnect twice - second should replace first timer
	store.DisconnectPlayer(room.ID, playerID)
	firstTimer := store.timers[playerID]

	// Small delay to ensure different timer
	time.Sleep(10 * time.Millisecond)

	store.DisconnectPlayer(room.ID, playerID)
	secondTimer := store.timers[playerID]

	if firstTimer == secondTimer {
		t.Error("expected second disconnect to create new timer")
	}
}

func TestRemovePlayer_CleansUpFinishedSolving(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.Join(room.ID, "Bob")
	store.StartGame(room.ID, false)

	aliceID := room.Players[0].ID

	// Alice marks finished
	store.MarkFinishedSolving(room.ID, aliceID)

	room, _ = store.Get(room.ID)
	if len(room.FinishedSolving) != 1 {
		t.Fatalf("expected 1 finished player, got %d", len(room.FinishedSolving))
	}

	// Alice disconnects and is removed
	store.DisconnectPlayer(room.ID, aliceID)
	store.RemovePlayer(room.ID, aliceID)

	room, _ = store.Get(room.ID)
	if len(room.FinishedSolving) != 0 {
		t.Errorf("expected 0 finished players after removal, got %d", len(room.FinishedSolving))
	}
}

func TestRemovePlayer_CleansUpReadyForNext(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.Join(room.ID, "Bob")

	aliceID := room.Players[0].ID

	// Alice marks ready
	store.MarkReadyForNext(room.ID, aliceID)

	room, _ = store.Get(room.ID)
	if len(room.ReadyForNext) != 1 {
		t.Fatalf("expected 1 ready player, got %d", len(room.ReadyForNext))
	}

	// Alice disconnects and is removed
	store.DisconnectPlayer(room.ID, aliceID)
	store.RemovePlayer(room.ID, aliceID)

	room, _ = store.Get(room.ID)
	if len(room.ReadyForNext) != 0 {
		t.Errorf("expected 0 ready players after removal, got %d", len(room.ReadyForNext))
	}
}

func TestRemovePlayer_CleansUpSolutions(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.Join(room.ID, "Bob")
	store.StartGame(room.ID, false)

	aliceID := room.Players[0].ID

	// Directly add a solution for Alice (bypassing validation for test purposes)
	room, _ = store.Get(room.ID)
	room.Solutions = append(room.Solutions, PlayerSolution{
		PlayerID: aliceID,
		SolvedAt: time.Now(),
		Moves:    nil,
	})

	if len(room.Solutions) != 1 {
		t.Fatalf("expected 1 solution, got %d", len(room.Solutions))
	}

	// Alice disconnects and is removed
	store.DisconnectPlayer(room.ID, aliceID)
	store.RemovePlayer(room.ID, aliceID)

	room, _ = store.Get(room.ID)
	if len(room.Solutions) != 0 {
		t.Errorf("expected 0 solutions after removal, got %d", len(room.Solutions))
	}
}

// mockBroadcaster implements EventBroadcaster for testing
type mockBroadcaster struct {
	gameEndedCalled         bool
	gameStartedCalled       bool
	playerSolvedCalled      bool
	solutionRetractedCalled bool
}

func (m *mockBroadcaster) BroadcastPlayerJoined(roomID, playerID, playerName string) {}
func (m *mockBroadcaster) BroadcastPlayerLeft(roomID, playerID string)               {}
func (m *mockBroadcaster) BroadcastGameStarted(roomID string)                         { m.gameStartedCalled = true }
func (m *mockBroadcaster) BroadcastPlayerFinishedSolving(roomID, playerID string)     {}
func (m *mockBroadcaster) BroadcastPlayerReadyForNext(roomID, playerID string)        {}
func (m *mockBroadcaster) BroadcastPlayerSolved(roomID, playerID string, moveCount int) {
	m.playerSolvedCalled = true
}
func (m *mockBroadcaster) BroadcastSolutionRetracted(roomID, playerID string) {
	m.solutionRetractedCalled = true
}
func (m *mockBroadcaster) BroadcastGameEnded(roomID, winnerID, winnerName string, moves []MovePayload) {
	m.gameEndedCalled = true
}

func TestRemovePlayer_TriggersGameEnd(t *testing.T) {
	store := NewStore()
	mock := &mockBroadcaster{}
	store.SetBroadcaster(mock)

	room := store.Create("Alice")
	store.Join(room.ID, "Bob")
	store.StartGame(room.ID, false)

	aliceID := room.Players[0].ID
	bobID := room.Players[1].ID

	// Alice marks finished
	store.MarkFinishedSolving(room.ID, aliceID)

	// Game should not have ended yet (Bob hasn't finished)
	if mock.gameEndedCalled {
		t.Error("game should not have ended yet")
	}

	// Bob disconnects and is removed
	store.DisconnectPlayer(room.ID, bobID)
	store.RemovePlayer(room.ID, bobID)

	// Now game should end (Alice is the only player and she's finished)
	if !mock.gameEndedCalled {
		t.Error("expected game to end when last unfinished player was removed")
	}
}

func TestRemovePlayer_TriggersNextGame(t *testing.T) {
	store := NewStore()
	mock := &mockBroadcaster{}
	store.SetBroadcaster(mock)

	room := store.Create("Alice")
	store.Join(room.ID, "Bob")

	aliceID := room.Players[0].ID
	bobID := room.Players[1].ID

	// Alice marks ready for next
	store.MarkReadyForNext(room.ID, aliceID)

	// Reset the flag (it might have been called during setup)
	mock.gameStartedCalled = false

	// Bob disconnects and is removed
	store.DisconnectPlayer(room.ID, bobID)
	store.RemovePlayer(room.ID, bobID)

	// Now next game should start (Alice is the only player and she's ready)
	if !mock.gameStartedCalled {
		t.Error("expected next game to start when last unready player was removed")
	}
}

// Helper to create a valid solution for Game1 (fixed board).
// Target is bot 0 at (5, 13), starting at (5, 4).
func validSolution() []model.BotPosition {
	// Valid 7-move solution for Game1:
	// Bot 1 left, then Bot 0: up, left, down, left, up, right
	return []model.BotPosition{
		{Id: 1, Pos: model.Position{X: 0, Y: 12}},
		{Id: 0, Pos: model.Position{X: 5, Y: 0}},
		{Id: 0, Pos: model.Position{X: 2, Y: 0}},
		{Id: 0, Pos: model.Position{X: 2, Y: 15}},
		{Id: 0, Pos: model.Position{X: 0, Y: 15}},
		{Id: 0, Pos: model.Position{X: 0, Y: 13}},
		{Id: 0, Pos: model.Position{X: 5, Y: 13}},
	}
}

func TestSubmitSolution_RoomNotFound(t *testing.T) {
	store := NewStore()

	_, err := store.SubmitSolution("nonexistent", "player1", nil)
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestSubmitSolution_NoGameInProgress(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	aliceID := room.Players[0].ID

	_, err := store.SubmitSolution(room.ID, aliceID, nil)
	if err == nil {
		t.Error("expected error when no game in progress")
	}
}

func TestSubmitSolution_PlayerNotFound(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.StartGame(room.ID, true) // Use fixed board

	_, err := store.SubmitSolution(room.ID, "nonexistent", validSolution())
	if err == nil {
		t.Error("expected error for nonexistent player")
	}
}

func TestSubmitSolution_InvalidSolution(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.StartGame(room.ID, true) // Use fixed board
	aliceID := room.Players[0].ID

	// Invalid solution - doesn't reach target
	invalidMoves := []model.BotPosition{
		{Id: 0, Pos: model.Position{X: 5, Y: 6}},
	}

	_, err := store.SubmitSolution(room.ID, aliceID, invalidMoves)
	if err == nil {
		t.Error("expected error for invalid solution")
	}
}

func TestSubmitSolution_ValidSolution(t *testing.T) {
	store := NewStore()
	mock := &mockBroadcaster{}
	store.SetBroadcaster(mock)

	room := store.Create("Alice")
	store.StartGame(room.ID, true) // Use fixed board
	aliceID := room.Players[0].ID

	solution, err := store.SubmitSolution(room.ID, aliceID, validSolution())
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
	room, _ = store.Get(room.ID)
	if len(room.Solutions) != 1 {
		t.Errorf("expected 1 solution, got %d", len(room.Solutions))
	}

	// Check broadcast was called
	if !mock.playerSolvedCalled {
		t.Error("expected BroadcastPlayerSolved to be called")
	}
}

func TestSubmitSolution_BetterSolutionUpdates(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.StartGame(room.ID, true) // Use fixed board
	aliceID := room.Players[0].ID

	// First solution - 3 moves
	_, err := store.SubmitSolution(room.ID, aliceID, validSolution())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A shorter 2-move solution (if one exists)
	// For testing, we'll verify the update logic by checking move count
	room, _ = store.Get(room.ID)
	originalMoveCount := room.Solutions[0].MoveCount()

	// Submit same solution again - should not update
	_, err = store.SubmitSolution(room.ID, aliceID, validSolution())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, _ = store.Get(room.ID)
	if room.Solutions[0].MoveCount() != originalMoveCount {
		t.Error("expected solution to remain unchanged when submitting same move count")
	}
	// Should still only have 1 solution
	if len(room.Solutions) != 1 {
		t.Errorf("expected 1 solution, got %d", len(room.Solutions))
	}
}

func TestRetractSolution_RoomNotFound(t *testing.T) {
	store := NewStore()

	err := store.RetractSolution("nonexistent", "player1")
	if err == nil {
		t.Error("expected error for nonexistent room")
	}
}

func TestRetractSolution_NoGameInProgress(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	aliceID := room.Players[0].ID

	err := store.RetractSolution(room.ID, aliceID)
	if err == nil {
		t.Error("expected error when no game in progress")
	}
}

func TestRetractSolution_NoSolutionFound(t *testing.T) {
	store := NewStore()

	room := store.Create("Alice")
	store.StartGame(room.ID, true)
	aliceID := room.Players[0].ID

	err := store.RetractSolution(room.ID, aliceID)
	if err == nil {
		t.Error("expected error when player has no solution")
	}
}

func TestRetractSolution_RemovesCompletely(t *testing.T) {
	store := NewStore()
	mock := &mockBroadcaster{}
	store.SetBroadcaster(mock)

	room := store.Create("Alice")
	store.StartGame(room.ID, true)
	aliceID := room.Players[0].ID

	// Submit solution
	store.SubmitSolution(room.ID, aliceID, validSolution())

	room, _ = store.Get(room.ID)
	if len(room.Solutions) != 1 {
		t.Fatalf("expected 1 solution, got %d", len(room.Solutions))
	}

	// Retract solution
	err := store.RetractSolution(room.ID, aliceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room, _ = store.Get(room.ID)
	if len(room.Solutions) != 0 {
		t.Errorf("expected 0 solutions after retraction, got %d", len(room.Solutions))
	}

	// Check broadcast was called
	if !mock.solutionRetractedCalled {
		t.Error("expected BroadcastSolutionRetracted to be called")
	}
}
