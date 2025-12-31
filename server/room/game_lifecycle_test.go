package room

import (
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

func TestGameLifecycle_StartGame(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:             "TEST",
		Players:        []Player{{ID: "alice", Name: "Alice", Status: PlayerStatusConnected}},
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
		Wins:           map[string]int{},
	}

	signals, err := gl.StartGame(room)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.CurrentGame == nil {
		t.Error("expected game to be set after StartGame")
	}
	if room.GameStartedAt == nil {
		t.Error("expected GameStartedAt to be set")
	}

	// Check broadcast signal
	if len(signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(signals))
	}
	broadcast, ok := signals[0].(BroadcastSignal)
	if !ok {
		t.Fatal("expected BroadcastSignal")
	}
	_, ok = broadcast.Event.(GameStartedEvent)
	if !ok {
		t.Error("expected GameStartedEvent")
	}
}

func TestGameLifecycle_StartGame_ClearsGameState(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:              "TEST",
		Players:         []Player{{ID: "alice", Name: "Alice"}},
		CreatedAt:       time.Now(),
		Wins:            map[string]int{},
		Solutions:       []PlayerSolution{{PlayerID: "alice"}},
		SolutionHistory: []PlayerSolutionHistory{{PlayerID: "alice"}},
		FinishedSolving: []string{"alice"},
		ReadyForNext:    []string{"alice"},
	}

	gl.StartGame(room)

	if len(room.Solutions) != 0 {
		t.Error("expected Solutions to be cleared")
	}
	if len(room.SolutionHistory) != 0 {
		t.Error("expected SolutionHistory to be cleared")
	}
	if len(room.FinishedSolving) != 0 {
		t.Error("expected FinishedSolving to be cleared")
	}
	if len(room.ReadyForNext) != 0 {
		t.Error("expected ReadyForNext to be cleared")
	}
}

func TestGameLifecycle_StartGame_Multiple(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:             "TEST",
		Players:        []Player{{ID: "alice", Name: "Alice"}},
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
		Wins:           map[string]int{},
	}

	// First game
	gl.StartGame(room)
	firstGameStartedAt := room.GameStartedAt

	time.Sleep(10 * time.Millisecond)

	// Second game
	gl.StartGame(room)

	if room.GameStartedAt == firstGameStartedAt {
		t.Error("expected GameStartedAt to be updated for new game")
	}
}

func TestGameLifecycle_MarkFinishedSolving(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:             "TEST",
		Players:        []Player{{ID: "alice", Name: "Alice"}, {ID: "bob", Name: "Bob"}},
		CurrentGame:    model.Game1(),
		LastActivityAt: time.Now(),
	}

	oldActivity := room.LastActivityAt
	time.Sleep(10 * time.Millisecond)

	signals, err := gl.MarkFinishedSolving(room, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check alice is finished
	if len(room.FinishedSolving) != 1 || room.FinishedSolving[0] != "alice" {
		t.Errorf("expected alice in FinishedSolving, got %v", room.FinishedSolving)
	}

	// Check activity updated
	if !room.LastActivityAt.After(oldActivity) {
		t.Error("expected LastActivityAt to be updated")
	}

	// Check broadcast signal (no EndGameSignal since bob hasn't finished)
	if len(signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(signals))
	}
	broadcast, ok := signals[0].(BroadcastSignal)
	if !ok {
		t.Fatal("expected BroadcastSignal")
	}
	event, ok := broadcast.Event.(PlayerFinishedSolvingEvent)
	if !ok {
		t.Fatal("expected PlayerFinishedSolvingEvent")
	}
	if event.PlayerID != "alice" {
		t.Errorf("expected alice in event, got %s", event.PlayerID)
	}
}

func TestGameLifecycle_MarkFinishedSolving_NoGameInProgress(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:          "TEST",
		Players:     []Player{{ID: "alice", Name: "Alice"}},
		CurrentGame: nil,
	}

	_, err := gl.MarkFinishedSolving(room, "alice")
	if err == nil {
		t.Error("expected error when no game in progress")
	}
}

func TestGameLifecycle_MarkFinishedSolving_PlayerNotFound(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:          "TEST",
		Players:     []Player{{ID: "alice", Name: "Alice"}},
		CurrentGame: model.Game1(),
	}

	_, err := gl.MarkFinishedSolving(room, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent player")
	}
}

func TestGameLifecycle_MarkFinishedSolving_AlreadyFinished(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:              "TEST",
		Players:         []Player{{ID: "alice", Name: "Alice"}},
		CurrentGame:     model.Game1(),
		FinishedSolving: []string{"alice"},
	}

	signals, err := gl.MarkFinishedSolving(room, "alice")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Errorf("expected no signals for already finished player, got %d", len(signals))
	}
}

func TestGameLifecycle_MarkFinishedSolving_TriggersEndGame(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:              "TEST",
		Players:         []Player{{ID: "alice", Name: "Alice"}, {ID: "bob", Name: "Bob"}},
		CurrentGame:     model.Game1(),
		FinishedSolving: []string{"alice"},
	}

	signals, err := gl.MarkFinishedSolving(room, "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should include EndGameSignal since all players are finished
	hasEndGame := false
	for _, sig := range signals {
		if _, ok := sig.(EndGameSignal); ok {
			hasEndGame = true
			break
		}
	}
	if !hasEndGame {
		t.Error("expected EndGameSignal when all players are finished")
	}
}

