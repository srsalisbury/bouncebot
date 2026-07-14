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

func TestNewContinuationGame_NoPossibleTargets(t *testing.T) {
	// Simulates a board that lost its possible-targets metadata (e.g. data
	// persisted before that field existed). Continuation must fall back to a
	// fresh random game instead of panicking.
	board := NewBoardWithTargets(16, nil, nil, nil)
	bots := map[BotId]Position{0: {X: 0, Y: 0}, 1: {X: 1, Y: 0}, 2: {X: 2, Y: 0}, 3: {X: 3, Y: 0}}
	prev := &Game{Board: board, Bots: bots, Target: BotPosition{Id: 0, Pos: Position{X: 5, Y: 5}}}

	game := NewContinuationGame(prev)

	if game == nil {
		t.Fatal("expected non-nil game")
	}
	if len(game.Board.PossibleTargets()) == 0 {
		t.Error("expected fallback random game to have possible targets")
	}
}

func TestNewContinuationGame_AfterJSONRoundTrip(t *testing.T) {
	// Regression test: a game that has been persisted to JSON and reloaded
	// (as happens on every server restart) must still work as the basis for a
	// continuation round, not panic.
	original := NewRandomGame()

	data, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	var restored Game
	if err := restored.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	continuation := NewContinuationGame(&restored)

	if continuation == nil {
		t.Fatal("expected non-nil continuation game")
	}
	for botId := BotId(0); botId < 4; botId++ {
		if continuation.Bots[botId] != original.Bots[botId] {
			t.Errorf("bot %d position changed across restart: %v -> %v", botId, original.Bots[botId], continuation.Bots[botId])
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
