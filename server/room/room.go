// Package room provides multiplayer game room management.
package room

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/srsalisbury/bouncebot/model"
	pb "github.com/srsalisbury/bouncebot/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Player represents a player in a room.
type Player struct {
	ID             string
	Name           string
	Status         PlayerStatus
	DisconnectedAt time.Time
}

// PlayerStatus represents the connection status of a player.
type PlayerStatus string

const (
	// PlayerStatusConnected means the player is actively connected.
	PlayerStatusConnected PlayerStatus = "connected"
	// PlayerStatusDisconnected means the player has disconnected but is within the grace period.
	PlayerStatusDisconnected PlayerStatus = "disconnected"
)

// PlayerSolution represents a player's solution to the current game.
type PlayerSolution struct {
	PlayerID string
	SolvedAt time.Time
	Moves    []model.BotPosition // The actual moves that solved the puzzle
}

// MoveCount returns the number of moves in the solution.
func (s *PlayerSolution) MoveCount() int {
	return len(s.Moves)
}

// PlayerSolutionHistory tracks all solutions a player has found (for restoring after retraction).
type PlayerSolutionHistory struct {
	PlayerID  string
	Solutions []PlayerSolution
}

// Room represents a multiplayer game room.
type Room struct {
	ID              string
	Players         []Player
	CreatedAt       time.Time
	LastActivityAt  time.Time               // Last user action timestamp (for cleanup)
	CurrentGame     *model.Game
	GameStartedAt   *time.Time
	Solutions       []PlayerSolution        // Current best solution per player
	SolutionHistory []PlayerSolutionHistory // All solutions per player (for retraction)
	Wins            map[string]int          // Wins per player ID
	GamesPlayed     int                     // Total games completed in room
	FinishedSolving []string                // Player IDs who are finished solving (triggers game end)
	ReadyForNext    []string                // Player IDs who are ready for next game
}

// GetPlayerName returns the name of the player with the given ID, or empty string if not found.
func (r *Room) GetPlayerName(playerID string) string {
	for _, p := range r.Players {
		if p.ID == playerID {
			return p.Name
		}
	}
	return ""
}

// ToProto converts a Room to its protobuf representation.
func (r *Room) ToProto() *pb.Room {
	players := make([]*pb.Player, len(r.Players))
	for i, p := range r.Players {
		players[i] = &pb.Player{
			Id:   p.ID,
			Name: p.Name,
		}
	}

	solutions := make([]*pb.PlayerSolution, len(r.Solutions))
	for i, sol := range r.Solutions {
		moves := make([]*pb.BotPos, len(sol.Moves))
		for j, move := range sol.Moves {
			moves[j] = move.ToProto()
		}
		solutions[i] = &pb.PlayerSolution{
			PlayerId: sol.PlayerID,
			SolvedAt: timestamppb.New(sol.SolvedAt),
			Moves:    moves,
		}
	}

	// Convert wins map to proto
	scores := make([]*pb.PlayerScore, 0, len(r.Wins))
	for playerID, wins := range r.Wins {
		scores = append(scores, &pb.PlayerScore{
			PlayerId: playerID,
			Wins:     int32(wins),
		})
	}

	room := &pb.Room{
		Id:              r.ID,
		Players:         players,
		CreatedAt:       timestamppb.New(r.CreatedAt),
		Solutions:       solutions,
		Scores:          scores,
		GamesPlayed:     int32(r.GamesPlayed),
		FinishedSolving: r.FinishedSolving,
		ReadyForNext:    r.ReadyForNext,
	}

	if r.CurrentGame != nil {
		room.CurrentGame = r.CurrentGame.ToProto()
	}

	if r.GameStartedAt != nil {
		room.GameStartedAt = timestamppb.New(*r.GameStartedAt)
	}

	return room
}

// MovePayload represents a single move for WebSocket broadcast.
type MovePayload struct {
	RobotId int `json:"robotId"`
	X       int `json:"x"`
	Y       int `json:"y"`
}