func TestGameLifecycle_MarkReadyForNext(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:             "TEST",
		Players:        []Player{{ID: "alice", Name: "Alice"}, {ID: "bob", Name: "Bob"}},
		LastActivityAt: time.Now(),
	}

	oldActivity := room.LastActivityAt
	time.Sleep(10 * time.Millisecond)

	signals, err := gl.MarkReadyForNext(room, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check alice is ready
	if len(room.ReadyForNext) != 1 || room.ReadyForNext[0] != "alice" {
		t.Errorf("expected alice in ReadyForNext, got %v", room.ReadyForNext)
	}

	// Check activity updated
	if !room.LastActivityAt.After(oldActivity) {
		t.Error("expected LastActivityAt to be updated")
	}

	// Check broadcast signal
	if len(signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(signals))
	}
}

func TestGameLifecycle_MarkReadyForNext_PlayerNotFound(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:      "TEST",
		Players: []Player{{ID: "alice", Name: "Alice"}},
	}

	_, err := gl.MarkReadyForNext(room, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent player")
	}
}

func TestGameLifecycle_MarkReadyForNext_AlreadyReady(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:           "TEST",
		Players:      []Player{{ID: "alice", Name: "Alice"}},
		ReadyForNext: []string{"alice"},
	}

	signals, err := gl.MarkReadyForNext(room, "alice")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Errorf("expected no signals for already ready player, got %d", len(signals))
	}
}

func TestGameLifecycle_MarkReadyForNext_TriggersNextGame(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:           "TEST",
		Players:      []Player{{ID: "alice", Name: "Alice"}, {ID: "bob", Name: "Bob"}},
		ReadyForNext: []string{"alice"},
	}

	signals, err := gl.MarkReadyForNext(room, "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should include StartNextGameSignal since all players are ready
	hasNextGame := false
	for _, sig := range signals {
		if _, ok := sig.(StartNextGameSignal); ok {
			hasNextGame = true
			break
		}
	}
	if !hasNextGame {
		t.Error("expected StartNextGameSignal when all players are ready")
	}
}

func TestGameLifecycle_EndGame(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:          "TEST",
		Players:     []Player{{ID: "alice", Name: "Alice"}, {ID: "bob", Name: "Bob"}},
		CurrentGame: model.Game1(),
		Wins:        map[string]int{},
		Solutions: []PlayerSolution{
			{PlayerID: "alice", SolvedAt: time.Now(), Moves: make([]model.BotPosition, 8)},
			{PlayerID: "bob", SolvedAt: time.Now(), Moves: make([]model.BotPosition, 5)},
		},
		GamesPlayed: 0,
	}

	signals := gl.EndGame(room)

	// Check winner credited
	if room.Wins["bob"] != 1 {
		t.Errorf("expected bob to have 1 win, got %d", room.Wins["bob"])
	}

	// Check games played incremented
	if room.GamesPlayed != 1 {
		t.Errorf("expected 1 game played, got %d", room.GamesPlayed)
	}

	// Check GameEndedEvent
	if len(signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(signals))
	}
	broadcast, ok := signals[0].(BroadcastSignal)
	if !ok {
		t.Fatal("expected BroadcastSignal")
	}
	event, ok := broadcast.Event.(GameEndedEvent)
	if !ok {
		t.Fatal("expected GameEndedEvent")
	}
	if event.WinnerID != "bob" {
		t.Errorf("expected winner bob, got %s", event.WinnerID)
	}
	if event.WinnerName != "Bob" {
		t.Errorf("expected winner name Bob, got %s", event.WinnerName)
	}
}

func TestGameLifecycle_EndGame_NoSolutions(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:          "TEST",
		Players:     []Player{{ID: "alice", Name: "Alice"}},
		CurrentGame: model.Game1(),
		Wins:        map[string]int{},
		Solutions:   []PlayerSolution{},
		GamesPlayed: 0,
	}

	signals := gl.EndGame(room)

	// Games played should still increment
	if room.GamesPlayed != 1 {
		t.Errorf("expected 1 game played, got %d", room.GamesPlayed)
	}

	// Check GameEndedEvent with no winner
	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	broadcast := signals[0].(BroadcastSignal)
	event := broadcast.Event.(GameEndedEvent)
	if event.WinnerID != "" {
		t.Errorf("expected no winner, got %s", event.WinnerID)
	}
}

func TestGameLifecycle_StartNextGame(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:              "TEST",
		Players:         []Player{{ID: "alice", Name: "Alice"}},
		CurrentGame:     model.Game1(),
		Solutions:       []PlayerSolution{{PlayerID: "alice"}},
		FinishedSolving: []string{"alice"},
		ReadyForNext:    []string{"alice"},
	}

	signals := gl.StartNextGame(room)

	// Check new game started
	if room.CurrentGame == nil {
		t.Error("expected new game to be set")
	}
	if room.GameStartedAt == nil {
		t.Error("expected GameStartedAt to be set")
	}

	// Check game state cleared
	if len(room.Solutions) != 0 {
		t.Error("expected Solutions to be cleared")
	}
	if len(room.FinishedSolving) != 0 {
		t.Error("expected FinishedSolving to be cleared")
	}
	if len(room.ReadyForNext) != 0 {
		t.Error("expected ReadyForNext to be cleared")
	}

	// Check broadcast signal
	if len(signals) != 1 {
		t.Errorf("expected 1 signal, got %d", len(signals))
	}
	broadcast, ok := signals[0].(BroadcastSignal)
	if !ok {
		t.Fatal("expected BroadcastSignal")
	}
	_, ok = broadcast.Event.(GameStartedEvent)
	if !ok {
		t.Error("expected GameStartedEvent")
	}
}
