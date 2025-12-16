// Package session provides multiplayer game session management.
package session

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

// Player represents a player in a session.
type Player struct {
	ID   string
	Name string
}

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

// Session represents a multiplayer game session.
type Session struct {
	ID              string
	Players         []Player
	CreatedAt       time.Time
	LastActivityAt  time.Time               // Last user action timestamp (for cleanup)
	CurrentGame     *model.Game
	GameStartedAt   *time.Time
	Solutions       []PlayerSolution        // Current best solution per player
	SolutionHistory []PlayerSolutionHistory // All solutions per player (for retraction)
	Wins            map[string]int          // Wins per player ID
	GamesPlayed     int                     // Total games completed in session
	FinishedSolving []string                // Player IDs who are finished solving (triggers game end)
	ReadyForNext    []string                // Player IDs who are ready for next game
}

// GetPlayerName returns the name of the player with the given ID, or empty string if not found.
func (s *Session) GetPlayerName(playerID string) string {
	for _, p := range s.Players {
		if p.ID == playerID {
			return p.Name
		}
	}
	return ""
}

// ToProto converts a Session to its protobuf representation.
func (s *Session) ToProto() *pb.Session {
	players := make([]*pb.Player, len(s.Players))
	for i, p := range s.Players {
		players[i] = &pb.Player{
			Id:   p.ID,
			Name: p.Name,
		}
	}

	solutions := make([]*pb.PlayerSolution, len(s.Solutions))
	for i, sol := range s.Solutions {
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
	scores := make([]*pb.PlayerScore, 0, len(s.Wins))
	for playerID, wins := range s.Wins {
		scores = append(scores, &pb.PlayerScore{
			PlayerId: playerID,
			Wins:     int32(wins),
		})
	}

	session := &pb.Session{
		Id:              s.ID,
		Players:         players,
		CreatedAt:       timestamppb.New(s.CreatedAt),
		Solutions:       solutions,
		Scores:          scores,
		GamesPlayed:     int32(s.GamesPlayed),
		FinishedSolving: s.FinishedSolving,
		ReadyForNext:    s.ReadyForNext,
	}

	if s.CurrentGame != nil {
		session.CurrentGame = s.CurrentGame.ToProto()
	}

	if s.GameStartedAt != nil {
		session.GameStartedAt = timestamppb.New(*s.GameStartedAt)
	}

	return session
}

// MovePayload represents a single move for WebSocket broadcast.
type MovePayload struct {
	RobotId int `json:"robotId"`
	X       int `json:"x"`
	Y       int `json:"y"`
}

// EventBroadcaster is an interface for broadcasting session events.
type EventBroadcaster interface {
	BroadcastPlayerJoined(sessionID, playerID, playerName string)
	BroadcastGameStarted(sessionID string)
	BroadcastPlayerFinishedSolving(sessionID, playerID string)
	BroadcastPlayerReadyForNext(sessionID, playerID string)
	BroadcastPlayerSolved(sessionID, playerID string, moveCount int)
	BroadcastSolutionRetracted(sessionID, playerID string)
	BroadcastGameEnded(sessionID, winnerID, winnerName string, moves []MovePayload)
}

// Store manages sessions in memory.
type Store struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	broadcaster EventBroadcaster
}

// NewStore creates a new session store.
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*Session),
	}
}

// SetBroadcaster sets the event broadcaster for the store.
func (store *Store) SetBroadcaster(b EventBroadcaster) {
	store.broadcaster = b
}

// sessionIDChars is the character set for session IDs (no 0, 1, I, O to avoid confusion)
const sessionIDChars = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// generateSessionID creates a random 4-character session ID.
func generateSessionID() string {
	result := make([]byte, 4)
	for i := range result {
		result[i] = sessionIDChars[rand.IntN(len(sessionIDChars))]
	}
	return string(result)
}

// generatePlayerID creates a random player ID.
func generatePlayerID() string {
	return fmt.Sprintf("%016x", rand.Uint64())
}

// Create creates a new session with the given player.
func (store *Store) Create(playerName string) *Session {
	store.mu.Lock()
	defer store.mu.Unlock()

	sessionID := generateSessionID()
	playerID := generatePlayerID()
	now := time.Now()

	session := &Session{
		ID: sessionID,
		Players: []Player{
			{ID: playerID, Name: playerName},
		},
		CreatedAt:      now,
		LastActivityAt: now,
		Wins:           make(map[string]int),
	}

	store.sessions[sessionID] = session
	return session
}

