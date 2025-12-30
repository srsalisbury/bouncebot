package room

import (
	"testing"
	"time"
)

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