// EventBroadcaster is an interface for broadcasting room events.
type EventBroadcaster interface {
	BroadcastPlayerJoined(roomID, playerID, playerName string)
	BroadcastPlayerLeft(roomID, playerID string)
	BroadcastGameStarted(roomID string)
	BroadcastPlayerFinishedSolving(roomID, playerID string)
	BroadcastPlayerReadyForNext(roomID, playerID string)
	BroadcastPlayerSolved(roomID, playerID string, moveCount int)
	BroadcastSolutionRetracted(roomID, playerID string)
	BroadcastGameEnded(roomID, winnerID, winnerName string, moves []MovePayload)
}

// Store manages rooms in memory.
type Store struct {
	mu                    sync.RWMutex
	rooms                 map[string]*Room
	broadcaster           EventBroadcaster
	timers                map[string]*time.Timer // playerID -> timer
	disconnectGracePeriod time.Duration
}

// NewStore creates a new room store.
func NewStore() *Store {
	return &Store{
		rooms:                 make(map[string]*Room),
		timers:                make(map[string]*time.Timer),
		disconnectGracePeriod: 30 * time.Second, // default
	}
}

// SetBroadcaster sets the event broadcaster for the store.
func (store *Store) SetBroadcaster(b EventBroadcaster) {
	store.broadcaster = b
}

// SetDisconnectGracePeriod sets the grace period for player disconnection.
func (store *Store) SetDisconnectGracePeriod(d time.Duration) {
	store.disconnectGracePeriod = d
}

// roomIDChars is the character set for room IDs (no 0, 1, I, O to avoid confusion)
const roomIDChars = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// generateRoomID creates a random 4-character room ID.
func generateRoomID() string {
	result := make([]byte, 4)
	for i := range result {
		result[i] = roomIDChars[rand.IntN(len(roomIDChars))]
	}
	return string(result)
}

// generatePlayerID creates a random player ID.
func generatePlayerID() string {
	return fmt.Sprintf("%016x", rand.Uint64())
}

// Create creates a new room with the given player.
func (store *Store) Create(playerName string) *Room {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Generate unique room ID (retry if collision)
	roomID := generateRoomID()
	for store.rooms[roomID] != nil {
		roomID = generateRoomID()
	}

	playerID := generatePlayerID()
	now := time.Now()

	room := &Room{
		ID: roomID,
		Players: []Player{
			{ID: playerID, Name: playerName, Status: PlayerStatusConnected},
		},
		CreatedAt:      now,
		LastActivityAt: now,
		Wins:           make(map[string]int),
	}

	store.rooms[roomID] = room
	return room
}

// Join adds a player to an existing room.
func (store *Store) Join(roomID, playerName string) (*Room, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Normalize room ID to uppercase for case-insensitive matching
	normalizedID := strings.ToUpper(roomID)
	room, ok := store.rooms[normalizedID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	playerID := generatePlayerID()
	room.Players = append(room.Players, Player{
		ID:     playerID,
		Name:   playerName,
		Status: PlayerStatusConnected,
	})
	room.LastActivityAt = time.Now()

	// Broadcast player joined event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerJoined(roomID, playerID, playerName)
	}

	return room, nil
}

