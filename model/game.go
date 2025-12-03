package model

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

type BotId int8
type BotPosition struct {
	Id  BotId
	Pos Position
}

// A full game state, including board, bot positions, and target bot position.
type Game struct {
	Board *Board
	Bots  map[BotId]Position
	// Where the given bot needs to end up.
	BotTarget BotPosition
}

// Creates a new Game instance, validating the inputs.
func NewGame(board *Board, bots map[BotId]Position, botTarget BotPosition) (*Game, error) {
	err := board.IsValid()
	if err != nil {
		return nil, err
	}
	// Validate that botTarget.Id exists in bots
	if _, ok := bots[botTarget.Id]; !ok {
		return nil, fmt.Errorf("botTarget.Id %d not found in bots", botTarget.Id)
	}
	err = board.ValidateBotWithin(botTarget.Pos)
	if err != nil {
		return nil, fmt.Errorf("botTarget %v", err)
	}
	for id, pos := range bots {
		err = board.ValidateBotWithin(pos)
		if err != nil {
			return nil, fmt.Errorf("bot %d %v", id, err)
		}
	}
	// Validate that no two bots start in the same position
	positionsSeen := make(map[Position]bool)
	for _, botPos := range bots {
		if positionsSeen[botPos] {
			return nil, fmt.Errorf("multiple bots starting in the same position %v", botPos)
		}
		positionsSeen[botPos] = true
	}
	return &Game{
		Board:     board,
		Bots:      bots,
		BotTarget: botTarget,
	}, nil
}

// ValidateMove checks if a bot's intended move is valid based on the game rules.
// The inputs are the board state, the starting positions of all bots, and a bot's intended end position.
func (g *Game) ValidateMove(botId BotId, botEndPos Position) error {
	botPos, ok := g.Bots[botId]
	if !ok {
		return fmt.Errorf("bot with id %d not found", botId)
	}

	if !g.Board.IsBotWithin(botEndPos) {
		return fmt.Errorf("end position %v is out of board boundaries", botEndPos)
	}
	if botEndPos == botPos {
		return fmt.Errorf("end position %v is the same as start position", botEndPos)
	}
	isStraightLine := func(pos1, pos2 Position) bool {
		return pos1.X == pos2.X || pos1.Y == pos2.Y
	}
	if !isStraightLine(botEndPos, botPos) {
		return fmt.Errorf("move from %v to %v is not in a straight line", botPos, botEndPos)
	}

	// Extract a coordinate from a Position
	type getCoordFunc func(Position) BoardDim

	// Check path for obstacles and end position validity for one axis
	// given start and end positions, walls, and motion/axis coordinate getters.
	// Returns an error if path is not clear or end position is invalid.
	// This handles the logic for either horizontal or vertical moves by extracting the relevant coordinate from the relevant positions.
	checkPathAlongAxis := func(startPos, endPos Position, walls []Position, getMotionCoord, getAxisCoord getCoordFunc) error {
		minCoord, maxCoord := getMotionCoord(startPos), getMotionCoord(endPos)
		// Is bot coordinate moving in increasing direction (right or down)?
		motionIsIncreasing := minCoord < maxCoord

		// Get correct min and max coordinates for bounds checking of walls and bots in the way
		if minCoord > maxCoord {
			minCoord, maxCoord = maxCoord, minCoord
		}

		// Check path for walls
		for _, wall := range walls {
			wallCoord := getMotionCoord(wall)
			// Check if wall in correct axis and in middle of path
			// Note: wall at maxCoord is allowed since it only blocks further movement
			if getAxisCoord(startPos) == getAxisCoord(wall) && wallCoord >= minCoord && wallCoord < maxCoord {
				return fmt.Errorf("path blocked by wall at %v", wall)
			}
		}
		// Check path for other bots
		for otherBotId, otherBotPos := range g.Bots {
			if otherBotId != botId { // only check other bots
				botCoord := getMotionCoord(otherBotPos)
				// Check if otherBot in correct axis and in middle of path
				if getAxisCoord(startPos) == getAxisCoord(otherBotPos) && botCoord >= minCoord && botCoord <= maxCoord {
					return fmt.Errorf("path blocked by bot at %v", otherBotPos)
				}
			}
		}

		// Check end position validity: must be against wall, border, or another bot
		if motionIsIncreasing { // Moving towards increasing coordinate (e.g., right or down)
			// At board edge
			if getMotionCoord(endPos) == g.Board.Size-1 {
				return nil
			}
			// Wall just beyond end position
			if slices.ContainsFunc(walls, func(wallPos Position) bool {
				return getMotionCoord(wallPos) == getMotionCoord(endPos) &&
					getAxisCoord(wallPos) == getAxisCoord(endPos)
			}) {
				return nil
			}

			// Bot just beyond end position
			if slices.ContainsFunc(slices.Collect(maps.Values(g.Bots)), func(otherBotPos Position) bool {
				return getMotionCoord(otherBotPos) == getMotionCoord(endPos)+1 &&
					getAxisCoord(otherBotPos) == getAxisCoord(endPos)
			}) {
				return nil
			}

		} else { // Moving towards decreasing coordinate (e.g., left or up)

			// At board edge
			if getMotionCoord(endPos) == 0 {
				return nil
			}

			// Wall just beyond end position
			if slices.ContainsFunc(walls, func(wallPos Position) bool {
				return getMotionCoord(wallPos) == getMotionCoord(endPos)-1 &&
					getAxisCoord(wallPos) == getAxisCoord(endPos)
			}) {
				return nil
			}

			// Bot just beyond end position
			if slices.ContainsFunc(slices.Collect(maps.Values(g.Bots)), func(otherBotPos Position) bool {
				return getMotionCoord(otherBotPos) == getMotionCoord(endPos)-1 &&
					getAxisCoord(otherBotPos) == getAxisCoord(endPos)
			}) {
				return nil
			}
		}

		return fmt.Errorf("end position %v is not against a wall, border, or another bot", endPos)
	}

	getX := func(p Position) BoardDim { return p.X }
	getY := func(p Position) BoardDim { return p.Y }

	// Check path for obstacles and end position validity
	if botEndPos.X == botPos.X {
		// Vertical move
		return checkPathAlongAxis(botPos, botEndPos, g.Board.HWallPos, getY, getX)
	} else {
		// Horizontal move
		return checkPathAlongAxis(botPos, botEndPos, g.Board.VWallPos, getX, getY)
	}
}

