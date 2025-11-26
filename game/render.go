package bouncebot

import (
	"fmt"
	"slices"
	"strings"
)

/*
+ -- + -- + -- +
| B1           |
+    +    +    +
|    | T0      |
+    + -- +    +
|           B0 |
+ -- + -- + -- +
*/
func Render(b Board, g BotPosition, botStart []BotPosition) string {
	var toReturn strings.Builder
	size := int(b.Size)

	// Iterate for every row
	for r := 0; r < int(size); r++ {

		// Iterate for each set of horizontal walls
		for c := 0; c < int(size); c++ {
			toReturn.WriteString("+ ")
			if r == 0 || slices.Contains(b.HWallPos, Position{int8(c), int8(r - 1)}) {
				toReturn.WriteString("-- ")
			} else {
				toReturn.WriteString("   ")
			}
		}
		toReturn.WriteString("+\n")

		// Iterate over vertical walls/cell contents
		for c := 0; c < int(size); c++ {

			// Check for vWall
			if c == 0 || slices.Contains(b.VWallPos, Position{int8(c - 1), int8(r)}) {
				toReturn.WriteString("|")
			} else {
				toReturn.WriteString(" ")
			}

			// Check for bot/target
			index := slices.IndexFunc(botStart, func(bp BotPosition) bool {
				return bp.Pos.X == int8(c) && bp.Pos.Y == int8(r)
			})
			if index != -1 {
				toReturn.WriteString(fmt.Sprintf(" B%v ", botStart[index].Id))
			} else if g.Pos.X == int8(c) && g.Pos.Y == int8(r) {
				toReturn.WriteString(fmt.Sprintf(" T%v ", g.Id))
			} else {
				toReturn.WriteString("    ")
			}
		}
		toReturn.WriteString("|\n")
	}

	// Print bottom edge
	for c := 0; c < int(size); c++ {
		toReturn.WriteString("+ -- ")
	}
	toReturn.WriteString("+")
	return toReturn.String()
}