// Get retrieves a room by ID.
func (store *Store) Get(roomID string) (*Room, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	// Normalize room ID to uppercase for case-insensitive matching
	normalizedID := strings.ToUpper(roomID)
	room, ok := store.rooms[normalizedID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	return room, nil
}

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

// getWinningSolution finds the winning solution (lowest moves, earliest time as tiebreaker)
func (store *Store) getWinningSolution(solutions []PlayerSolution) *PlayerSolution {
	if len(solutions) == 0 {
		return nil
	}

	best := &solutions[0]
	for i := range solutions[1:] {
		sol := &solutions[i+1]
		if sol.MoveCount() < best.MoveCount() {
			best = sol
		} else if sol.MoveCount() == best.MoveCount() && sol.SolvedAt.Before(best.SolvedAt) {
			best = sol
		}
	}
	return best
}

// SubmitSolution records a player's solution for the current game.
// If moves are provided, they are verified against the current game state.
func (store *Store) SubmitSolution(roomID, playerID string, moves []model.BotPosition) (*PlayerSolution, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	if room.CurrentGame == nil {
		return nil, fmt.Errorf("no game in progress")
	}

	// Verify player exists
	if room.GetPlayerName(playerID) == "" {
		return nil, fmt.Errorf("player not found: %s", playerID)
	}

	// Verify the solution
	isValid, _ := room.CurrentGame.CheckSolution(moves)
	if !isValid {
		return nil, fmt.Errorf("invalid solution")
	}

	moveCount := len(moves)
	now := time.Now()
	room.LastActivityAt = now

	// Add to solution history
	store.addToHistory(room, playerID, moves, now)

	// Check if player already submitted a solution for this game
	for i := range room.Solutions {
		if room.Solutions[i].PlayerID == playerID {
			// Update if better solution
			if moveCount < room.Solutions[i].MoveCount() {
				room.Solutions[i].SolvedAt = now
				room.Solutions[i].Moves = moves
				// Broadcast updated solution
				if store.broadcaster != nil {
					store.broadcaster.BroadcastPlayerSolved(roomID, playerID, moveCount)
				}
			}
			// Return existing (possibly updated) solution
			return &room.Solutions[i], nil
		}
	}

	solution := PlayerSolution{
		PlayerID: playerID,
		SolvedAt: now,
		Moves:    moves,
	}
	room.Solutions = append(room.Solutions, solution)

	// Broadcast player solved event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerSolved(roomID, playerID, moveCount)
	}

	return &solution, nil
}

// addToHistory adds a solution to the player's history (if not already present with same move count).
func (store *Store) addToHistory(room *Room, playerID string, moves []model.BotPosition, solvedAt time.Time) {
	// Find or create history entry for this player
	var history *PlayerSolutionHistory
	for i := range room.SolutionHistory {
		if room.SolutionHistory[i].PlayerID == playerID {
			history = &room.SolutionHistory[i]
			break
		}
	}
	if history == nil {
		room.SolutionHistory = append(room.SolutionHistory, PlayerSolutionHistory{
			PlayerID: playerID,
		})
		history = &room.SolutionHistory[len(room.SolutionHistory)-1]
	}

	// Check if we already have this move count in history
	moveCount := len(moves)
	for _, sol := range history.Solutions {
		if sol.MoveCount() == moveCount {
			return // Already have this solution
		}
	}

	// Add to history
	history.Solutions = append(history.Solutions, PlayerSolution{
		PlayerID: playerID,
		SolvedAt: solvedAt,
		Moves:    moves,
	})
}

// RetractSolution removes a player's current best solution and restores the previous one from history.
func (store *Store) RetractSolution(roomID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	if room.CurrentGame == nil {
		return fmt.Errorf("no game in progress")
	}

	room.LastActivityAt = time.Now()

	// Find the player's current solution
	var currentMoveCount int
	var solutionIndex int = -1
	for i, sol := range room.Solutions {
		if sol.PlayerID == playerID {
			currentMoveCount = sol.MoveCount()
			solutionIndex = i
			break
		}
	}

	if solutionIndex == -1 {
		return fmt.Errorf("no solution found for player: %s", playerID)
	}

	// Find the player's history and remove the current solution from it
	var history *PlayerSolutionHistory
	for i := range room.SolutionHistory {
		if room.SolutionHistory[i].PlayerID == playerID {
			history = &room.SolutionHistory[i]
			break
		}
	}

	// Remove current move count from history
	if history != nil {
		for i, sol := range history.Solutions {
			if sol.MoveCount() == currentMoveCount {
				history.Solutions = append(history.Solutions[:i], history.Solutions[i+1:]...)
				break
			}
		}

		// Find the next best solution in history (smallest move count remaining)
		if len(history.Solutions) > 0 {
			bestIdx := 0
			for i, sol := range history.Solutions {
				if sol.MoveCount() < history.Solutions[bestIdx].MoveCount() {
					bestIdx = i
				}
			}
			// Restore the previous best solution
			room.Solutions[solutionIndex] = history.Solutions[bestIdx]

			// Broadcast the restored solution
			if store.broadcaster != nil {
				store.broadcaster.BroadcastPlayerSolved(roomID, playerID, history.Solutions[bestIdx].MoveCount())
			}
			return nil
		}
	}

	// No previous solution to restore - remove completely
	room.Solutions = append(room.Solutions[:solutionIndex], room.Solutions[solutionIndex+1:]...)

	// Broadcast solution retracted event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastSolutionRetracted(roomID, playerID)
	}

	return nil
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