// Returns a new Game with the given bot moved to the given position,
// or an error if the move is invalid.
func (g *Game) MoveBot(id BotId, pos Position) (*Game, error) {
	err := g.ValidateMove(id, pos)
	if err != nil {
		return nil, fmt.Errorf("invalid move for bot %d to position %v: %v", id, pos, err)
	}

	// Create new Bots map with updated position for the moved bot
	newBots := make(map[BotId]Position)
	for botId, botPos := range g.Bots {
		if botId == id {
			newBots[botId] = pos
		} else {
			newBots[botId] = botPos
		}
	}

	return NewGame(g.Board, newBots, g.BotTarget)
}

func (g *Game) IsWin() bool {
	targetPos, ok := g.Bots[g.BotTarget.Id]
	if !ok {
		return false
	}
	return targetPos == g.BotTarget.Pos
}

func (g *Game) Equals(o *Game) bool {
	return reflect.DeepEqual(g, o)
}

func (g *Game) String() string {
	return g.render()
}

// Returns whether there is a bot at the given position, and the ID of the bot if present.
func (g *Game) hasBotAtPosition(pos Position) (bool, BotId) {
	for id, botPos := range g.Bots {
		if botPos == pos {
			return true, id
		}
	}
	return false, -1
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
func (g *Game) render() string {
	renderHWall := func(x, y BoardDim) string {
		if g.Board.HasHWallAt(Position{x, y}) {
			return "----"
		}
		return "    "
	}
	// Render a row with horizontal walls
	renderHwallRow := func(y BoardDim) string {
		var rowstr strings.Builder
		rowstr.WriteString("+")
		for x := range g.Board.Size {
			rowstr.WriteString(renderHWall(x, y))
			rowstr.WriteString("+")
		}
		return rowstr.String()
	}
	renderVWall := func(x, y BoardDim) string {
		if g.Board.HasVWallAt(Position{x, y}) {
			return "|"
		}
		return " "
	}
	renderCell := func(x, y BoardDim) string {
		cellPos := Position{x, y}
		hasBot, botId := g.hasBotAtPosition(cellPos)
		if hasBot {
			return fmt.Sprintf(" B%v ", botId)
		}
		if g.BotTarget.Pos == cellPos {
			return fmt.Sprintf(" T%v ", g.BotTarget.Id)
		}
		return "    "
	}
	// Render a row with vertical walls and cell contents
	renderVwallRow := func(y BoardDim) string {
		var rowstr strings.Builder
		// Leftmost VWall
		rowstr.WriteString(renderVWall(-1, y))
		// Iterate over vertical walls and cell contents
		for x := range g.Board.Size {
			rowstr.WriteString(renderCell(x, y))
			rowstr.WriteString(renderVWall(x, y))
		}
		return rowstr.String()
	}

	var boardstr strings.Builder
	// Top HWall border
	boardstr.WriteString(renderHwallRow(-1))
	for y := range g.Board.Size {
		boardstr.WriteString("\n")
		boardstr.WriteString(renderVwallRow(y))
		boardstr.WriteString("\n")
		boardstr.WriteString(renderHwallRow(y))
	}

	return boardstr.String()
}
