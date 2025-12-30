package room

import (
	"fmt"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// StartGame starts a new game in the room.
// If useFixedBoard is true, uses the fixed Game1() configuration instead of random.
// If there's an existing game, continues with same board/robots but new target.
// Robot positions are taken from the winning solution's final state.
func (store *Store) StartGame(roomID string, useFixedBoard bool) (*Room, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	// If there was a previous game with solutions, determine and record the winner
	// and get the final game state from the winning solution
	var winningGameState *model.Game
	if room.CurrentGame != nil && len(room.Solutions) > 0 {
		winningSolution := store.getWinningSolution(room.Solutions)
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
	room.Solutions = nil         // Clear solutions for new game
	room.SolutionHistory = nil   // Clear history for new game
	room.FinishedSolving = nil   // Clear finished players for new game
	room.ReadyForNext = nil      // Clear ready players for new game

	// Broadcast game started event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastGameStarted(roomID)
	}

	return room, nil
}

// MarkFinishedSolving marks a player as finished looking for solutions.
func (store *Store) MarkFinishedSolving(roomID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	if room.CurrentGame == nil {
		return fmt.Errorf("no game in progress")
	}

	// Verify player exists
	if room.GetPlayerName(playerID) == "" {
		return fmt.Errorf("player not found: %s", playerID)
	}

	room.LastActivityAt = time.Now()

	// Check if already finished
	for _, id := range room.FinishedSolving {
		if id == playerID {
			return nil // Already finished
		}
	}

	room.FinishedSolving = append(room.FinishedSolving, playerID)

	// Broadcast player finished solving event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerFinishedSolving(roomID, playerID)
	}

	// Check if all players are finished
	if len(room.FinishedSolving) == len(room.Players) {
		store.endGame(room)
	}

	return nil
}

// MarkReadyForNext marks a player as ready for the next game.
func (store *Store) MarkReadyForNext(roomID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	// Verify player exists
	if room.GetPlayerName(playerID) == "" {
		return fmt.Errorf("player not found: %s", playerID)
	}

	room.LastActivityAt = time.Now()

	// Check if already ready
	for _, id := range room.ReadyForNext {
		if id == playerID {
			return nil // Already ready
		}
	}

	room.ReadyForNext = append(room.ReadyForNext, playerID)

	// Broadcast player ready for next event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerReadyForNext(roomID, playerID)
	}

	// Check if all players are ready - auto-start next game
	if len(room.ReadyForNext) == len(room.Players) {
		store.startNextGame(room)
	}

	return nil
}

// startNextGame starts the next game when all players are ready.
func (store *Store) startNextGame(room *Room) {
	// Get winning game state for continuation (wins already credited in endGame)
	var winningGameState *model.Game
	if room.CurrentGame != nil && len(room.Solutions) > 0 {
		winningSolution := store.getWinningSolution(room.Solutions)
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
	room.Solutions = nil
	room.SolutionHistory = nil
	room.FinishedSolving = nil
	room.ReadyForNext = nil

	// Broadcast game started event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastGameStarted(room.ID)
	}
}

// endGame determines the winner, credits the win, and broadcasts game_ended event.
func (store *Store) endGame(room *Room) {
	// Credit the win and increment games played
	winner := store.getWinningSolution(room.Solutions)
	if winner != nil {
		room.Wins[winner.PlayerID]++
	}
	room.GamesPlayed++

	if store.broadcaster == nil {
		return
	}

	// Broadcast game ended
	if winner != nil {
		winnerName := room.GetPlayerName(winner.PlayerID)
		// Convert moves to MovePayload format
		moves := make([]MovePayload, len(winner.Moves))
		for i, move := range winner.Moves {
			moves[i] = MovePayload{
				RobotId: int(move.Id),
				X:       int(move.Pos.X),
				Y:       int(move.Pos.Y),
			}
		}
		store.broadcaster.BroadcastGameEnded(room.ID, winner.PlayerID, winnerName, moves)
	} else {
		// No solutions submitted - no winner
		store.broadcaster.BroadcastGameEnded(room.ID, "", "", nil)
	}
}
