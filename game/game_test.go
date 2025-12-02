package bouncebot

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/lithammer/dedent"
)

func TestNewGame_Valid(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	bots := map[BotId]Position{
		0: {2, 2},
		1: {0, 0}}
	target := BotPosition{0, Position{1, 1}}
	board := &Board{Size: 3, VWallPos: vW, HWallPos: hW}

	game, err := NewGame(board, bots, target)
	if err != nil {
		t.Fatalf("Failed to create new game: %v", err)
	}

	if game.Board.Size != 3 {
		t.Errorf("Expected board size 3, got %d", game.Board.Size)
	}
	if len(game.Board.VWallPos) != 1 || len(game.Board.HWallPos) != 1 {
		t.Errorf("Unexpected wall positions in board")
	}
	if !reflect.DeepEqual(game.Bots, bots) {
		t.Errorf("Bots mismatch: got %v, want %v", game.Bots, bots)
	}
	if game.BotTarget != target {
		t.Errorf("BotTarget mismatch: got %v, want %v", game.BotTarget, target)
	}
}

func TestNewGame_InvalidBotTarget(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	bots := map[BotId]Position{
		0: {2, 2},
		1: {0, 0}}
	goal := BotPosition{2, Position{3, 3}} // Invalid target position (out of bounds)
	board := &Board{Size: 3, VWallPos: vW, HWallPos: hW}

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

func TestGame_Render(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	bots := map[BotId]Position{
		0: {2, 2},
		1: {0, 0}}
	goal := BotPosition{0, Position{1, 1}}
	board := &Board{Size: 3, VWallPos: vW, HWallPos: hW}
	want := dedentTestString(`
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
	got := game.render()
	if got != want {
		t.Errorf(`render()
got:
%v

want:
%v`, got, want)
	}
}

// Dedent and remove leading/trailing blank lines for easier comparison in tests.
func dedentTestString(s string) string {
	return strings.TrimSpace(dedent.Dedent(s))
}

// Parse a game from a string representation for testing purposes.
/* e.g. ParseGameString(`
     +----+----+----+
     |              |
     +    +    +    +
     | B2 | T1      |
     +    +----+    +
     | B1        B0 |
     +----+----+----+
   `)
*/
func ParseGameString(bs string) (*Game, error) {
	bs = dedentTestString(bs)
	lines := strings.Split(bs, "\n")
	size := BoardDim((len(lines) - 1) / 2)

	// Check that board is square
	expectedLineLength := int(size)*5 + 1
	for i, line := range lines {
		if len(line) != expectedLineLength {
			return nil, fmt.Errorf("line %d length %d does not match expected %d for size %d", i, len(line), expectedLineLength, size)
		}
	}
	// Populate hWalls
	var hWalls []Position
	for y := BoardDim(0); y < size-1; y++ {
		lineIdx := (y + 1) * 2
		line := lines[lineIdx]
		for x := range size {
			charIdx := int(x)*5 + 2
			if line[charIdx:charIdx+2] == "--" {
				hWalls = append(hWalls, Position{x, y})
			}
		}
	}
	// Populate vWalls
	var vWalls []Position
	for y := range size {
		lineIdx := y*2 + 1
		line := lines[lineIdx]
		for x := BoardDim(0); x < size-1; x++ {
			charIdx := int(x+1) * 5
			if line[charIdx:charIdx+1] == "|" {
				vWalls = append(vWalls, Position{x, y})
			}
		}
	}
	// Populate botPositions
	botPositions := make(map[BotId]Position)
	botTarget := BotPosition{Id: -1}
	for y := range size {
		lineIdx := int(y*2) + 1
		line := lines[lineIdx]
		for x := range size {
			charIdx := int(x) * 5
			cellContent := line[charIdx+2 : charIdx+4]
			if strings.HasPrefix(cellContent, "B") {
				var botId BotId
				_, err := fmt.Sscanf(cellContent, "B%d", &botId)
				if err != nil {
					return nil, fmt.Errorf("Unable to parse bot ID: %v", err)
				}
				if _, exists := botPositions[botId]; exists {
					return nil, fmt.Errorf("Duplicate bot ID found: %d", botId)
				}
				botPositions[botId] = Position{x, y}
			} else if strings.HasPrefix(cellContent, "T") {
				var botId BotId
				_, err := fmt.Sscanf(cellContent, "T%d", &botId)
				if err != nil {
					return nil, fmt.Errorf("Unable to parse target bot ID: %v", err)
				}
				botTarget = BotPosition{botId, Position{x, y}}
			}
		}
	}
	if botTarget.Id == -1 {
		return nil, fmt.Errorf("No target bot found in game string")
	}
	board := &Board{Size: size, VWallPos: vWalls, HWallPos: hWalls}
	return NewGame(board, botPositions, botTarget)
}

// MustParseGameString is like ParseGameString but panics on error.
func MustParseGameString(bs string) *Game {
	game, err := ParseGameString(bs)
	if err != nil {
		panic(err)
	}
	return game
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
			wantGameStr := dedentTestString(tc.gameStr)
			// Assume that if we can re-serialize to the same string, parsing was correct.
			if gotGameStr != wantGameStr {
				t.Errorf("Reserialized game string mismatch:\ngot:\n%v\nwant:\n%v", gotGameStr, wantGameStr)
			}
		})
	}
}
