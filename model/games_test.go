package model

import (
	"slices"
	"testing"
)

func TestNewRandomGame(t *testing.T) {
	// Run multiple times to test randomness
	for i := 0; i < 10; i++ {
		game := NewRandomGame()

		// Board should be 16x16 (two 8x8 panels combined)
		if game.Board.Size() != 16 {
			t.Errorf("Expected board size 16, got %d", game.Board.Size())
		}

		// Should have 4 robots
		if len(game.Bots) != 4 {
			t.Errorf("Expected 4 bots, got %d", len(game.Bots))
		}

		// Target should be at a valid possible target location
		possibleTargets := game.Board.PossibleTargets()
		if !slices.Contains(possibleTargets, game.Target.Pos) {
			t.Errorf("Target position %v is not a valid possible target", game.Target.Pos)
		}

		// Center cells that should be avoided
		centerCells := []Position{
			{X: 7, Y: 7},
			{X: 8, Y: 7},
			{X: 7, Y: 8},
			{X: 8, Y: 8},
		}

		// Verify robots don't overlap with each other
		positions := make([]Position, 0, 4)
		for _, pos := range game.Bots {
			positions = append(positions, pos)
		}
		for i := 0; i < len(positions); i++ {
			for j := i + 1; j < len(positions); j++ {
				if positions[i] == positions[j] {
					t.Errorf("Robots overlap at position %v", positions[i])
				}
			}
		}

		// Verify no robot is on the target
		for botId, pos := range game.Bots {
			if pos == game.Target.Pos {
				t.Errorf("Bot %d is on target position %v", botId, pos)
			}
		}

		// Verify no robot is in center cells
		for botId, pos := range game.Bots {
			if slices.Contains(centerCells, pos) {
				t.Errorf("Bot %d is in center cell %v", botId, pos)
			}
		}

		// Target bot ID should be 0-3
		if game.Target.Id < 0 || game.Target.Id > 3 {
			t.Errorf("Target bot ID %d is out of range", game.Target.Id)
		}
	}
}

func TestBuildBoardFromPanels(t *testing.T) {
	tests := []struct {
		name     string
		panelStr string
		wantStr  string
	}{
		{
			"Case 1",
			`
			+----+----+----+
			|    |         |
			+    +    +    +
			|              |
			+    +----+    +
			|              |
			+----+----+----+
			`,
			`
			+----+----+----+----+----+----+
			|    |                        |
			+    +    +    +    +    +----+
			|                   |         |
			+    +----+    +    +    +    +
			|                             |
			+    +    +    +    +    +    +
			|                             |
			+    +    +    +    +----+    +
			|         |                   |
			+----+    +    +    +    +    +
			|                        |    |
			+----+----+----+----+----+----+
			`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			panel, err := ParseBoardString(tc.panelStr)
			if err != nil {
				t.Errorf("Unexpected error parsing panel string: %v", err)
			}
			wantBoard, err := ParseBoardString(tc.wantStr)
			if err != nil {
				t.Errorf("Unexpected error parsing board string: %v", err)
			}
			gotBoard := BuildBoardFromPanels(panel, panel, panel, panel)
			// Compare string forms because it normalizes wall order.
			if gotBoard.String() != wantBoard.String() {
				t.Errorf("BoardFromPanels()\nGot:\n%+v\nWant:\n%+v", gotBoard, wantBoard)
			}
		})
	}
}
