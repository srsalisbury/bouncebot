package room

import (
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

func TestPlayerManager_AddPlayer(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID:             "TEST",
		Players:        []Player{{ID: "alice", Name: "Alice", Status: PlayerStatusConnected}},
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}
	oldActivity := room.LastActivityAt

	time.Sleep(10 * time.Millisecond)

	signals, err := pm.AddPlayer(room, "Bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check player was added
	if len(room.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(room.Players))
	}
	if room.Players[1].Name != "Bob" {
		t.Errorf("expected player name 'Bob', got '%s'", room.Players[1].Name)
	}
	if room.Players[1].ID == "" {
		t.Error("expected player ID to be set")
	}
	if room.Players[1].Status != PlayerStatusConnected {
		t.Errorf("expected player status 'connected', got '%s'", room.Players[1].Status)
	}

	// Check activity updated
	if !room.LastActivityAt.After(oldActivity) {
		t.Error("expected LastActivityAt to be updated")
	}

	// Check broadcast signal
	if len(signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(signals))
	}
	broadcast, ok := signals[0].(BroadcastSignal)
	if !ok {
		t.Fatal("expected BroadcastSignal")
	}
	event, ok := broadcast.Event.(PlayerJoinedEvent)
	if !ok {
		t.Fatal("expected PlayerJoinedEvent")
	}
	if event.PlayerName != "Bob" {
		t.Errorf("expected player name 'Bob' in event, got '%s'", event.PlayerName)
	}
}

func TestPlayerManager_DisconnectPlayer(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID:      "TEST",
		Players: []Player{{ID: "alice", Name: "Alice", Status: PlayerStatusConnected}},
	}

	signals, err := pm.DisconnectPlayer(room, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check player status
	if room.Players[0].Status != PlayerStatusDisconnected {
		t.Errorf("expected player status 'disconnected', got '%s'", room.Players[0].Status)
	}
	if room.Players[0].DisconnectedAt.IsZero() {
		t.Error("expected DisconnectedAt to be set")
	}

	// Check start timer signal
	if len(signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(signals))
	}
	timerSignal, ok := signals[0].(StartTimerSignal)
	if !ok {
		t.Fatal("expected StartTimerSignal")
	}
	if timerSignal.PlayerID != "alice" {
		t.Errorf("expected timer for alice, got %s", timerSignal.PlayerID)
	}
	if timerSignal.RoomID != "TEST" {
		t.Errorf("expected room TEST, got %s", timerSignal.RoomID)
	}
}

func TestPlayerManager_DisconnectPlayer_NotFound(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID:      "TEST",
		Players: []Player{{ID: "alice", Name: "Alice", Status: PlayerStatusConnected}},
	}

	// Should not error for nonexistent player (may have been removed already)
	signals, err := pm.DisconnectPlayer(room, "nonexistent")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(signals) != 0 {
		t.Errorf("expected no signals, got %d", len(signals))
	}
}

func TestPlayerManager_ReconnectPlayer(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID: "TEST",
		Players: []Player{{
			ID:             "alice",
			Name:           "Alice",
			Status:         PlayerStatusDisconnected,
			DisconnectedAt: time.Now(),
		}},
	}

	signals, err := pm.ReconnectPlayer(room, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check player status
	if room.Players[0].Status != PlayerStatusConnected {
		t.Errorf("expected player status 'connected', got '%s'", room.Players[0].Status)
	}

	// Check cancel timer signal
	if len(signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(signals))
	}
	cancelSignal, ok := signals[0].(CancelTimerSignal)
	if !ok {
		t.Fatal("expected CancelTimerSignal")
	}
	if cancelSignal.PlayerID != "alice" {
		t.Errorf("expected cancel for alice, got %s", cancelSignal.PlayerID)
	}
}

func TestPlayerManager_ReconnectPlayer_NotFound(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID:      "TEST",
		Players: []Player{{ID: "alice", Name: "Alice", Status: PlayerStatusConnected}},
	}

	_, err := pm.ReconnectPlayer(room, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent player")
	}
}

func TestPlayerManager_ReconnectPlayer_AlreadyConnected(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID:      "TEST",
		Players: []Player{{ID: "alice", Name: "Alice", Status: PlayerStatusConnected}},
	}

	signals, err := pm.ReconnectPlayer(room, "alice")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Errorf("expected no signals for already connected player, got %d", len(signals))
	}
}

func TestPlayerManager_RemovePlayer(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusDisconnected},
			{ID: "bob", Name: "Bob", Status: PlayerStatusConnected},
		},
	}

	signals := pm.RemovePlayer(room, "alice")

	// Check player removed
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(room.Players))
	}
	if room.Players[0].ID != "bob" {
		t.Errorf("expected remaining player to be bob, got %s", room.Players[0].ID)
	}

	// Check signals
	if len(signals) < 2 {
		t.Fatalf("expected at least 2 signals, got %d", len(signals))
	}

	// Should have cancel timer signal
	hasCancel := false
	for _, sig := range signals {
		if cancel, ok := sig.(CancelTimerSignal); ok && cancel.PlayerID == "alice" {
			hasCancel = true
			break
		}
	}
	if !hasCancel {
		t.Error("expected CancelTimerSignal for alice")
	}

	// Should have player left broadcast
	hasBroadcast := false
	for _, sig := range signals {
		if broadcast, ok := sig.(BroadcastSignal); ok {
			if event, ok := broadcast.Event.(PlayerLeftEvent); ok && event.PlayerID == "alice" {
				hasBroadcast = true
				break
			}
		}
	}
	if !hasBroadcast {
		t.Error("expected PlayerLeftEvent broadcast")
	}
}