// Join adds a player to an existing session.
func (store *Store) Join(sessionID, playerName string) (*Session, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Normalize session ID to uppercase for case-insensitive matching
	normalizedID := strings.ToUpper(sessionID)
	session, ok := store.sessions[normalizedID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	playerID := generatePlayerID()
	session.Players = append(session.Players, Player{
		ID:   playerID,
		Name: playerName,
	})
	session.LastActivityAt = time.Now()

	// Broadcast player joined event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerJoined(sessionID, playerID, playerName)
	}

	return session, nil
}

// Get retrieves a session by ID.
func (store *Store) Get(sessionID string) (*Session, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	// Normalize session ID to uppercase for case-insensitive matching
	normalizedID := strings.ToUpper(sessionID)
	session, ok := store.sessions[normalizedID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// StartGame starts a new game in the session.
// If useFixedBoard is true, uses the fixed Game1() configuration instead of random.
// If there's an existing game, continues with same board/robots but new target.
// Robot positions are taken from the winning solution's final state.
func (store *Store) StartGame(sessionID string, useFixedBoard bool) (*Session, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// If there was a previous game with solutions, determine and record the winner
	// and get the final game state from the winning solution
	var winningGameState *model.Game
	if session.CurrentGame != nil && len(session.Solutions) > 0 {
		winningSolution := store.getWinningSolution(session.Solutions)
		if winningSolution != nil {
			session.Wins[winningSolution.PlayerID]++
			// Apply winning moves to get final robot positions
			if len(winningSolution.Moves) > 0 {
				_, winningGameState = session.CurrentGame.CheckSolution(winningSolution.Moves)
			}
		}
		session.GamesPlayed++
	}

	// Generate game
	var game *model.Game
	if useFixedBoard {
		game = model.Game1()
	} else if winningGameState != nil {
		// Continue from winning game state: same board, robots at final positions
		game = model.NewContinuationGame(winningGameState)
	} else if session.CurrentGame != nil {
		// No winning solution with moves, continue from existing game
		game = model.NewContinuationGame(session.CurrentGame)
	} else {
		// First game: fully random
		game = model.NewRandomGame()
	}
	now := time.Now()

	session.CurrentGame = game
	session.GameStartedAt = &now
	session.LastActivityAt = now
	session.Solutions = nil         // Clear solutions for new game
	session.SolutionHistory = nil   // Clear history for new game
	session.FinishedSolving = nil   // Clear finished players for new game
	session.ReadyForNext = nil      // Clear ready players for new game

	// Broadcast game started event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastGameStarted(sessionID)
	}

	return session, nil
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
func (store *Store) SubmitSolution(sessionID, playerID string, moves []model.BotPosition) (*PlayerSolution, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if session.CurrentGame == nil {
		return nil, fmt.Errorf("no game in progress")
	}

	// Verify player exists
	if session.GetPlayerName(playerID) == "" {
		return nil, fmt.Errorf("player not found: %s", playerID)
	}

	// Verify the solution
	isValid, _ := session.CurrentGame.CheckSolution(moves)
	if !isValid {
		return nil, fmt.Errorf("invalid solution")
	}

	moveCount := len(moves)
	now := time.Now()
	session.LastActivityAt = now

	// Add to solution history
	store.addToHistory(session, playerID, moves, now)

	// Check if player already submitted a solution for this game
	for i := range session.Solutions {
		if session.Solutions[i].PlayerID == playerID {
			// Update if better solution
			if moveCount < session.Solutions[i].MoveCount() {
				session.Solutions[i].SolvedAt = now
				session.Solutions[i].Moves = moves
				// Broadcast updated solution
				if store.broadcaster != nil {
					store.broadcaster.BroadcastPlayerSolved(sessionID, playerID, moveCount)
				}
			}
			// Return existing (possibly updated) solution
			return &session.Solutions[i], nil
		}
	}

	solution := PlayerSolution{
		PlayerID: playerID,
		SolvedAt: now,
		Moves:    moves,
	}
	session.Solutions = append(session.Solutions, solution)

	// Broadcast player solved event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerSolved(sessionID, playerID, moveCount)
	}

	return &solution, nil
}

// addToHistory adds a solution to the player's history (if not already present with same move count).
func (store *Store) addToHistory(session *Session, playerID string, moves []model.BotPosition, solvedAt time.Time) {
	// Find or create history entry for this player
	var history *PlayerSolutionHistory
	for i := range session.SolutionHistory {
		if session.SolutionHistory[i].PlayerID == playerID {
			history = &session.SolutionHistory[i]
			break
		}
	}
	if history == nil {
		session.SolutionHistory = append(session.SolutionHistory, PlayerSolutionHistory{
			PlayerID: playerID,
		})
		history = &session.SolutionHistory[len(session.SolutionHistory)-1]
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
func (store *Store) RetractSolution(sessionID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.CurrentGame == nil {
		return fmt.Errorf("no game in progress")
	}

	session.LastActivityAt = time.Now()

	// Find the player's current solution
	var currentMoveCount int
	var solutionIndex int = -1
	for i, sol := range session.Solutions {
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
	for i := range session.SolutionHistory {
		if session.SolutionHistory[i].PlayerID == playerID {
			history = &session.SolutionHistory[i]
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
			session.Solutions[solutionIndex] = history.Solutions[bestIdx]

			// Broadcast the restored solution
			if store.broadcaster != nil {
				store.broadcaster.BroadcastPlayerSolved(sessionID, playerID, history.Solutions[bestIdx].MoveCount())
			}
			return nil
		}
	}

	// No previous solution to restore - remove completely
	session.Solutions = append(session.Solutions[:solutionIndex], session.Solutions[solutionIndex+1:]...)

	// Broadcast solution retracted event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastSolutionRetracted(sessionID, playerID)
	}

	return nil
}

// MarkFinishedSolving marks a player as finished looking for solutions.
func (store *Store) MarkFinishedSolving(sessionID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.CurrentGame == nil {
		return fmt.Errorf("no game in progress")
	}

	// Verify player exists
	if session.GetPlayerName(playerID) == "" {
		return fmt.Errorf("player not found: %s", playerID)
	}

	session.LastActivityAt = time.Now()

	// Check if already finished
	for _, id := range session.FinishedSolving {
		if id == playerID {
			return nil // Already finished
		}
	}

	session.FinishedSolving = append(session.FinishedSolving, playerID)

	// Broadcast player finished solving event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerFinishedSolving(sessionID, playerID)
	}

	// Check if all players are finished
	if len(session.FinishedSolving) == len(session.Players) {
		store.endGame(session)
	}

	return nil
}

// MarkReadyForNext marks a player as ready for the next game.
func (store *Store) MarkReadyForNext(sessionID, playerID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	session, ok := store.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Verify player exists
	if session.GetPlayerName(playerID) == "" {
		return fmt.Errorf("player not found: %s", playerID)
	}

	session.LastActivityAt = time.Now()

	// Check if already ready
	for _, id := range session.ReadyForNext {
		if id == playerID {
			return nil // Already ready
		}
	}

	session.ReadyForNext = append(session.ReadyForNext, playerID)

	// Broadcast player ready for next event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastPlayerReadyForNext(sessionID, playerID)
	}

	// Check if all players are ready - auto-start next game
	if len(session.ReadyForNext) == len(session.Players) {
		store.startNextGame(session)
	}

	return nil
}

// startNextGame starts the next game when all players are ready.
func (store *Store) startNextGame(session *Session) {
	// Get winning game state for continuation (wins already credited in endGame)
	var winningGameState *model.Game
	if session.CurrentGame != nil && len(session.Solutions) > 0 {
		winningSolution := store.getWinningSolution(session.Solutions)
		if winningSolution != nil {
			// Apply winning moves to get final robot positions
			if len(winningSolution.Moves) > 0 {
				_, winningGameState = session.CurrentGame.CheckSolution(winningSolution.Moves)
			}
		}
	}

	// Generate next game
	var game *model.Game
	if winningGameState != nil {
		game = model.NewContinuationGame(winningGameState)
	} else if session.CurrentGame != nil {
		game = model.NewContinuationGame(session.CurrentGame)
	} else {
		game = model.NewRandomGame()
	}
	now := time.Now()

	session.CurrentGame = game
	session.GameStartedAt = &now
	session.Solutions = nil
	session.SolutionHistory = nil
	session.FinishedSolving = nil
	session.ReadyForNext = nil

	// Broadcast game started event
	if store.broadcaster != nil {
		store.broadcaster.BroadcastGameStarted(session.ID)
	}
}

// endGame determines the winner, credits the win, and broadcasts game_ended event.
func (store *Store) endGame(session *Session) {
	// Credit the win and increment games played
	winner := store.getWinningSolution(session.Solutions)
	if winner != nil {
		session.Wins[winner.PlayerID]++
	}
	session.GamesPlayed++

	if store.broadcaster == nil {
		return
	}

	// Broadcast game ended
	if winner != nil {
		winnerName := session.GetPlayerName(winner.PlayerID)
		// Convert moves to MovePayload format
		moves := make([]MovePayload, len(winner.Moves))
		for i, move := range winner.Moves {
			moves[i] = MovePayload{
				RobotId: int(move.Id),
				X:       int(move.Pos.X),
				Y:       int(move.Pos.Y),
			}
		}
		store.broadcaster.BroadcastGameEnded(session.ID, winner.PlayerID, winnerName, moves)
	} else {
		// No solutions submitted - no winner
		store.broadcaster.BroadcastGameEnded(session.ID, "", "", nil)
	}
}
