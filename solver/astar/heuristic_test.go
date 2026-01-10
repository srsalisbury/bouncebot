package astar

import (
	"fmt"
	"os"
	"testing"

	"github.com/srsalisbury/bouncebot/model"
)

func TestCalculateBaseHeuristic_Visual(t *testing.T) {
	game := model.Game1()
	ht := NewHeuristicTable(game)

	size := int(game.Board.Size())
	targetX := int(game.Target.Pos.X)
	targetY := int(game.Target.Pos.Y)

	// Build wall lookup maps for quick access
	vWalls := make(map[string]bool)
	for _, w := range game.Board.VWalls() {
		vWalls[fmt.Sprintf("%d,%d", w.X, w.Y)] = true
	}
	hWalls := make(map[string]bool)
	for _, w := range game.Board.HWalls() {
		hWalls[fmt.Sprintf("%d,%d", w.X, w.Y)] = true
	}

	hasVWall := func(x, y int) bool {
		return vWalls[fmt.Sprintf("%d,%d", x, y)]
	}
	hasHWall := func(x, y int) bool {
		return hWalls[fmt.Sprintf("%d,%d", x, y)]
	}

	// Build visual output with walls
	var output string
	output += fmt.Sprintf("Board size: %d x %d\n", size, size)
	output += fmt.Sprintf("Target: bot %d at (%d, %d)\n\n", game.Target.Id, targetX, targetY)

	// Top border
	output += "   +"
	for x := 0; x < size; x++ {
		output += "---+"
	}
	output += "\n"

	// Each row with walls
	for y := 0; y < size; y++ {
		// Row with cell values and vertical walls
		output += fmt.Sprintf("%2d |", y)
		for x := 0; x < size; x++ {
			idx := y*size + x
			d := ht.table[idx]

			// Cell value
			if x == targetX && y == targetY {
				output += " * " // Mark target with *
			} else if d == -1 {
				output += " . " // Unreachable
			} else {
				output += fmt.Sprintf("%2d ", d)
			}

			// Right edge: vertical wall or space
			if x == size-1 {
				output += "|" // Board edge
			} else if hasVWall(x, y) {
				output += "|"
			} else {
				output += " "
			}
		}
		output += "\n"

		// Row with horizontal walls
		output += "   +"
		for x := 0; x < size; x++ {
			if y == size-1 {
				output += "---+" // Bottom board edge
			} else if hasHWall(x, y) {
				output += "---+"
			} else {
				output += "   +"
			}
		}
		output += "\n"
	}

	// Summary stats
	output += "\nLegend:\n"
	output += "  * = target position (distance 0)\n"
	output += "  . = unreachable from target\n"
	output += "  N = minimum rook-moves to reach target\n"
	output += "  | = vertical wall\n"
	output += "  --- = horizontal wall\n"

	// Count reachable/unreachable
	reachable := 0
	unreachable := 0
	maxDist := 0
	for _, d := range ht.table {
		if d == -1 {
			unreachable++
		} else {
			reachable++
			if d > maxDist {
				maxDist = d
			}
		}
	}
	output += fmt.Sprintf("\nStats: %d reachable, %d unreachable, max distance = %d\n", reachable, unreachable, maxDist)

	// Write to file
	filename := "heuristic_visual.txt"
	err := os.WriteFile(filename, []byte(output), 0644)
	if err != nil {
		t.Fatalf("Failed to write output file: %v", err)
	}

	t.Logf("Wrote heuristic visualization to %s", filename)
	t.Log("\n" + output)
}
