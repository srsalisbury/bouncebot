package room

import "github.com/srsalisbury/bouncebot/model"

// mockBroadcaster implements EventBroadcaster for testing
type mockBroadcaster struct {
	gameEndedCalled    bool
	gameStartedCalled  bool
	playerSolvedCalled bool
	playerLeftCalled   bool
}

func (m *mockBroadcaster) BroadcastPlayerJoined(roomID, playerID, playerName string) {}
func (m *mockBroadcaster) BroadcastPlayerLeft(roomID, playerID string)               { m.playerLeftCalled = true }
func (m *mockBroadcaster) BroadcastGameStarted(roomID string)                         { m.gameStartedCalled = true }
func (m *mockBroadcaster) BroadcastPlayerFinishedSolving(roomID, playerID string)     {}
func (m *mockBroadcaster) BroadcastPlayerReadyForNext(roomID, playerID string)        {}
func (m *mockBroadcaster) BroadcastPlayerSolved(roomID, playerID string, moveCount int) {
	m.playerSolvedCalled = true
}
func (m *mockBroadcaster) BroadcastGameEnded(roomID, winnerID, winnerName string, moves []MovePayload) {
	m.gameEndedCalled = true
}
func (m *mockBroadcaster) BroadcastRoomSettingsChanged(roomID string) {}

// validSolution returns model.Game1OptimalSolution for convenience.
func validSolution() []model.BotPosition {
	return model.Game1OptimalSolution()
}

// movePayloadsFromBotPositions is the test-side inverse of
// movePayloadsToBotPositions, for constructing SolverResult.Moves fixtures.
func movePayloadsFromBotPositions(moves []model.BotPosition) []MovePayload {
	result := make([]MovePayload, len(moves))
	for i, m := range moves {
		result[i] = MovePayload{RobotId: int(m.Id), X: int(m.Pos.X), Y: int(m.Pos.Y)}
	}
	return result
}

// startGame drives the NewGameSource+CommitNewGame pair the same way
// RoomService.StartGame does, for tests that used to call the old single-step
// gl.StartGame(room) directly. Room settings have no minimum configured in
// these tests, so this always takes generateBoard's single-candidate fast path.
func startGame(gl GameLifecycle, room *Room) []Signal {
	candidateFn, minLength := gl.NewGameSource(room)
	game, _ := generateBoard(minLength, candidateFn, nil)
	return gl.CommitNewGame(room, game)
}

// startNextGame drives the PromotePendingPlayers+NewGameSource+CommitNewGame
// sequence the same way the StartNextGameSignal handler does, for tests that
// used to call the old single-step gl.StartNextGame(room) directly.
func startNextGame(gl GameLifecycle, room *Room) []Signal {
	gl.PromotePendingPlayers(room)
	candidateFn, minLength := gl.NewGameSource(room)
	game, _ := generateBoard(minLength, candidateFn, nil)
	return gl.CommitNewGame(room, game)
}
