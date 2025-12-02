package bouncebot

import (
	"fmt"
	"slices"
	"strings"
)

// Renders the board as a string.
// Example output:
// +----+----+----+
// | B1           |
// +    +    +    +
// |    | T0      |
// +    +----+    +
// |           B0 |
// +----+----+----+
// TODO: Move to board.go as method on Board.
func Render(b *Board, botTarget BotPosition, botsStart BotPositionMap) string {
	botsStartSlice := make([]BotPosition, 0, len(botsStart))
	for id, pos := range botsStart {
		botsStartSlice = append(botsStartSlice, BotPosition{Id: id, Pos: pos})
	}
	return render(b, botTarget, botsStartSlice)
}

// TODO: Rewrite to directly use BotPositionMap instead of converting to slice.
// Then integrate it back into Render above.
func render(b *Board, g BotPosition, botStart []BotPosition) string {
	renderHWall := func(x, y int8) string {
		if b.HasHWallAt(Position{x, y}) {
			return "----"
		}
		return "    "
	}
	// Render a row with horizontal walls
	renderHwallRow := func(y int8) string {
		var rowstr strings.Builder
		rowstr.WriteString("+")
		for x := range b.Size {
			rowstr.WriteString(renderHWall(x, y))
			rowstr.WriteString("+")
		}
		return rowstr.String()
	}
	renderVWall := func(x, y int8) string {
		if b.HasVWallAt(Position{x, y}) {
			return "|"
		}
		return " "
	}
	renderCell := func(x, y int8) string {
		// Check for bot/target
		index := slices.IndexFunc(botStart, func(bp BotPosition) bool {
			return bp.Pos.X == x && bp.Pos.Y == y
		})
		if index != -1 {
			return fmt.Sprintf(" B%v ", botStart[index].Id)
		}
		if g.Pos.X == x && g.Pos.Y == y {
			return fmt.Sprintf(" T%v ", g.Id)
		}
		return "    "
	}
	// Render a row with vertical walls and cell contents
	renderVwallRow := func(y int8) string {
		var rowstr strings.Builder
		// Leftmost VWall
		rowstr.WriteString(renderVWall(-1, y))
		// Iterate over vertical walls and cell contents
		for x := range b.Size {
			rowstr.WriteString(renderCell(x, y))
			rowstr.WriteString(renderVWall(x, y))

		}
		return rowstr.String()
	}

	var boardstr strings.Builder
	// Top HWall border
	boardstr.WriteString(renderHwallRow(-1))
	for y := range b.Size {
		boardstr.WriteString("\n")
		boardstr.WriteString(renderVwallRow(y))
		boardstr.WriteString("\n")
		boardstr.WriteString(renderHwallRow(y))
	}

	return boardstr.String()
}
