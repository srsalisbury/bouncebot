package bfs2

import (
	"context"
	"fmt"
	"sort"

	"github.com/srsalisbury/bouncebot/model"
	"github.com/srsalisbury/bouncebot/solver"
)

// BFSSolver implements a breadth-first search solver.
// It finds the optimal (shortest) solution.
type BFSSolver struct{}

// Name returns the solver name.
func (s *BFSSolver) Name() string {
	return "bfs2"
}

// state represents the game state as a hashable key.
type state struct {
	// Bot positions encoded as a string for hashing.
	key string
}

// stateNode is a node in the BFS search tree.
type stateNode struct {
	bots  map[model.BotId]model.Position
	moves []model.BotPosition
}

// encodeState creates a hashable key from bot positions.
func encodeState(bots map[model.BotId]model.Position) string {
	// Sort bot IDs for consistent ordering
	ids := make([]model.BotId, 0, len(bots))
	for id := range bots {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	// Encode as string
	result := ""
	for _, id := range ids {
		pos := bots[id]
		result += fmt.Sprintf("%d:%d,%d;", id, pos.X, pos.Y)
	}
	return result
}

// Solve finds the optimal solution using breadth-first search.
func (s *BFSSolver) Solve(ctx context.Context, game *model.Game) solver.Result {
	directions := []model.Direction{model.Up, model.Down, model.Left, model.Right}

	// Initial state
	initialNode := &stateNode{
		bots:  copyBots(game.Bots),
		moves: nil,
	}

	// Check if already solved
	if isWin(initialNode.bots, game.Target) {
		return solver.Result{
			SolverName: s.Name(),
			Solution:   &solver.Solution{Moves: nil},
			Completed:  true,
		}
	}

	// BFS queue and visited set
	queue := []*stateNode{initialNode}
	visited := make(map[string]bool)
	visited[encodeState(initialNode.bots)] = true

	checkCount := 0
	const checkInterval = 1000 // Check context every N iterations

	for len(queue) > 0 {
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
		current := queue[0]
		queue = queue[1:]

		// Try all moves: each bot in each direction
		for botId := range current.bots {
			for _, dir := range directions {
				// Compute destination
				dest, err := computeDestination(current.bots, game, botId, dir)
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

				// Check if visited
				stateKey := encodeState(newBots)
				if visited[stateKey] {
					continue
				}
				visited[stateKey] = true

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
				queue = append(queue, &stateNode{
					bots:  newBots,
					moves: newMoves,
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
func computeDestination(bots map[model.BotId]model.Position, game *model.Game, botId model.BotId, dir model.Direction) (model.Position, error) {
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
		if hasWallBlocking(game, pos, dir) {
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

// hasWallBlocking checks if there's a wall or board edge blocking movement.
func hasWallBlocking(game *model.Game, pos model.Position, dir model.Direction) bool {
	switch dir {
	case model.Up:
		if pos.Y == 0 {
			return true
		}
		for _, w := range game.Board.HWalls() {
			if w.X == pos.X && w.Y == pos.Y-1 {
				return true
			}
		}
	case model.Down:
		if pos.Y == game.Board.Size()-1 {
			return true
		}
		for _, w := range game.Board.HWalls() {
			if w.X == pos.X && w.Y == pos.Y {
				return true
			}
		}
	case model.Left:
		if pos.X == 0 {
			return true
		}
		for _, w := range game.Board.VWalls() {
			if w.X == pos.X-1 && w.Y == pos.Y {
				return true
			}
		}
	case model.Right:
		if pos.X == game.Board.Size()-1 {
			return true
		}
		for _, w := range game.Board.VWalls() {
			if w.X == pos.X && w.Y == pos.Y {
				return true
			}
		}
	}
	return false
}

func init() {
	solver.Register(&BFSSolver{})
}
