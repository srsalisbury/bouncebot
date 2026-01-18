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
