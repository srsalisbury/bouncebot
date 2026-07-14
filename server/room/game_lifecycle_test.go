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

	signals := startGame(gl, room)

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
		FinishedSolving: []string{"alice"},
		ReadyForNext:    []string{"alice"},
	}

	startGame(gl, room)

	if len(room.Solutions) != 0 {
		t.Error("expected Solutions to be cleared")
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
	startGame(gl, room)
	firstGameStartedAt := room.GameStartedAt

	time.Sleep(10 * time.Millisecond)

	// Second game
	startGame(gl, room)

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

	signals := startNextGame(gl, room)

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

func TestGameLifecycle_StartNextGame_NoWinner_ContinuesFromSolverSolution(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:          "TEST",
		Players:     []Player{{ID: "alice", Name: "Alice"}},
		CurrentGame: model.Game1(),
		Solutions:   nil, // nobody solved
		SolverResults: map[string]*SolverResult{
			"A-Star": {
				SolverName: "A-Star",
				Completed:  true,
				Moves:      movePayloadsFromBotPositions(model.Game1OptimalSolution()),
			},
		},
	}

	startNextGame(gl, room)

	// Bot 0 (the target bot) and bot 1 (moved as a blocker) should start the
	// next round where the solver's solution left them, not at Game1's
	// original starting positions.
	if got, want := room.CurrentGame.Bots[0], (model.Position{X: 4, Y: 13}); got != want {
		t.Errorf("bot 0 position = %v, want %v (solver's final position)", got, want)
	}
	if got, want := room.CurrentGame.Bots[1], (model.Position{X: 0, Y: 12}); got != want {
		t.Errorf("bot 1 position = %v, want %v (solver's final position)", got, want)
	}
	// Bots the solution never touched should be unchanged.
	if got, want := room.CurrentGame.Bots[2], (model.Position{X: 3, Y: 9}); got != want {
		t.Errorf("bot 2 position = %v, want %v (untouched by solution)", got, want)
	}
	if got, want := room.CurrentGame.Bots[3], (model.Position{X: 12, Y: 4}); got != want {
		t.Errorf("bot 3 position = %v, want %v (untouched by solution)", got, want)
	}
}

func TestGameLifecycle_StartNextGame_NoWinner_NoSolverResult_KeepsStartingPositions(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:            "TEST",
		Players:       []Player{{ID: "alice", Name: "Alice"}},
		CurrentGame:   model.Game1(),
		Solutions:     nil, // nobody solved
		SolverResults: nil, // and the solver hasn't finished (or isn't registered)
	}

	startNextGame(gl, room)

	// Falls back to today's behavior: robots stay exactly where Game1 started them.
	original := model.Game1()
	for id, pos := range original.Bots {
		if got := room.CurrentGame.Bots[id]; got != pos {
			t.Errorf("bot %d position = %v, want %v (unchanged, no solver result available)", id, got, pos)
		}
	}
}

func TestGameLifecycle_StartNextGame_PlayerWinTakesPriorityOverSolver(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID:          "TEST",
		Players:     []Player{{ID: "alice", Name: "Alice"}},
		CurrentGame: model.Game1(),
		Solutions:   []PlayerSolution{{PlayerID: "alice", Moves: model.Game1OptimalSolution()}},
		// Deliberately bogus - not a valid move for this board. If the code
		// incorrectly consulted the solver path after a player already won,
		// CheckSolution would fail on this and wipe out the correct
		// (player-derived) winningGameState, changing the assertions below.
		SolverResults: map[string]*SolverResult{
			"A-Star": {
				SolverName: "A-Star",
				Completed:  true,
				Moves:      []MovePayload{{RobotId: 2, X: 15, Y: 15}},
			},
		},
	}

	startNextGame(gl, room)

	// Should reflect the player's winning end state - bot 0 at target, bot 1
	// at its blocker spot - not be wiped out by the bogus solver data.
	if got, want := room.CurrentGame.Bots[0], (model.Position{X: 4, Y: 13}); got != want {
		t.Errorf("bot 0 position = %v, want %v", got, want)
	}
	if got, want := room.CurrentGame.Bots[1], (model.Position{X: 0, Y: 12}); got != want {
		t.Errorf("bot 1 position = %v, want %v", got, want)
	}
}

