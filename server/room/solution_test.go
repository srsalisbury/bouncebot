package room

import (
	"testing"

	"github.com/srsalisbury/bouncebot/model"
)

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

	// First solution - 7 moves
	_, err := store.SubmitSolution(room.ID, aliceID, validSolution())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
