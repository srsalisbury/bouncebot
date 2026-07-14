package room

import (
	"fmt"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// GameLifecycle manages game state transitions.
type GameLifecycle interface {
	// NewGameSource inspects room.CurrentGame/Solutions/Settings (the winning
	// solution/continuation logic shared by starting the first game and every
	// continuation round) and returns a function that produces a new candidate
	// *model.Game on each call, plus the room's configured minimum solution
	// length. Read-only — does not mutate room. Must be called with the room
	// lock held, but the returned candidateFn is safe to call repeatedly
	// WITHOUT the lock: model.NewContinuationGame/NewRandomGame never mutate
	// their inputs.
	NewGameSource(room *Room) (candidateFn func() *model.Game, minLength int)

	// CommitNewGame finalizes game as the room's new current game: sets
	// CurrentGame, GameStartedAt, LastActivityAt, and clears round-scoped state
	// via ClearGameState. Returns broadcast signals. Must be called with the
	// room lock held.
	CommitNewGame(room *Room, game *model.Game) []Signal

	// PromotePendingPlayers moves room.PendingPlayers into room.Players. Called
	// as part of starting the next game in a room. Must be called with the room
	// lock held.
	PromotePendingPlayers(room *Room)

	// MarkFinishedSolving marks a player as finished solving.
	// Returns signals or error.
	MarkFinishedSolving(room *Room, playerID string) ([]Signal, error)

	// MarkReadyForNext marks a player as ready for the next game.
	// Returns signals or error.
	MarkReadyForNext(room *Room, playerID string) ([]Signal, error)

	// EndGame ends the current game and determines the winner.
	// Returns signals.
	EndGame(room *Room) []Signal
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

func (gl *gameLifecycle) NewGameSource(room *Room) (func() *model.Game, int) {
	// If there was a previous game with solutions, the winning solution's final
	// robot positions are what the next board should continue from.
	// Determine winning game state for continuation (wins are credited in EndGame only)
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

	// No player won: fall back to the solver's (BBot's) solution end state, if
	// one has finished in time. Otherwise (rare - the solver hasn't completed
	// yet), leave the robots where they started, same as before this fallback
	// existed.
	if winningGameState == nil && room.CurrentGame != nil {
		if solverResult := bestSolverResult(room.SolverResults); solverResult != nil {
			moves := movePayloadsToBotPositions(solverResult.Moves)
			_, winningGameState = room.CurrentGame.CheckSolution(moves)
		}
	}

	var candidateFn func() *model.Game
	switch {
	case winningGameState != nil:
		// Continue from winning game state: same board, robots at final positions
		candidateFn = func() *model.Game { return model.NewContinuationGame(winningGameState) }
	case room.CurrentGame != nil:
		// No winning solution with moves, continue from existing game
		candidateFn = func() *model.Game { return model.NewContinuationGame(room.CurrentGame) }
	default:
		// First game: fully random
		candidateFn = model.NewRandomGame
	}

	return candidateFn, room.Settings.MinSolutionLength
}

// bestSolverResult returns the shortest completed solver solution in results,
// or nil if none has finished successfully yet. Ties break on solver name for
// determinism.
func bestSolverResult(results map[string]*SolverResult) *SolverResult {
	var best *SolverResult
	for _, r := range results {
		if !r.Completed || len(r.Moves) == 0 {
			continue
		}
		if best == nil || len(r.Moves) < len(best.Moves) ||
			(len(r.Moves) == len(best.Moves) && r.SolverName < best.SolverName) {
			best = r
		}
	}
	return best
}

// movePayloadsToBotPositions converts WebSocket-broadcast move payloads back
// into the model.BotPosition form Game.CheckSolution expects.
func movePayloadsToBotPositions(moves []MovePayload) []model.BotPosition {
	result := make([]model.BotPosition, len(moves))
	for i, m := range moves {
		result[i] = model.BotPosition{
			Id:  model.BotId(m.RobotId),
			Pos: model.Position{X: model.BoardDim(m.X), Y: model.BoardDim(m.Y)},
		}
	}
	return result
}

func (gl *gameLifecycle) CommitNewGame(room *Room, game *model.Game) []Signal {
	now := time.Now()

	room.CurrentGame = game
	room.GameStartedAt = &now
	room.LastActivityAt = now
	room.ClearGameState()

	return []Signal{
		BroadcastSignal{Event: GameStartedEvent{RoomID: room.ID}},
	}
}

func (gl *gameLifecycle) PromotePendingPlayers(room *Room) {
	if len(room.PendingPlayers) > 0 {
		room.Players = append(room.Players, room.PendingPlayers...)
		room.PendingPlayers = nil
	}
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
