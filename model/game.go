package model

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	pb "github.com/srsalisbury/bouncebot/proto"
)

type BotId int8
type BotPosition struct {
	Id  BotId
	Pos Position
}

func (bp BotPosition) String() string {
	return fmt.Sprintf("Bot %d at %v", bp.Id, bp.Pos)
}

func NewBotPosition(id BotId, x, y BoardDim) BotPosition {
	return BotPosition{
		Id:  id,
		Pos: Position{X: x, Y: y},
	}
}

func NewBotPositionFromProto(bpp *pb.BotPos) BotPosition {
	return BotPosition{
		Id:  BotId(bpp.Id),
		Pos: NewPositionFromProto(bpp.Pos),
	}
}

func NewBotPositionsFromProto(bpp []*pb.BotPos) []BotPosition {
	bps := make([]BotPosition, len(bpp))
	for i, bp := range bpp {
		bps[i] = NewBotPositionFromProto(bp)
	}
	return bps
}

func (bp BotPosition) ToProto() *pb.BotPos {
	return &pb.BotPos{
		Id:  int32(bp.Id),
		Pos: bp.Pos.ToProto(),
	}
}

// A full game state, including board, bot positions, and target bot position.
type Game struct {
	Board Board
	Bots  map[BotId]Position
	// Where the given bot needs to end up.
	Target BotPosition
}

func NewGameFromProto(gp *pb.Game) *Game {
	bots := make(map[BotId]Position)
	for _, bot := range gp.Bots {
		bots[BotId(bot.Id)] = NewPositionFromProto(bot.Pos)
	}
	return &Game{
		Board:  NewBoardFromProto(gp.Board),
		Bots:   bots,
		Target: NewBotPositionFromProto(gp.Target),
	}
}

func (g *Game) ToProto() *pb.Game {
	bots := []*pb.BotPos{}
	for id, pos := range g.Bots {
		bots = append(bots, BotPosition{Id: id, Pos: pos}.ToProto())
	}
	return &pb.Game{
		Board:  g.Board.ToProto(),
		Bots:   bots,
		Target: g.Target.ToProto(),
	}
}

// Creates a new Game instance, validating the inputs.
func NewGame(board Board, bots map[BotId]Position, target BotPosition) (*Game, error) {
	err := board.IsValid()
	if err != nil {
		return nil, err
	}
	// Validate that target.Id exists in bots
	if _, ok := bots[target.Id]; !ok {
		return nil, fmt.Errorf("target.Id %d not found in bots", target.Id)
	}
	err = board.ValidateBotWithin(target.Pos)
	if err != nil {
		return nil, fmt.Errorf("target %v", err)
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
		Board:  board,
		Bots:   bots,
		Target: target,
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
	// This handles the logic for either horizontal or vertical moves by
	// extracting the relevant coordinate from the relevant positions.
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
			if getMotionCoord(endPos) == g.Board.Size()-1 {
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
		return checkPathAlongAxis(botPos, botEndPos, g.Board.HWalls(), getY, getX)
	} else {
		// Horizontal move
		return checkPathAlongAxis(botPos, botEndPos, g.Board.VWalls(), getX, getY)
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

	return NewGame(g.Board, newBots, g.Target)
}

func (g *Game) IsWin() bool {
	targetPos, ok := g.Bots[g.Target.Id]
	if !ok {
		return false
	}
	return targetPos == g.Target.Pos
}

func (g *Game) Equals(o *Game) bool {
	return reflect.DeepEqual(g, o)
}

func (g *Game) String() string {
	return renderGame(g.Board, g.Bots, &g.Target)
}

// Returns whether there is a bot at the given position, and the ID of the bot if present.
func hasBotAtPosition(bots map[BotId]Position, pos Position) (bool, BotId) {
	if bots == nil {
		return false, -1
	}
	for id, botPos := range bots {
		if botPos == pos {
			return true, id
		}
	}
	return false, -1
}

// Returns isValid and the resulting game after applying the given moves.
func (g *Game) CheckSolution(moves []BotPosition) (bool, *Game) {
	currentGame := g
	for _, move := range moves {
		var err error
		currentGame, err = currentGame.MoveBot(move.Id, move.Pos)
		if err != nil {
			//			return fmt.Errorf("invalid move %d (%v): %v", moveIdx, move, err)
			return false, nil
		}
	}
	if !currentGame.IsWin() {
		return false, nil
	}
	return true, currentGame
}