// DisconnectPlayer marks a player as disconnected and starts a timer to remove them.
func (store *Store) DisconnectPlayer(roomID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	var player *Player
	for i := range room.Players {
		if room.Players[i].ID == playerID {
			player = &room.Players[i]
			break
		}
	}

	if player == nil {
		// Player might have been removed already, which is fine
		return nil
	}

	player.Status = PlayerStatusDisconnected
	player.DisconnectedAt = time.Now()

	// Clean up any existing timer for this player
	if oldTimer, ok := store.timers[playerID]; ok {
		oldTimer.Stop()
		delete(store.timers, playerID)
	}

	timer := time.AfterFunc(store.disconnectGracePeriod, func() {
		store.RemovePlayer(roomID, playerID)
	})
	store.timers[playerID] = timer

	return nil
}

// ReconnectPlayer marks a player as connected and cancels their removal timer.
func (store *Store) ReconnectPlayer(roomID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	var player *Player
	for i := range room.Players {
		if room.Players[i].ID == playerID {
			player = &room.Players[i]
			break
		}
	}

	if player == nil {
		return fmt.Errorf("player not found: %s", playerID)
	}

	if player.Status == PlayerStatusDisconnected {
		player.Status = PlayerStatusConnected
		if timer, ok := store.timers[playerID]; ok {
			timer.Stop()
			delete(store.timers, playerID)
		}
	}

	return nil
}

// RemovePlayer removes a player from a room if they are still disconnected.
func (store *Store) RemovePlayer(roomID, playerID string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.rooms[roomID]
	if !ok {
		return // room already gone
	}

	var playerIndex = -1
	for i, p := range room.Players {
		if p.ID == playerID {
			playerIndex = i
			break
		}
	}

	if playerIndex != -1 {
		// Only remove if still disconnected
		if room.Players[playerIndex].Status == PlayerStatusDisconnected {
			room.Players = append(room.Players[:playerIndex], room.Players[playerIndex+1:]...)
			delete(store.timers, playerID)

			// Remove from FinishedSolving
			for i, id := range room.FinishedSolving {
				if id == playerID {
					room.FinishedSolving = append(room.FinishedSolving[:i], room.FinishedSolving[i+1:]...)
					break
				}
			}

			// Remove from ReadyForNext
			for i, id := range room.ReadyForNext {
				if id == playerID {
					room.ReadyForNext = append(room.ReadyForNext[:i], room.ReadyForNext[i+1:]...)
					break
				}
			}

			// Remove from Solutions
			for i, sol := range room.Solutions {
				if sol.PlayerID == playerID {
					room.Solutions = append(room.Solutions[:i], room.Solutions[i+1:]...)
					break
				}
			}

			if store.broadcaster != nil {
				store.broadcaster.BroadcastPlayerLeft(roomID, playerID)
			}

			// Check if remaining players now satisfy conditions
			if len(room.Players) > 0 {
				// If game is active and all remaining players are finished, end the game
				if room.CurrentGame != nil && len(room.FinishedSolving) == len(room.Players) {
					store.endGame(room)
				}
				// If all remaining players are ready for next, start next game
				if len(room.ReadyForNext) == len(room.Players) {
					store.startNextGame(room)
				}
			}
		}
	}
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
