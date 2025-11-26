package bouncebot

import (
	"fmt"
	"slices"
)

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
func ValidateMove(b Board, botsStart []BotPosition, botEnd BotPosition) bool {
	botStart, err := getBotById(botsStart, botEnd.Id)
	if err != nil {
		return false
	}

	// Check if pos is outside board boundaries, the same as start position,
	// or not in a straight line from start position
	isOutsideBoard := func(pos Position) bool {
		return pos.X < 0 || pos.X >= b.Size || pos.Y < 0 || pos.Y >= b.Size
	}
	isSamePosition := func(pos1, pos2 Position) bool {
		return pos1.X == pos2.X && pos1.Y == pos2.Y
	}
	isStraightLine := func(pos1, pos2 Position) bool {
		return pos1.X == pos2.X || pos1.Y == pos2.Y
	}

	if isOutsideBoard(botEnd.Pos) ||
		isSamePosition(botEnd.Pos, botStart.Pos) ||
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

	getX := func(p Position) int8 {
		return p.X
	}
	getY := func(p Position) int8 {
		return p.Y
	}

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

	// Check path for obstacles and end position validity
	if botEnd.Pos.X == botStart.Pos.X {
		// Vertical move
		return checkPathAlongAxis(botStart.Pos, botEnd.Pos, b.HWallPos, getY, getX)
	} else {
		// Horizontal move
		return checkPathAlongAxis(botStart.Pos, botEnd.Pos, b.VWallPos, getX, getY)
	}
}
