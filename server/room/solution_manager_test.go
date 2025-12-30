package room

import (
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// createTestRoom creates a room with a game in progress for testing
func createTestRoom() *Room {
	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusConnected},
			{ID: "bob", Name: "Bob", Status: PlayerStatusConnected},
		},
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
		Wins:           map[string]int{},
		CurrentGame:    model.Game1(), // Use fixed board
	}
	return room
}

func TestSolutionManager_SubmitSolution_NoGameInProgress(t *testing.T) {
	sm := NewSolutionManager()

	room := &Room{
		ID:          "TEST",
		Players:     []Player{{ID: "alice", Name: "Alice"}},
		CurrentGame: nil,
	}

	_, _, err := sm.SubmitSolution(room, "alice", validSolution())
	if err == nil {
		t.Error("expected error when no game in progress")
	}
}

func TestSolutionManager_SubmitSolution_PlayerNotFound(t *testing.T) {
	sm := NewSolutionManager()
	room := createTestRoom()

	_, _, err := sm.SubmitSolution(room, "nonexistent", validSolution())
	if err == nil {
		t.Error("expected error for nonexistent player")
	}
}

func TestSolutionManager_SubmitSolution_InvalidSolution(t *testing.T) {
	sm := NewSolutionManager()
	room := createTestRoom()

	invalidMoves := []model.BotPosition{
		{Id: 0, Pos: model.Position{X: 5, Y: 6}},
	}

	_, _, err := sm.SubmitSolution(room, "alice", invalidMoves)
	if err == nil {
		t.Error("expected error for invalid solution")
	}
}

func TestSolutionManager_SubmitSolution_ValidSolution(t *testing.T) {
	sm := NewSolutionManager()
	room := createTestRoom()

	solution, signals, err := sm.SubmitSolution(room, "alice", validSolution())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if solution == nil {
		t.Fatal("expected solution to be returned")
	}
	if solution.PlayerID != "alice" {
		t.Errorf("expected player ID 'alice', got '%s'", solution.PlayerID)
	}
	if solution.MoveCount() != 7 {
		t.Errorf("expected 7 moves, got %d", solution.MoveCount())
	}

	// Check room has solution
	if len(room.Solutions) != 1 {
		t.Errorf("expected 1 solution in room, got %d", len(room.Solutions))
	}

	// Check broadcast signal was returned
	if len(signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(signals))
	}
	broadcast, ok := signals[0].(BroadcastSignal)
	if !ok {
		t.Fatal("expected BroadcastSignal")
	}
	event, ok := broadcast.Event.(PlayerSolvedEvent)
	if !ok {
		t.Fatal("expected PlayerSolvedEvent")
	}
	if event.PlayerID != "alice" || event.MoveCount != 7 {
		t.Errorf("unexpected event: %+v", event)
	}
}

func TestSolutionManager_SubmitSolution_BetterSolutionUpdates(t *testing.T) {
	sm := NewSolutionManager()
	room := createTestRoom()

	// First solution
	sm.SubmitSolution(room, "alice", validSolution())

	originalMoveCount := room.Solutions[0].MoveCount()

	// Submit same solution again - should not update
	_, signals, err := sm.SubmitSolution(room, "alice", validSolution())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.Solutions[0].MoveCount() != originalMoveCount {
		t.Error("expected solution to remain unchanged when submitting same move count")
	}
	if len(room.Solutions) != 1 {
		t.Errorf("expected 1 solution, got %d", len(room.Solutions))
	}

	// No broadcast for same solution
	if len(signals) != 0 {
		t.Errorf("expected no signals for same solution, got %d", len(signals))
	}
}

func TestSolutionManager_SubmitSolution_UpdatesLastActivityAt(t *testing.T) {
	sm := NewSolutionManager()
	room := createTestRoom()

	oldActivity := room.LastActivityAt
	time.Sleep(10 * time.Millisecond)

	sm.SubmitSolution(room, "alice", validSolution())

	if !room.LastActivityAt.After(oldActivity) {
		t.Error("expected LastActivityAt to be updated")
	}
}

func TestSolutionManager_SubmitSolution_AddsToHistory(t *testing.T) {
	sm := NewSolutionManager()
	room := createTestRoom()

	sm.SubmitSolution(room, "alice", validSolution())

	if len(room.SolutionHistory) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(room.SolutionHistory))
	}
	if room.SolutionHistory[0].PlayerID != "alice" {
		t.Errorf("expected history for alice, got %s", room.SolutionHistory[0].PlayerID)
	}
	if len(room.SolutionHistory[0].Solutions) != 1 {
		t.Errorf("expected 1 solution in history, got %d", len(room.SolutionHistory[0].Solutions))
	}
}

