package astar

import (
	"github.com/srsalisbury/bouncebot/model"
)

// HeuristicTable stores precomputed minimum rook-move distances from each cell to the target.
type HeuristicTable struct {
	table [256]int
	size  int
}

// Get retrieves the heuristic value for a given position.
func (ht *HeuristicTable) Get(pos model.Position) int {
	return ht.table[int(pos.Y)*ht.size+int(pos.X)]
}

// NewHeuristicTable computes the base heuristic distances from each cell to the target
// using reverse BFS with rook movement.
func NewHeuristicTable(game *model.Game) *HeuristicTable {
	directions := []model.Direction{model.Up, model.Down, model.Left, model.Right}
	size := int(game.Board.Size())
	var dist [256]int
	for i := range dist {
		dist[i] = -1 // -1 means unseen
	}

	targetIdx := int(game.Target.Pos.Y)*size + int(game.Target.Pos.X)
	dist[targetIdx] = 0

	queue := []int{targetIdx}

	for len(queue) > 0 {
		currentIdx := queue[0]
		queue = queue[1:]
		currentX := model.BoardDim(currentIdx % size)
		currentY := model.BoardDim(currentIdx / size)
		currentPos := model.Position{X: currentX, Y: currentY}
		currentDist := dist[currentIdx]

		for _, dir := range directions {
			// Use computeDestination with a single fake bot to find where a slide would end
			endPos, err := computeDestination(map[model.BotId]model.Position{game.Target.Id: currentPos}, game, game.Target.Id, dir)
			if err != nil || endPos == currentPos {
				continue
			}
			queue = markPathCells(currentPos, endPos, currentDist, size, &dist, queue)
		}
	}
	return &HeuristicTable{table: dist, size: size}
}

// markPathCells marks all cells along the path from start to end (exclusive of start)
// with distance newDist, and adds newly marked cells to the queue.
func markPathCells(start, end model.Position, currentDist, size int, dist *[256]int, queue []int) []int {
	var stepX, stepY model.BoardDim
	if end.X > start.X {
		stepX = 1
	} else if end.X < start.X {
		stepX = -1
	}
	if end.Y > start.Y {
		stepY = 1
	} else if end.Y < start.Y {
		stepY = -1
	}

	pos := start
	for pos != end {
		pos = model.Position{X: pos.X + stepX, Y: pos.Y + stepY}
		idx := int(pos.Y)*size + int(pos.X)
		if dist[idx] == -1 || dist[idx] > currentDist+1 {
			dist[idx] = currentDist + 1
			queue = append(queue, idx)
		}
	}
	return queue
}
