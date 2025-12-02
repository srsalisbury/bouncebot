package bouncebot

import (
	"fmt"
	"reflect"
	"slices"
)

// A full game state, including board, starting bot positions, and target bot position.
type Game struct {
	B         *Board
	BotsStart BotPositionMap
	// Where the given bot needs to end up.
	BotTarget BotPosition
}

func NewGame(b *Board, botsStart BotPositionMap, botTarget BotPosition) (*Game, error) {
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

func getBotById(botPositions []BotPosition, id int8) (BotPosition, error) {
	for _, bot := range botPositions {
		if bot.Id == id {
			return bot, nil
		}
	}
	return BotPosition{}, fmt.Errorf("bot with id %d not found", id)
}

// ValidateMove checks if a bot's intended move is valid based on the game rules.
// The inputs are the board state, the starting positions of all bots, and a bot's intended end position.
// TODO: Do we need to describe why the move isn't valid?
// TODO: Consider moving this as a method of Game.
func ValidateMove(b *Board, botsStart BotPositionMap, botEnd BotPosition) bool {
	botsStartSlice := make([]BotPosition, 0, len(botsStart))
	for id, pos := range botsStart {
		botsStartSlice = append(botsStartSlice, BotPosition{Id: id, Pos: pos})
	}
	return validateMove(b, botsStartSlice, botEnd)
}

// TODO: Rewrite this function to directly use BotPositionMap instead of converting to slice.
// Then integrate it back into ValidateMove above.
func validateMove(b *Board, botsStart []BotPosition, botEnd BotPosition) bool {
	botStart, err := getBotById(botsStart, botEnd.Id)
	if err != nil {
		return false
	}

	isStraightLine := func(pos1, pos2 Position) bool {
		return pos1.X == pos2.X || pos1.Y == pos2.Y
	}

	if !b.IsBotWithin(botEnd.Pos) ||
		botEnd.Pos == botStart.Pos ||
		!isStraightLine(botEnd.Pos, botStart.Pos) {
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
		for _, otherBot := range botsStart {
			if otherBot.Id != botStart.Id { // only check other bots
				botCoord := getMotionCoord(otherBot.Pos)
				// Check if otherBot in correct axis and in middle of path
				if getAxisCoord(start) == getAxisCoord(otherBot.Pos) && botCoord >= minCoord && botCoord <= maxCoord {
					return false
				}
			}
		}

		if direction == 1 { // Moving towards increasing coordinate (e.g., right or down)

			// At board edge
			if getMotionCoord(end) == b.Size-1 {
				return true
			}
			// Wall just beyond end position
			if slices.ContainsFunc(walls, func(p Position) bool {
				return getMotionCoord(p) == getMotionCoord(end) && getAxisCoord(p) == getAxisCoord(end)
			}) {
				return true
			}

			// Bot just beyond end position
			if slices.ContainsFunc(botsStart, func(bp BotPosition) bool {
				return getMotionCoord(bp.Pos)-1 == getMotionCoord(end) && getAxisCoord(bp.Pos) == getAxisCoord(end)
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
			if slices.ContainsFunc(botsStart, func(bp BotPosition) bool {
				return getMotionCoord(bp.Pos)+1 == getMotionCoord(end) && getAxisCoord(bp.Pos) == getAxisCoord(end)
			}) {
				return true
			}
		}

		return false
	}

	getX := func(p Position) int8 { return p.X }
	getY := func(p Position) int8 { return p.Y }

	// Check path for obstacles and end position validity
	if botEnd.Pos.X == botStart.Pos.X {
		// Vertical move
		return checkPathAlongAxis(botStart.Pos, botEnd.Pos, b.HWallPos, getY, getX)
	} else {
		// Horizontal move
		return checkPathAlongAxis(botStart.Pos, botEnd.Pos, b.VWallPos, getX, getY)
	}
}

// Returns a new Game with the given bot moved to the given position,
// or an error if the move is invalid.
func (g *Game) MoveBot(id int8, pos Position) (*Game, error) {
	botEnd := BotPosition{Id: id, Pos: pos}
	if !ValidateMove(g.B, g.BotsStart, botEnd) {
		return nil, fmt.Errorf("invalid move for bot %d to position (%d, %d)", id, pos.X, pos.Y)
	}

	// Create new BotsStart map with updated position for the moved bot
	newBotsStart := make(BotPositionMap)
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
