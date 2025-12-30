package room

import (
	"fmt"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// GameLifecycle manages game state transitions.
type GameLifecycle interface {
	// StartGame starts a new game in the room.
	// Returns signals or error.
	StartGame(room *Room, useFixedBoard bool) ([]Signal, error)

	// MarkFinishedSolving marks a player as finished solving.
	// Returns signals or error.
	MarkFinishedSolving(room *Room, playerID string) ([]Signal, error)

	// MarkReadyForNext marks a player as ready for the next game.
	// Returns signals or error.
	MarkReadyForNext(room *Room, playerID string) ([]Signal, error)

	// EndGame ends the current game and determines the winner.
	// Returns signals.
	EndGame(room *Room) []Signal

	// StartNextGame starts the next game (continuation from current).
	// Returns signals.
	StartNextGame(room *Room) []Signal
}

// gameLifecycle is the concrete implementation of GameLifecycle.
type gameLifecycle struct {
	solutionMgr SolutionManager
}

// NewGameLifecycle creates a new GameLifecycle.
// Requires SolutionManager for determining winners.
func NewGameLifecycle(solutionMgr SolutionManager) GameLifecycle {
	return &gameLifecycle{solutionMgr: solutionMgr}
}

func (gl *gameLifecycle) StartGame(room *Room, useFixedBoard bool) ([]Signal, error) {
	// If there was a previous game with solutions, determine and record the winner
	// and get the final game state from the winning solution
	var winningGameState *model.Game
	if room.CurrentGame != nil && len(room.Solutions) > 0 {
		winningSolution := gl.solutionMgr.GetWinningSolution(room.Solutions)
		if winningSolution != nil {
			room.Wins[winningSolution.PlayerID]++
			// Apply winning moves to get final robot positions
			if len(winningSolution.Moves) > 0 {
				_, winningGameState = room.CurrentGame.CheckSolution(winningSolution.Moves)
			}
		}
		room.GamesPlayed++
	}

	// Generate game
	var game *model.Game
	if useFixedBoard {
		game = model.Game1()
	} else if winningGameState != nil {
		// Continue from winning game state: same board, robots at final positions
		game = model.NewContinuationGame(winningGameState)
	} else if room.CurrentGame != nil {
		// No winning solution with moves, continue from existing game
		game = model.NewContinuationGame(room.CurrentGame)
	} else {
		// First game: fully random
		game = model.NewRandomGame()
	}
	now := time.Now()

	room.CurrentGame = game
	room.GameStartedAt = &now
	room.LastActivityAt = now
	room.ClearGameState()

	signals := []Signal{
		BroadcastSignal{Event: GameStartedEvent{RoomID: room.ID}},
	}

	return signals, nil
}

func (gl *gameLifecycle) MarkFinishedSolving(room *Room, playerID string) ([]Signal, error) {
	if room.CurrentGame == nil {
		return nil, fmt.Errorf("no game in progress")
	}

	// Verify player exists
	if room.GetPlayerName(playerID) == "" {
		return nil, fmt.Errorf("player not found: %s", playerID)
	}

	room.LastActivityAt = time.Now()

	// Check if already finished
	if containsString(room.FinishedSolving, playerID) {
		return nil, nil
	}

	room.FinishedSolving = append(room.FinishedSolving, playerID)

	signals := []Signal{
		BroadcastSignal{Event: PlayerFinishedSolvingEvent{
			RoomID:   room.ID,
			PlayerID: playerID,
		}},
	}

	// Check if all players are finished -> signal end game
	if len(room.FinishedSolving) == len(room.Players) {
		signals = append(signals, EndGameSignal{RoomID: room.ID})
	}

	return signals, nil
}

func (gl *gameLifecycle) MarkReadyForNext(room *Room, playerID string) ([]Signal, error) {
	// Verify player exists
	if room.GetPlayerName(playerID) == "" {
		return nil, fmt.Errorf("player not found: %s", playerID)
	}

	room.LastActivityAt = time.Now()

	// Check if already ready
	if containsString(room.ReadyForNext, playerID) {
		return nil, nil
	}

	room.ReadyForNext = append(room.ReadyForNext, playerID)

	signals := []Signal{
		BroadcastSignal{Event: PlayerReadyForNextEvent{
			RoomID:   room.ID,
			PlayerID: playerID,
		}},
	}

	// Check if all players are ready -> signal start next game
	if len(room.ReadyForNext) == len(room.Players) {
		signals = append(signals, StartNextGameSignal{RoomID: room.ID})
	}

	return signals, nil
}

func (gl *gameLifecycle) EndGame(room *Room) []Signal {
	// Credit the win and increment games played
	winner := gl.solutionMgr.GetWinningSolution(room.Solutions)
	if winner != nil {
		room.Wins[winner.PlayerID]++
	}
	room.GamesPlayed++

	// Build game ended event
	var winnerID, winnerName string
	var moves []MovePayload

	if winner != nil {
		winnerID = winner.PlayerID
		winnerName = room.GetPlayerName(winner.PlayerID)
		moves = make([]MovePayload, len(winner.Moves))
		for i, move := range winner.Moves {
			moves[i] = MovePayload{
				RobotId: int(move.Id),
				X:       int(move.Pos.X),
				Y:       int(move.Pos.Y),
			}
		}
	}

	signals := []Signal{
		BroadcastSignal{Event: GameEndedEvent{
			RoomID:     room.ID,
			WinnerID:   winnerID,
			WinnerName: winnerName,
			Moves:      moves,
		}},
	}

	return signals
}

func (gl *gameLifecycle) StartNextGame(room *Room) []Signal {
	// Get winning game state for continuation (wins already credited in EndGame)
	var winningGameState *model.Game
	if room.CurrentGame != nil && len(room.Solutions) > 0 {
		winningSolution := gl.solutionMgr.GetWinningSolution(room.Solutions)
		if winningSolution != nil {
			// Apply winning moves to get final robot positions
			if len(winningSolution.Moves) > 0 {
				_, winningGameState = room.CurrentGame.CheckSolution(winningSolution.Moves)
			}
		}
	}

	// Generate next game
	var game *model.Game
	if winningGameState != nil {
		game = model.NewContinuationGame(winningGameState)
	} else if room.CurrentGame != nil {
		game = model.NewContinuationGame(room.CurrentGame)
	} else {
		game = model.NewRandomGame()
	}
	now := time.Now()

	room.CurrentGame = game
	room.GameStartedAt = &now
	room.ClearGameState()

	signals := []Signal{
		BroadcastSignal{Event: GameStartedEvent{RoomID: room.ID}},
	}

	return signals
}
