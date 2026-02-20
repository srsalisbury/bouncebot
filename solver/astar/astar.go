package astar

import (
	"context"
	"fmt"
	"sort"

	"github.com/srsalisbury/bouncebot/model"
	"github.com/srsalisbury/bouncebot/solver"
)

// AStarSolver implements a breadth-first search solver.
// It finds the optimal (shortest) solution.
type AStarSolver struct{}

// Name returns the solver name.
func (s *AStarSolver) Name() string {
	return "A-Star"
}

// stateNode is a node in the A* search tree.
type stateNode struct {
	bots      map[model.BotId]model.Position
	moves     []model.BotPosition
	heuristic int // h: estimated distance to goal
}

// priority returns the f-score for A* (f = g + h).
// g = number of moves so far, h = heuristic estimate to goal.
func (n *stateNode) priority() int {
	return len(n.moves) + n.heuristic
}

// stateEncoder encodes bot positions into a uint64 for use as a map key,
// avoiding the overhead of fmt.Sprintf string construction.
type stateEncoder struct {
	sortedBotIds []model.BotId
	bitsPerCoord int
}

// newStateEncoder creates a state encoder based on the bot IDs and board size.
// Supports boards up to 16x16 with 8 bots using 8 bits per bot (4 bits per coordinate).
func newStateEncoder(bots map[model.BotId]model.Position, boardSize model.BoardDim) *stateEncoder {
	ids := make([]model.BotId, 0, len(bots))
	for id := range bots {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	bitsPerCoord := 4 // supports coordinates 0-15
	if boardSize > 16 {
		bitsPerCoord = 7 // supports coordinates 0-127
	}

	return &stateEncoder{sortedBotIds: ids, bitsPerCoord: bitsPerCoord}
}

// encode packs all bot positions into a uint64.
func (e *stateEncoder) encode(bots map[model.BotId]model.Position) uint64 {
	var key uint64
	bitsPerBot := e.bitsPerCoord * 2
	for i, id := range e.sortedBotIds {
		pos := bots[id]
		shift := i * bitsPerBot
		key |= uint64(pos.X) << shift
		key |= uint64(pos.Y) << (shift + e.bitsPerCoord)
	}
	return key
}

// wallLookup provides O(1) wall-blocking checks by precomputing lookup sets
// from the game board, avoiding repeated linear scans and defensive copies
// of the wall slices.
type wallLookup struct {
	boardSize model.BoardDim
	hWalls    map[model.Position]bool
	vWalls    map[model.Position]bool
}

// newWallLookup precomputes wall lookup sets from the game board.
func newWallLookup(game *model.Game) *wallLookup {
	hWallSlice := game.Board.HWalls()
	vWallSlice := game.Board.VWalls()
	wl := &wallLookup{
		boardSize: game.Board.Size(),
		hWalls:    make(map[model.Position]bool, len(hWallSlice)),
		vWalls:    make(map[model.Position]bool, len(vWallSlice)),
	}
	for _, w := range hWallSlice {
		wl.hWalls[w] = true
	}
	for _, w := range vWallSlice {
		wl.vWalls[w] = true
	}
	return wl
}

// hasWallBlocking checks if there's a wall or board edge blocking movement from pos in dir.
func (wl *wallLookup) hasWallBlocking(pos model.Position, dir model.Direction) bool {
	switch dir {
	case model.Up:
		if pos.Y == 0 {
			return true
		}
		return wl.hWalls[model.Position{X: pos.X, Y: pos.Y - 1}]
	case model.Down:
		if pos.Y == wl.boardSize-1 {
			return true
		}
		return wl.hWalls[model.Position{X: pos.X, Y: pos.Y}]
	case model.Left:
		if pos.X == 0 {
			return true
		}
		return wl.vWalls[model.Position{X: pos.X - 1, Y: pos.Y}]
	case model.Right:
		if pos.X == wl.boardSize-1 {
			return true
		}
		return wl.vWalls[model.Position{X: pos.X, Y: pos.Y}]
	}
	return false
}

// Solve finds the optimal solution using A* search.
func (s *AStarSolver) Solve(ctx context.Context, game *model.Game) solver.Result {
	directions := []model.Direction{model.Up, model.Down, model.Left, model.Right}
	walls := newWallLookup(game)
	encoder := newStateEncoder(game.Bots, game.Board.Size())
	baseHeuristic := NewHeuristicTable(game, walls)

	// Initial state
	initialNode := &stateNode{
		bots:      copyBots(game.Bots),
		moves:     nil,
		heuristic: baseHeuristic.Get(game.Bots[game.Target.Id]),
	}

	// Check if already solved
	if isWin(initialNode.bots, game.Target) {
		return solver.Result{
			SolverName: s.Name(),
			Solution:   &solver.Solution{Moves: nil},
			Completed:  true,
		}
	}

	// AStar queue and closed set (states fully expanded)
	pqueue := NewPriorityQueue()
	pqueue.Enqueue(initialNode)
	closed := make(map[uint64]bool)

	checkCount := 0
	const checkInterval = 1000 // Check context every N iterations

	for pqueue.Len() > 0 {
		// Check for timeout periodically
		checkCount++
		if checkCount >= checkInterval {
			checkCount = 0
			select {
			case <-ctx.Done():
				return solver.Result{
					SolverName: s.Name(),
					Error:      ctx.Err(),
					Completed:  false,
				}
			default:
			}
		}

		// Dequeue
		current := pqueue.Dequeue()

		// Mark as closed (fully expanded) - only now, not when first seen
		currentKey := encoder.encode(current.bots)
		if closed[currentKey] {
			continue // Already expanded via a better path
		}
		closed[currentKey] = true

		// Try all moves: each bot in each direction
		for botId := range current.bots {
			for _, dir := range directions {
				// Compute destination
				dest, err := computeDestination(current.bots, walls, botId, dir)
				if err != nil {
					continue
				}

				// Skip if bot doesn't move
				if dest == current.bots[botId] {
					continue
				}

				// Create new state
				newBots := copyBots(current.bots)
				newBots[botId] = dest

				// Skip if already fully expanded
				stateKey := encoder.encode(newBots)
				if closed[stateKey] {
					continue
				}

				// Create new move
				move := model.BotPosition{Id: botId, Pos: dest}
				newMoves := append(append([]model.BotPosition{}, current.moves...), move)

				// Check for win
				if isWin(newBots, game.Target) {
					return solver.Result{
						SolverName: s.Name(),
						Solution:   &solver.Solution{Moves: newMoves},
						Completed:  true,
					}
				}

				// Enqueue
				pqueue.Enqueue(&stateNode{
					bots:      newBots,
					moves:     newMoves,
					heuristic: baseHeuristic.Get(newBots[game.Target.Id]),
				})
			}
		}
	}

	// No solution found
	return solver.Result{
		SolverName: s.Name(),
		Error:      fmt.Errorf("no solution found"),
		Completed:  true,
	}
}

// copyBots creates a copy of the bot positions map.
func copyBots(bots map[model.BotId]model.Position) map[model.BotId]model.Position {
	result := make(map[model.BotId]model.Position, len(bots))
	for id, pos := range bots {
		result[id] = pos
	}
	return result
}

// isWin checks if the target bot is at the target position.
func isWin(bots map[model.BotId]model.Position, target model.BotPosition) bool {
	pos, ok := bots[target.Id]
	if !ok {
		return false
	}
	return pos == target.Pos
}

// computeDestination calculates where a bot ends up when sliding.
func computeDestination(bots map[model.BotId]model.Position, walls *wallLookup, botId model.BotId, dir model.Direction) (model.Position, error) {
	pos, ok := bots[botId]
	if !ok {
		return model.Position{}, fmt.Errorf("bot not found")
	}

	var dx, dy model.BoardDim
	switch dir {
	case model.Up:
		dy = -1
	case model.Down:
		dy = 1
	case model.Left:
		dx = -1
	case model.Right:
		dx = 1
	}

	// Slide until hitting an obstacle
	for {
		// Check for wall blocking movement
		if walls.hasWallBlocking(pos, dir) {
			break
		}

		nextPos := model.Position{X: pos.X + dx, Y: pos.Y + dy}

		// Check for other bots
		blocked := false
		for otherId, otherPos := range bots {
			if otherId != botId && otherPos == nextPos {
				blocked = true
				break
			}
		}
		if blocked {
			break
		}

		pos = nextPos
	}

	return pos, nil
}

func init() {
	solver.Register(&AStarSolver{})
}
