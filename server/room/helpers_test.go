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

// validSolution returns a valid 7-move solution for Game1 (fixed board).
// Target is bot 0 at (5, 13), starting at (5, 4).
// Moves: Bot 1 left, then Bot 0: up, left, down, left, up, right
func validSolution() []model.BotPosition {
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
