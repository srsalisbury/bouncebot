package bouncebot

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

type BotPosition struct {
	Id  int8
	Pos Position
}

// A full game state, including board, starting bot positions, and target bot position.
type Game struct {
	B         *Board
	BotsStart map[int8]Position
	// Where the given bot needs to end up.
	BotTarget BotPosition
}

func NewGame(b *Board, botsStart map[int8]Position, botTarget BotPosition) (*Game, error) {
	err := b.IsValid()
	if err != nil {
		return nil, err
	}
	// Validate that botTarget.Id exists in botsStart
	if _, ok := botsStart[botTarget.Id]; !ok {
		return nil, fmt.Errorf("botTarget.Id %d not found in botsStart", botTarget.Id)
	}
	err = b.ValidateBotWithin(botTarget.Pos)
	if err != nil {
		return nil, fmt.Errorf("botTarget %v", err)
	}
	for id, pos := range botsStart {
		err = b.ValidateBotWithin(pos)
		if err != nil {
			return nil, fmt.Errorf("botsStart for bot %d %v", id, err)
		}
	}
	// Validate that no two bots start in the same position
	positionsSeen := make(map[Position]bool)
	for _, pos := range botsStart {
		if positionsSeen[pos] {
			return nil, fmt.Errorf("multiple bots starting in the same position (%d, %d)", pos.X, pos.Y)
		}
		positionsSeen[pos] = true
	}
	return &Game{
		B:         b,
		BotsStart: botsStart,
		BotTarget: botTarget,
	}, nil
}

// ValidateMove checks if a bot's intended move is valid based on the game rules.
// The inputs are the board state, the starting positions of all bots, and a bot's intended end position.
// TODO: Do we need to describe why the move isn't valid?
func (g *Game) ValidateMove(botEndId int8, botEndPos Position) bool {
	botStartPos, ok := g.BotsStart[botEndId]
	if !ok {
		return false
		// 	return fmt.Errorf("bot with id %d not found", botEndId)
	}

	isStraightLine := func(pos1, pos2 Position) bool {
		return pos1.X == pos2.X || pos1.Y == pos2.Y
	}

	if !g.B.IsBotWithin(botEndPos) ||
		botEndPos == botStartPos ||
		!isStraightLine(botEndPos, botStartPos) {
		return false
	}

	// Determine direction of movement from a to b (+1 if a < b, -1 if a >= b)
	dir := func(a, b int8) int8 {
		if a < b {
			return 1
		}
		return -1
	}

	// Extract a coordinate from a Position
	type getCoordFunc func(Position) int8

	// Check path for obstacles and end position validity for one axis
	// (Given start and end positions, walls, and motion/axis coordinate getters)
	// Returns true if path is clear and end position is valid.
	checkPathAlongAxis := func(start, end Position, walls []Position, getMotionCoord, getAxisCoord getCoordFunc) bool {
		minCoord, maxCoord := getMotionCoord(start), getMotionCoord(end)
		direction := dir(minCoord, maxCoord)

		// Get correct min and max coordinates for bounds checking of walls and bots in the way
		if minCoord > maxCoord {
			minCoord, maxCoord = maxCoord, minCoord
		}

		// Check path for walls or other bots
		for _, wall := range walls {
			wallCoord := getMotionCoord(wall)
			// Check if wall in correct axis and in middle of path
			// Note: wall at maxCoord is allowed since it only blocks further movement
			if getAxisCoord(start) == getAxisCoord(wall) && wallCoord >= minCoord && wallCoord < maxCoord {
				return false
			}
		}
		for otherBotId, otherBotPos := range g.BotsStart {
			if otherBotId != botEndId { // only check other bots
				botCoord := getMotionCoord(otherBotPos)
				// Check if otherBot in correct axis and in middle of path
				if getAxisCoord(start) == getAxisCoord(otherBotPos) && botCoord >= minCoord && botCoord <= maxCoord {
					return false
				}
			}
		}

		if direction == 1 { // Moving towards increasing coordinate (e.g., right or down)

			// At board edge
			if getMotionCoord(end) == g.B.Size-1 {
				return true
			}
			// Wall just beyond end position
			if slices.ContainsFunc(walls, func(p Position) bool {
				return getMotionCoord(p) == getMotionCoord(end) && getAxisCoord(p) == getAxisCoord(end)
			}) {
				return true
			}

			// Bot just beyond end position
			if slices.ContainsFunc(slices.Collect(maps.Values(g.BotsStart)), func(p Position) bool {
				return getMotionCoord(p)-1 == getMotionCoord(end) && getAxisCoord(p) == getAxisCoord(end)
			}) {
				return true
			}

		} else { // Moving towards decreasing coordinate (e.g., left or up)

			// At board edge
			if getMotionCoord(end) == 0 {
				return true
			}

			// Wall just beyond end position
			if slices.ContainsFunc(walls, func(p Position) bool {
				return getMotionCoord(p)+1 == getMotionCoord(end) && getAxisCoord(p) == getAxisCoord(end)
			}) {
				return true
			}

			// Bot just beyond end position
			if slices.ContainsFunc(slices.Collect(maps.Values(g.BotsStart)), func(p Position) bool {
				return getMotionCoord(p)+1 == getMotionCoord(end) && getAxisCoord(p) == getAxisCoord(end)
			}) {
				return true
			}
		}

		return false
	}

	getX := func(p Position) int8 { return p.X }
	getY := func(p Position) int8 { return p.Y }

	// Check path for obstacles and end position validity
	if botEndPos.X == botStartPos.X {
		// Vertical move
		return checkPathAlongAxis(botStartPos, botEndPos, g.B.HWallPos, getY, getX)
	} else {
		// Horizontal move
		return checkPathAlongAxis(botStartPos, botEndPos, g.B.VWallPos, getX, getY)
	}
}

// Returns a new Game with the given bot moved to the given position,
// or an error if the move is invalid.
func (g *Game) MoveBot(id int8, pos Position) (*Game, error) {
	if !g.ValidateMove(id, pos) {
		return nil, fmt.Errorf("invalid move for bot %d to position (%d, %d)", id, pos.X, pos.Y)
	}

	// Create new BotsStart map with updated position for the moved bot
	newBotsStart := make(map[int8]Position)
	for botId, botPos := range g.BotsStart {
		if botId == id {
			newBotsStart[botId] = pos
		} else {
			newBotsStart[botId] = botPos
		}
	}

	return NewGame(g.B, newBotsStart, g.BotTarget)
}

func (g *Game) IsWin() bool {
	targetPos, ok := g.BotsStart[g.BotTarget.Id]
	if !ok {
		return false
	}
	return targetPos == g.BotTarget.Pos
}

func (g *Game) Equals(o *Game) bool {
	return reflect.DeepEqual(g, o)
}

func (g *Game) String() string {
	return Render(g.B, g.BotTarget, g.BotsStart)
}

// Renders the board as a string.
// Example output:
// +----+----+----+
// | B1           |
// +    +    +    +
// |    | T0      |
// +    +----+    +
// |           B0 |
// +----+----+----+
func Render(b *Board, botTarget BotPosition, botsStart map[int8]Position) string {
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
