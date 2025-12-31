package room

import "github.com/srsalisbury/bouncebot/model"

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

// validSolution returns model.Game1Solution for convenience.
func validSolution() []model.BotPosition {
	return model.Game1Solution()
}
