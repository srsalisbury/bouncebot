package model

import (
	"testing"
)

func TestRenderBoard(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	board := NewBoard(3, vW, hW)
	want := dedentBoardString(`
		+----+----+----+
		|              |
		+    +    +    +
		|    |         |
		+    +----+    +
		|              |
		+----+----+----+
		`)
	got := renderBoard(board)
	if got != want {
		t.Errorf(`String()
got:
%v

want:
%v`, got, want)
	}
}

func TestRenderPanel(t *testing.T) {
	vW := []Position{{0, 1}, {2, 2}}
	hW := []Position{{1, 1}, {2, 2}}
	board := NewPanel(3, vW, hW)
	want := dedentBoardString(`
		+----+----+----+
		|               
		+    +    +    +
		|    |          
		+    +----+    +
		|              |
		+    +    +----+
		`)
	got := renderBoard(board)
	if got != want {
		t.Errorf(`String()
got:
%v

want:
%v`, got, want)
	}
}

func TestRenderGame(t *testing.T) {
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
	got := renderGame(board, bots, &goal)
	if got != want {
		t.Errorf(`renderGame()
got:
%v

want:
%v`, got, want)
	}
}

func TestParseBoardString(t *testing.T) {
	tests := []struct {
		name     string
		isPanel  bool
		isValid  bool
		boardStr string
	}{
		{
			"Valid Board - Size 3", false, true,
			`
			+----+----+----+
			|              |
			+    +    +    +
			|    |         |
			+    +----+    +
			|              |
			+----+----+----+
			`,
		},
		{
			"Valid Panel - Size 3", true, true,
			`
			+----+----+----+
			|               
			+    +    +    +
			|    |          
			+    +----+    +
			|              |
			+    +    +----+
			`,
		},
		{
			"Valid Board - Size 2", false, true,
			`
			+----+----+
			|         |
			+----+    +
			|         |
			+----+----+
			`,
		},
		{
			"Valid Board - Size 4", false, true,
			`
			+----+----+----+----+
			|                   |
			+    +----+    +    +
			|    |         |    |
			+    +    +----+    +
			|    |              |
			+    +----+----+    +
			|                   |
			+----+----+----+----+
			`,
		},
		{
			"Invalid Board - not square", false, false,
			`
			+----+----+
			|         |
			+    +    +
			|         |
			+----+    +
			|         |
			+----+----+
			`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var board *Board
			var err error
			if tc.isPanel {
				board, err = ParsePanelString(tc.boardStr)
			} else {
				board, err = ParseBoardString(tc.boardStr)
			}
			if !tc.isValid {
				if err == nil {
					t.Errorf("Expected error parsing invalid board string, got nil")
				}
				return
			}
			// tc.isValid
			if err != nil {
				t.Errorf("Unexpected error parsing valid board string: %v", err)
			}
			gotBoardStr := board.String()
			wantBoardStr := dedentBoardString(tc.boardStr)
			// Assume that if we can re-serialize to the same string, parsing was correct.
			if gotBoardStr != wantBoardStr {
				t.Errorf("Reserialized board string mismatch:\ngot:\n%v\nwant:\n%v", gotBoardStr, wantBoardStr)
			}
		})
	}
}

func TestParseGameString(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
		gameStr string
	}{
		{
			"Valid Game - Size 3", true,
			`
			+----+----+----+
			|              |
			+    +    +    +
			| B2 | T0      |
			+    +----+    +
			| B1        B0 |
			+----+----+----+
			`,
		},
		{
			"Valid Game - Size 2", true,
			`
			+----+----+
			| T0      |
			+----+    +
			|      B0 |
			+----+----+
			`,
		},
		{
			"Valid Game - Size 4", true,
			`
			+----+----+----+----+
			|           B1      |
			+    +----+    +    +
			|    |         |    |
			+    +    +----+    +
			|    | B0        T2 |
			+    +----+----+    +
			| B2                |
			+----+----+----+----+
			`,
		},
		{
			"Invalid Game - no bot for target", false,
			`
			+----+----+----+
			|              |
			+    +    +    +
			| B2 | T4      |
			+    +----+    +
			| B1        B0 |
			+----+----+----+
			`,
		},
		{
			"Invalid Game - missing target", false,
			`
			+----+----+----+
			|              |
			+    +    +    +
			| B2 |         |
			+    +----+    +
			| B1        B0 |
			+----+----+----+
			`,
		},
		{
			"Invalid Game - duplicate bot id", false,
			`
			+----+----+----+
			|              |
			+    +    +    +
			| B2 | T1      |
			+    +----+    +
			| B1        B2 |
			+----+----+----+
			`,
		},
		{
			"Invalid Game - not square", false,
			`
			+----+----+
			|         |
			+    +    +
			| T1      |
			+----+    +
			|      B1 |
			+----+----+
			`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			game, err := ParseGameString(tc.gameStr)
			if !tc.isValid {
				if err == nil {
					t.Errorf("Expected error parsing invalid game string, got nil")
				}
				return
			}
			// tc.isValid
			if err != nil {
				t.Errorf("Unexpected error parsing valid game string: %v", err)
			}
			gotGameStr := game.String()
			wantGameStr := dedentBoardString(tc.gameStr)
			// Assume that if we can re-serialize to the same string, parsing was correct.
			if gotGameStr != wantGameStr {
				t.Errorf("Reserialized game string mismatch:\ngot:\n%v\nwant:\n%v", gotGameStr, wantGameStr)
			}
		})
	}
}