func TestSolutionManager_RetractSolution_NoGameInProgress(t *testing.T) {
	sm := NewSolutionManager()

	room := &Room{
		ID:          "TEST",
		Players:     []Player{{ID: "alice", Name: "Alice"}},
		CurrentGame: nil,
	}

	_, err := sm.RetractSolution(room, "alice")
	if err == nil {
		t.Error("expected error when no game in progress")
	}
}

func TestSolutionManager_RetractSolution_NoSolutionFound(t *testing.T) {
	sm := NewSolutionManager()
	room := createTestRoom()

	_, err := sm.RetractSolution(room, "alice")
	if err == nil {
		t.Error("expected error when player has no solution")
	}
}

func TestSolutionManager_RetractSolution_RemovesCompletely(t *testing.T) {
	sm := NewSolutionManager()
	room := createTestRoom()

	// Submit solution
	sm.SubmitSolution(room, "alice", validSolution())

	if len(room.Solutions) != 1 {
		t.Fatalf("expected 1 solution, got %d", len(room.Solutions))
	}

	// Retract
	signals, err := sm.RetractSolution(room, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(room.Solutions) != 0 {
		t.Errorf("expected 0 solutions after retraction, got %d", len(room.Solutions))
	}

	// Check broadcast signal
	if len(signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(signals))
	}
	broadcast, ok := signals[0].(BroadcastSignal)
	if !ok {
		t.Fatal("expected BroadcastSignal")
	}
	_, ok = broadcast.Event.(SolutionRetractedEvent)
	if !ok {
		t.Error("expected SolutionRetractedEvent")
	}
}

func TestSolutionManager_RetractSolution_UpdatesLastActivityAt(t *testing.T) {
	sm := NewSolutionManager()
	room := createTestRoom()

	sm.SubmitSolution(room, "alice", validSolution())

	oldActivity := room.LastActivityAt
	time.Sleep(10 * time.Millisecond)

	sm.RetractSolution(room, "alice")

	if !room.LastActivityAt.After(oldActivity) {
		t.Error("expected LastActivityAt to be updated")
	}
}

func TestSolutionManager_GetWinningSolution_Empty(t *testing.T) {
	sm := NewSolutionManager()

	winner := sm.GetWinningSolution(nil)
	if winner != nil {
		t.Error("expected nil for empty solutions")
	}

	winner = sm.GetWinningSolution([]PlayerSolution{})
	if winner != nil {
		t.Error("expected nil for empty slice")
	}
}

func TestSolutionManager_GetWinningSolution_Single(t *testing.T) {
	sm := NewSolutionManager()

	solutions := []PlayerSolution{
		{PlayerID: "alice", SolvedAt: time.Now(), Moves: make([]model.BotPosition, 5)},
	}

	winner := sm.GetWinningSolution(solutions)
	if winner == nil {
		t.Fatal("expected winner")
	}
	if winner.PlayerID != "alice" {
		t.Errorf("expected alice, got %s", winner.PlayerID)
	}
}

func TestSolutionManager_GetWinningSolution_FewerMoves(t *testing.T) {
	sm := NewSolutionManager()

	now := time.Now()
	solutions := []PlayerSolution{
		{PlayerID: "alice", SolvedAt: now, Moves: make([]model.BotPosition, 10)},
		{PlayerID: "bob", SolvedAt: now.Add(time.Second), Moves: make([]model.BotPosition, 5)},
	}

	winner := sm.GetWinningSolution(solutions)
	if winner.PlayerID != "bob" {
		t.Errorf("expected bob (fewer moves), got %s", winner.PlayerID)
	}
}

func TestSolutionManager_GetWinningSolution_TiebreakerByTime(t *testing.T) {
	sm := NewSolutionManager()

	now := time.Now()
	solutions := []PlayerSolution{
		{PlayerID: "alice", SolvedAt: now.Add(time.Second), Moves: make([]model.BotPosition, 5)},
		{PlayerID: "bob", SolvedAt: now, Moves: make([]model.BotPosition, 5)},
	}

	winner := sm.GetWinningSolution(solutions)
	if winner.PlayerID != "bob" {
		t.Errorf("expected bob (solved earlier), got %s", winner.PlayerID)
	}
}

func TestSolutionManager_GetWinningSolution_Multiple(t *testing.T) {
	sm := NewSolutionManager()

	now := time.Now()
	solutions := []PlayerSolution{
		{PlayerID: "alice", SolvedAt: now, Moves: make([]model.BotPosition, 8)},
		{PlayerID: "bob", SolvedAt: now.Add(time.Second), Moves: make([]model.BotPosition, 5)},
		{PlayerID: "charlie", SolvedAt: now.Add(2 * time.Second), Moves: make([]model.BotPosition, 6)},
	}

	winner := sm.GetWinningSolution(solutions)
	if winner.PlayerID != "bob" {
		t.Errorf("expected bob (5 moves), got %s with %d moves", winner.PlayerID, winner.MoveCount())
	}
}
