package model

import (
	"encoding/json"
	"fmt"
	"maps"
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

// Direction represents a movement direction.
type Direction string

const (
	Up    Direction = "up"
	Down  Direction = "down"
	Left  Direction = "left"
	Right Direction = "right"
)

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

// MarshalJSON implements json.Marshaler for Game.
// Converts to proto format for serialization since Board is an interface.
func (g *Game) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.ToProto())
}

// UnmarshalJSON implements json.Unmarshaler for Game.
// Converts from proto format since Board is an interface.
func (g *Game) UnmarshalJSON(data []byte) error {
	var gp pb.Game
	if err := json.Unmarshal(data, &gp); err != nil {
		return err
	}
	game := NewGameFromProto(&gp)
	*g = *game
	return nil
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

// ComputeDestination calculates where a bot will end up when sliding in a direction.
// The bot slides until it hits a wall, board edge, or another bot.
func (g *Game) ComputeDestination(botId BotId, dir Direction) (Position, error) {
	pos, ok := g.Bots[botId]
	if !ok {
		return Position{}, fmt.Errorf("bot with id %d not found", botId)
	}

	var dx, dy BoardDim
	switch dir {
	case Up:
		dy = -1
	case Down:
		dy = 1
	case Left:
		dx = -1
	case Right:
		dx = 1
	default:
		return Position{}, fmt.Errorf("invalid direction: %s", dir)
	}

	// Slide until hitting an obstacle
	for {
		// Check for wall blocking movement
		if g.hasWallBlocking(pos, dir) {
			break
		}

		nextPos := Position{X: pos.X + dx, Y: pos.Y + dy}

		// Check for other bots
		blocked := false
		for otherId, otherPos := range g.Bots {
			if otherId != botId && otherPos == nextPos {
				blocked = true
				break
			}
		}
		if blocked {
			break
		}

		pos = nextPos
	}

	return pos, nil
}

// hasWallBlocking checks if there's a wall or board edge blocking movement from pos in dir.
func (g *Game) hasWallBlocking(pos Position, dir Direction) bool {
	switch dir {
	case Up:
		if pos.Y == 0 {
			return true
		}
		for _, w := range g.Board.HWalls() {
			if w.X == pos.X && w.Y == pos.Y-1 {
				return true
			}
		}
	case Down:
		if pos.Y == g.Board.Size()-1 {
			return true
		}
		for _, w := range g.Board.HWalls() {
			if w.X == pos.X && w.Y == pos.Y {
				return true
			}
		}
	case Left:
		if pos.X == 0 {
			return true
		}
		for _, w := range g.Board.VWalls() {
			if w.X == pos.X-1 && w.Y == pos.Y {
				return true
			}
		}
	case Right:
		if pos.X == g.Board.Size()-1 {
			return true
		}
		for _, w := range g.Board.VWalls() {
			if w.X == pos.X && w.Y == pos.Y {
				return true
			}
		}
	}
	return false
}

// ValidateMove checks if a bot's intended move is valid based on the game rules.
// The inputs are the board state, the starting positions of all bots, and a bot's intended end position.
func (g *Game) ValidateMove(botId BotId, botEndPos Position) error {
	botPos, ok := g.Bots[botId]
	if !ok {
		return fmt.Errorf("bot with id %d not found", botId)
	}

	if botEndPos == botPos {
		return fmt.Errorf("end position %v is the same as start position", botEndPos)
	}

	// Determine direction from start to end position
	var dir Direction
	switch {
	case botEndPos.X == botPos.X && botEndPos.Y < botPos.Y:
		dir = Up
	case botEndPos.X == botPos.X && botEndPos.Y > botPos.Y:
		dir = Down
	case botEndPos.Y == botPos.Y && botEndPos.X < botPos.X:
		dir = Left
	case botEndPos.Y == botPos.Y && botEndPos.X > botPos.X:
		dir = Right
	default:
		return fmt.Errorf("move from %v to %v is not in a straight line", botPos, botEndPos)
	}

	// Compute where the bot would actually end up
	actualEnd, err := g.ComputeDestination(botId, dir)
	if err != nil {
		return err
	}

	if actualEnd != botEndPos {
		return fmt.Errorf("move to %v is invalid; bot would end at %v", botEndPos, actualEnd)
	}

	return nil
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
	if g.Board.Size() != o.Board.Size() {
		return false
	}
	if !positionsEqualUnordered(g.Board.VWalls(), o.Board.VWalls()) {
		return false
	}
	if !positionsEqualUnordered(g.Board.HWalls(), o.Board.HWalls()) {
		return false
	}
	if !maps.Equal(g.Bots, o.Bots) {
		return false
	}
	if g.Target != o.Target {
		return false
	}
	return true
}

// positionsEqualUnordered returns true if two position slices contain the same elements,
// regardless of order.
func positionsEqualUnordered(a, b []Position) bool {
	if len(a) != len(b) {
		return false
	}
	aSorted := slices.Clone(a)
	bSorted := slices.Clone(b)
	slices.SortFunc(aSorted, comparePositions)
	slices.SortFunc(bSorted, comparePositions)
	return slices.Equal(aSorted, bSorted)
}

func comparePositions(a, b Position) int {
	if a.X != b.X {
		return int(a.X - b.X)
	}
	return int(a.Y - b.Y)
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
