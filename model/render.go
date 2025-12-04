package model

import (
	"fmt"
	"strings"

	"github.com/lithammer/dedent"
)

// Rendering and parsing games, boards, and panels as strings.

// Renders the game as a string.
// bots and target may be nil to omit them.
// Example output:
// +----+----+----+
// | B1           |
// +    +    +    +
// |    | T0      |
// +    +----+    +
// |           B0 |
// +----+----+----+
func renderGame(board *Board, bots map[BotId]Position, target *BotPosition) string {
	renderHWall := func(x, y BoardDim) string {
		if board.HasHWallAt(Position{x, y}) {
			return "----"
		}
		return "    "
	}
	// Render a row with horizontal walls
	renderHwallRow := func(y BoardDim) string {
		var rowstr strings.Builder
		rowstr.WriteString("+")
		for x := range board.Size {
			rowstr.WriteString(renderHWall(x, y))
			rowstr.WriteString("+")
		}
		return rowstr.String()
	}
	renderVWall := func(x, y BoardDim) string {
		if board.HasVWallAt(Position{x, y}) {
			return "|"
		}
		return " "
	}
	renderCell := func(x, y BoardDim) string {
		cellPos := Position{x, y}
		hasBot, botId := hasBotAtPosition(bots, cellPos)
		if hasBot {
			return fmt.Sprintf(" B%v ", botId)
		}
		if target != nil && target.Pos == cellPos {
			return fmt.Sprintf(" T%v ", target.Id)
		}
		return "    "
	}
	// Render a row with vertical walls and cell contents
	renderVwallRow := func(y BoardDim) string {
		var rowstr strings.Builder
		// Leftmost VWall
		rowstr.WriteString(renderVWall(-1, y))
		// Iterate over vertical walls and cell contents
		for x := range board.Size {
			rowstr.WriteString(renderCell(x, y))
			rowstr.WriteString(renderVWall(x, y))
		}
		return rowstr.String()
	}

	var boardstr strings.Builder
	// Top HWall border
	boardstr.WriteString(renderHwallRow(-1))
	for y := range board.Size {
		boardstr.WriteString("\n")
		boardstr.WriteString(renderVwallRow(y))
		boardstr.WriteString("\n")
		boardstr.WriteString(renderHwallRow(y))
	}

	return boardstr.String()
}

func renderBoard(b *Board) string {
	return renderGame(b, nil, nil)
}

// Parse a board from a string representation.
/* e.g. ParseBoardString(`
     +----+----+----+
     |              |
     +    +    +    +
     | B2 | T1      |
     +    +----+    +
     | B1        B0 |
     +----+----+----+
   `)
*/
func ParseBoardString(bs string) (*Board, error) {
	return ParseGenericBoardString(bs, false)
}
func ParsePanelString(bs string) (*Board, error) {
	return ParseGenericBoardString(bs, true)
}
func ParseGenericBoardString(bs string, isPanel bool) (*Board, error) {
	bs = dedentBoardString(bs)
	lines := strings.Split(bs, "\n")
	size := BoardDim((len(lines) - 1) / 2)
	scanSize := size
	if isPanel {
		// For panels, we need to check for explicit walls on the right and bottom edges.
		scanSize++
	}

	// Check that board is square
	expectedLineLength := int(size)*5 + 1
	for i, line := range lines {
		if len(line) != expectedLineLength {
			return nil, fmt.Errorf("line %d length %d does not match expected %d for size %d", i, len(line), expectedLineLength, size)
		}
	}
	// Populate hWalls
	var hWalls []Position
	for y := range scanSize - 1 {
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
		for x := range scanSize - 1 {
			charIdx := int(x+1) * 5
			if line[charIdx:charIdx+1] == "|" {
				vWalls = append(vWalls, Position{x, y})
			}
		}
	}
	if isPanel {
		return NewPanel(size, vWalls, hWalls), nil
	}
	return NewBoard(size, vWalls, hWalls), nil
}

// MustParseBoardString is like ParseBoardString but panics on error.
func MustParseBoardString(bs string) *Board {
	board, err := ParseBoardString(bs)
	if err != nil {
		panic(err)
	}
	return board
}

// MustParsePanelString is like ParsePanelString but panics on error.
func MustParsePanelString(bs string) *Board {
	board, err := ParsePanelString(bs)
	if err != nil {
		panic(err)
	}
	return board
}

// Parse a game from a string representation.
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
	bs = dedentBoardString(bs)
	lines := strings.Split(bs, "\n")
	size := BoardDim((len(lines) - 1) / 2)

	board, err := ParseBoardString(bs)
	if err != nil {
		return nil, err
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
					return nil, fmt.Errorf("unable to parse bot ID: %v", err)
				}
				if _, exists := botPositions[botId]; exists {
					return nil, fmt.Errorf("duplicate bot ID found: %d", botId)
				}
				botPositions[botId] = Position{x, y}
			} else if strings.HasPrefix(cellContent, "T") {
				var botId BotId
				_, err := fmt.Sscanf(cellContent, "T%d", &botId)
				if err != nil {
					return nil, fmt.Errorf("unable to parse target bot ID: %v", err)
				}
				botTarget = BotPosition{botId, Position{x, y}}
			}
		}
	}
	if botTarget.Id == -1 {
		return nil, fmt.Errorf("no target bot found in game string")
	}
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

// Dedent and remove leading/trailing blank lines for easier parsing and test comparison.
func dedentBoardString(s string) string {
	return strings.TrimSpace(dedent.Dedent(s))
}
