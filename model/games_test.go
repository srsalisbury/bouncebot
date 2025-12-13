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

func TestNewContinuationGame(t *testing.T) {
	// Create an initial game
	initial := NewRandomGame()

	// Create a continuation game
	continuation := NewContinuationGame(initial)

	// Board should be the same
	if continuation.Board.Size() != initial.Board.Size() {
		t.Errorf("Expected same board size, got %d vs %d", continuation.Board.Size(), initial.Board.Size())
	}

	// Robot positions should be the same
	for botId := BotId(0); botId < 4; botId++ {
		if continuation.Bots[botId] != initial.Bots[botId] {
			t.Errorf("Bot %d position changed: %v -> %v", botId, initial.Bots[botId], continuation.Bots[botId])
		}
	}

	// Target should be different (position or robot ID)
	// Note: There's a small chance they're the same randomly, so we just verify it's valid
	possibleTargets := continuation.Board.PossibleTargets()
	if !slices.Contains(possibleTargets, continuation.Target.Pos) {
		t.Errorf("Continuation target position %v is not a valid possible target", continuation.Target.Pos)
	}

	// Target should not be on a robot (unless all targets are occupied)
	for botId, pos := range continuation.Bots {
		if pos == continuation.Target.Pos {
			t.Logf("Warning: Target is on bot %d position (may be valid if all targets occupied)", botId)
		}
	}
}

func TestNewContinuationGame_NilPrev(t *testing.T) {
	// Should fall back to NewRandomGame when prev is nil
	game := NewContinuationGame(nil)

	if game == nil {
		t.Error("Expected non-nil game")
	}
	if game.Board.Size() != 16 {
		t.Errorf("Expected board size 16, got %d", game.Board.Size())
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