func TestBestSolverResult(t *testing.T) {
	results := map[string]*SolverResult{
		"BFS":      {SolverName: "BFS", Completed: true, Moves: make([]MovePayload, 12)},
		"A-Star":   {SolverName: "A-Star", Completed: true, Moves: make([]MovePayload, 8)},
		"Timedout": {SolverName: "Timedout", Completed: false, Moves: make([]MovePayload, 3)},
		"Empty":    {SolverName: "Empty", Completed: true, Moves: nil},
	}

	best := bestSolverResult(results)
	if best == nil {
		t.Fatal("expected a best result")
	}
	if best.SolverName != "A-Star" {
		t.Errorf("expected shortest completed result (A-Star, 8 moves), got %s (%d moves)", best.SolverName, len(best.Moves))
	}
}

func TestBestSolverResult_NoneCompleted(t *testing.T) {
	results := map[string]*SolverResult{
		"A-Star": {SolverName: "A-Star", Completed: false, Moves: make([]MovePayload, 8)},
	}

	if best := bestSolverResult(results); best != nil {
		t.Errorf("expected nil when nothing completed, got %v", best)
	}
}

func TestGameLifecycle_StartNextGame_MovesPendingPlayers(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusConnected},
		},
		PendingPlayers: []Player{
			{ID: "bob", Name: "Bob", Status: PlayerStatusConnected},
			{ID: "charlie", Name: "Charlie", Status: PlayerStatusConnected},
		},
		CurrentGame: model.Game1(),
		Wins:        make(map[string]int),
	}

	startNextGame(gl, room)

	// Check pending players were moved to active
	if len(room.Players) != 3 {
		t.Errorf("expected 3 active players, got %d", len(room.Players))
	}
	if len(room.PendingPlayers) != 0 {
		t.Errorf("expected 0 pending players, got %d", len(room.PendingPlayers))
	}

	// Check player names
	playerNames := make(map[string]bool)
	for _, p := range room.Players {
		playerNames[p.Name] = true
	}
	if !playerNames["Alice"] {
		t.Error("expected Alice in players")
	}
	if !playerNames["Bob"] {
		t.Error("expected Bob in players")
	}
	if !playerNames["Charlie"] {
		t.Error("expected Charlie in players")
	}
}

func TestGameLifecycle_StartNextGame_NoPendingPlayers(t *testing.T) {
	sm := NewSolutionManager()
	gl := NewGameLifecycle(sm)

	room := &Room{
		ID: "TEST",
		Players: []Player{
			{ID: "alice", Name: "Alice", Status: PlayerStatusConnected},
		},
		PendingPlayers: nil,
		CurrentGame:    model.Game1(),
		Wins:           make(map[string]int),
	}

	startNextGame(gl, room)

	// Check players unchanged
	if len(room.Players) != 1 {
		t.Errorf("expected 1 active player, got %d", len(room.Players))
	}
	if room.Players[0].Name != "Alice" {
		t.Errorf("expected Alice, got %s", room.Players[0].Name)
	}
}

func TestGameLifecycle_EndGameThenStartGame_NoDuplicateWins(t *testing.T) {
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

	// EndGame credits bob with 1 win and increments GamesPlayed
	gl.EndGame(room)

	if room.Wins["bob"] != 1 {
		t.Fatalf("expected bob to have 1 win after EndGame, got %d", room.Wins["bob"])
	}
	if room.GamesPlayed != 1 {
		t.Fatalf("expected 1 game played after EndGame, got %d", room.GamesPlayed)
	}

	// StartGame should NOT credit wins again
	startGame(gl, room)

	if room.Wins["bob"] != 1 {
		t.Errorf("expected bob to still have 1 win after StartGame, got %d (double-counted!)", room.Wins["bob"])
	}
	if room.GamesPlayed != 1 {
		t.Errorf("expected 1 game played after StartGame, got %d (double-counted!)", room.GamesPlayed)
	}
}

func TestGameLifecycle_EndGameThenStartNextGame_NoDuplicateWins(t *testing.T) {
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

	// EndGame credits bob with 1 win
	gl.EndGame(room)

	if room.Wins["bob"] != 1 {
		t.Fatalf("expected bob to have 1 win after EndGame, got %d", room.Wins["bob"])
	}

	// StartNextGame should NOT credit wins again
	startNextGame(gl, room)

	if room.Wins["bob"] != 1 {
		t.Errorf("expected bob to still have 1 win after StartNextGame, got %d (double-counted!)", room.Wins["bob"])
	}
	if room.GamesPlayed != 1 {
		t.Errorf("expected 1 game played, got %d (double-counted!)", room.GamesPlayed)
	}
}
