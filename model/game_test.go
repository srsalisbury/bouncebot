package model

import (
	"reflect"
	"testing"
)

func TestNewGame_Valid(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	bots := map[BotId]Position{
		0: {2, 2},
		1: {0, 0}}
	target := BotPosition{0, Position{1, 1}}
	board := NewBoard(3, vW, hW)

	game, err := NewGame(board, bots, target)
	if err != nil {
		t.Fatalf("Failed to create new game: %v", err)
	}

	if game.Board.Size() != 3 {
		t.Errorf("Expected board size 3, got %d", game.Board.Size())
	}
	if len(game.Board.VWalls()) != 1 || len(game.Board.HWalls()) != 1 {
		t.Errorf("Unexpected wall positions in board")
	}
	if !reflect.DeepEqual(game.Bots, bots) {
		t.Errorf("Bots mismatch: got %v, want %v", game.Bots, bots)
	}
	if game.Target != target {
		t.Errorf("BotTarget mismatch: got %v, want %v", game.Target, target)
	}
}

func TestNewGame_InvalidBotTarget(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	bots := map[BotId]Position{
		0: {2, 2},
		1: {0, 0}}
	goal := BotPosition{2, Position{3, 3}} // Invalid target position (out of bounds)
	board := NewBoard(3, vW, hW)

	_, err := NewGame(board, bots, goal)
	if err == nil {
		t.Errorf("Expected error for invalid bot target position, got nil")
	}
}

func TestGame_Equals(t *testing.T) {
	game1 := MustParseGameString(`
		+----+----+----+
		|              |
		+    +    +    +
		| B2 | T0      |
		+    +----+    +
		| B1        B0 |
		+----+----+----+
	`)
	game2 := MustParseGameString(`
		+----+----+----+
		|              |
		+    +    +    +
		| B2 | T0      |
		+    +----+    +
		| B1        B0 |
		+----+----+----+
	`)
	game3 := MustParseGameString(`
		+----+----+----+
		|              |
		+    +    +    +
		|    | T0   B2 |
		+    +----+    +
		| B1        B0 |
		+----+----+----+
	`)

	if !game1.Equals(game1) {
		t.Errorf("Expected game1 to equal game1")
	}

	if !game1.Equals(game2) {
		t.Errorf("Expected game1 to equal game2")
	}

	if game1.Equals(game3) {
		t.Errorf("Expected game1 to not equal game3")
	}
}

func TestGame_IsWin(t *testing.T) {
	game := MustParseGameString(`
		+----+----+----+
		|              |
		+    +    +    +
		| B2 | T0   B0 |
		+    +----+    +
		| B1           |
		+----+----+----+
	`)

	if game.IsWin() {
		t.Errorf("Expected IsWin to be false initially")
	}

	// Move bot 0 to target position
	gameMoved, err := game.MoveBot(0, Position{1, 1})
	if err != nil {
		t.Fatalf("Failed to move bot: %v", err)
	}

	if !gameMoved.IsWin() {
		t.Errorf("Expected IsWin to be true after moving bot to target position")
	}
}

func TestValidate(t *testing.T) {
	game := MustParseGameString(`
		+----+----+----+
		|              |
		+    +    +    +
		| B2 | T0      |
		+    +----+    +
		| B1        B0 |
		+----+----+----+
	`)
	var tests = []struct {
		name      string
		botEndId  BotId
		botEndPos Position // Bot's intended end position
		wantErr   bool
	}{
		{"Valid - bot 0 upwards to border", 0, Position{2, 0}, false},
		{"Valid - left to bot 1", 0, Position{1, 2}, false},
		{"Valid - bot 2 upwards to border", 2, Position{0, 0}, false},
		{"Valid - right to bot 0", 1, Position{1, 2}, false},
		{"Invalid - same position", 0, Position{2, 2}, true},
		{"Invalid - out of bounds", 0, Position{-1, 2}, true},
		{"Invalid - through bot 1", 0, Position{0, 2}, true},
		{"Invalid - diagonal move", 0, Position{0, 1}, true},
		{"Invalid - through wall", 2, Position{1, 1}, true},
		{"Invalid - through bot 2", 1, Position{0, 0}, true},
		{"Invalid - not against wall/border/bot", 0, Position{2, 1}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := game.ValidateMove(tc.botEndId, tc.botEndPos)
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("%s: gotErr %v, wantErr %v", tc.name, gotErr, tc.wantErr)
			}
		})
	}
}

func TestGame_MoveBot(t *testing.T) {
	game := MustParseGameString(`
		+----+----+----+
		|              |
		+    +    +    +
		| B2 | T0      |
		+    +----+    +
		|      B1   B0 |
		+----+----+----+
	`)

	// Valid move
	gameMoved, err := game.MoveBot(0, Position{2, 0})
	if err != nil {
		t.Fatalf("Failed to move bot: %v", err)
	}
	expectedPos := Position{2, 0}
	if gameMoved.Bots[0] != expectedPos {
		t.Errorf("Expected bot 0 position to be %v, got %v", expectedPos, gameMoved.Bots[0])
	}

	// Invalid move
	_, err = game.MoveBot(0, Position{0, 2}) // Through bot 1
	if err == nil {
		t.Errorf("Expected error for invalid move, got nil")
	}
}

func TestGame_String(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	bots := map[BotId]Position{
		0: {2, 2},
		1: {0, 0}}
	goal := BotPosition{0, Position{1, 1}}
	board := NewBoard(3, vW, hW)
	want := dedentBoardString(`
		+----+----+----+
		| B1           |
		+    +    +    +
		|    | T0      |
		+    +----+    +
		|           B0 |
		+----+----+----+
		`)
	game, err := NewGame(board, bots, goal)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}
	got := game.String()
	if got != want {
		t.Errorf(`String()
got:
%v

want:
%v`, got, want)
	}
}