func TestPlayerManager_RemovePlayer_OnlyRemovesDisconnected(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusConnected},
			{ID: "bob", Name: "Bob", Status: PlayerStatusConnected},
		},
	}

	signals := pm.RemovePlayer(room, "alice")

	// Should not remove connected player
	if len(room.Players) != 2 {
		t.Errorf("expected 2 players (connected player should not be removed), got %d", len(room.Players))
	}
	if len(signals) != 0 {
		t.Errorf("expected no signals, got %d", len(signals))
	}
}

func TestPlayerManager_RemovePlayer_NotFound(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID:      "TEST",
		Players: []Player{{ID: "alice", Name: "Alice", Status: PlayerStatusConnected}},
	}

	signals := pm.RemovePlayer(room, "nonexistent")
	if len(signals) != 0 {
		t.Errorf("expected no signals for nonexistent player, got %d", len(signals))
	}
}

func TestPlayerManager_RemovePlayer_CleansUpFinishedSolving(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusDisconnected},
			{ID: "bob", Name: "Bob", Status: PlayerStatusConnected},
		},
		FinishedSolving: []string{"alice", "bob"},
	}

	pm.RemovePlayer(room, "alice")

	if len(room.FinishedSolving) != 1 {
		t.Errorf("expected 1 finished player, got %d", len(room.FinishedSolving))
	}
	if room.FinishedSolving[0] != "bob" {
		t.Errorf("expected bob in FinishedSolving, got %s", room.FinishedSolving[0])
	}
}

func TestPlayerManager_RemovePlayer_CleansUpReadyForNext(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusDisconnected},
			{ID: "bob", Name: "Bob", Status: PlayerStatusConnected},
		},
		ReadyForNext: []string{"alice", "bob"},
	}

	pm.RemovePlayer(room, "alice")

	if len(room.ReadyForNext) != 1 {
		t.Errorf("expected 1 ready player, got %d", len(room.ReadyForNext))
	}
	if room.ReadyForNext[0] != "bob" {
		t.Errorf("expected bob in ReadyForNext, got %s", room.ReadyForNext[0])
	}
}

func TestPlayerManager_RemovePlayer_CleansUpSolutions(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusDisconnected},
			{ID: "bob", Name: "Bob", Status: PlayerStatusConnected},
		},
		Solutions: []PlayerSolution{
			{PlayerID: "alice", SolvedAt: time.Now()},
			{PlayerID: "bob", SolvedAt: time.Now()},
		},
	}

	pm.RemovePlayer(room, "alice")

	if len(room.Solutions) != 1 {
		t.Errorf("expected 1 solution, got %d", len(room.Solutions))
	}
	if room.Solutions[0].PlayerID != "bob" {
		t.Errorf("expected bob's solution, got %s", room.Solutions[0].PlayerID)
	}
}

func TestPlayerManager_RemovePlayer_TriggersEndGame(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusConnected},
			{ID: "bob", Name: "Bob", Status: PlayerStatusDisconnected},
		},
		CurrentGame:     model.Game1(),
		FinishedSolving: []string{"alice"},
	}

	signals := pm.RemovePlayer(room, "bob")

	// Should include EndGameSignal since alice is only player and she's finished
	hasEndGame := false
	for _, sig := range signals {
		if endGame, ok := sig.(EndGameSignal); ok && endGame.RoomID == "TEST" {
			hasEndGame = true
			break
		}
	}
	if !hasEndGame {
		t.Error("expected EndGameSignal when last unfinished player is removed")
	}
}

func TestPlayerManager_RemovePlayer_TriggersNextGame(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusConnected},
			{ID: "bob", Name: "Bob", Status: PlayerStatusDisconnected},
		},
		ReadyForNext: []string{"alice"},
	}

	signals := pm.RemovePlayer(room, "bob")

	// Should include StartNextGameSignal since alice is only player and she's ready
	hasNextGame := false
	for _, sig := range signals {
		if nextGame, ok := sig.(StartNextGameSignal); ok && nextGame.RoomID == "TEST" {
			hasNextGame = true
			break
		}
	}
	if !hasNextGame {
		t.Error("expected StartNextGameSignal when last unready player is removed")
	}
}

func TestPlayerManager_RemovePlayer_NoTriggersIfNoPlayersLeft(t *testing.T) {
	pm := NewPlayerManager()

	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusDisconnected},
		},
		CurrentGame:     model.Game1(),
		FinishedSolving: []string{"alice"},
		ReadyForNext:    []string{"alice"},
	}

	signals := pm.RemovePlayer(room, "alice")

	// Should NOT include game signals since no players remain
	for _, sig := range signals {
		if _, ok := sig.(EndGameSignal); ok {
			t.Error("should not trigger EndGame with no players")
		}
		if _, ok := sig.(StartNextGameSignal); ok {
			t.Error("should not trigger StartNextGame with no players")
		}
	}
}
