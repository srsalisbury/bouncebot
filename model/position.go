package model

import (
	"fmt"

	pb "github.com/srsalisbury/bouncebot/proto"
)

type BoardDim int8
type Position struct {
	X BoardDim
	Y BoardDim
}

func NewPositionFromProto(pp *pb.Position) Position {
	return Position{
		X: BoardDim(pp.X),
		Y: BoardDim(pp.Y),
	}
}

func (p Position) ToProto() *pb.Position {
	return &pb.Position{
		X: int32(p.X),
		Y: int32(p.Y),
	}
}

func (p Position) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}
