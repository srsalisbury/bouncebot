package model

import (
	"fmt"
	"slices"

	pb "github.com/srsalisbury/bouncebot/proto"
)

// Board represents the static bits of the a game board.
// The board is square with Size x Size cells.
// Bots will occupy cells, while walls exist between cells.
// Implicit walls exist around the edges of the board.
type Board interface {
	ToProto() *pb.Board
	String() string
	Size() BoardDim

	// Returns all horizontal wall positions
	HWalls() []Position
	// Returns all vertical wall positions
	VWalls() []Position

	// Checks if there is a vertical wall at the given position
	HasVWallAt(pos Position) bool

	// Checks if there is a horizontal wall at the given position
	HasHWallAt(pos Position) bool

	// IsBotWithin checks if a given bot position is within the board boundaries.
	IsBotWithin(pos Position) bool

	// ValidateBotWithin validates if a given bot position is within the board boundaries.
	ValidateBotWithin(pos Position) error

	// IsVWallWithin checks if a given vertical wall position is within the board boundaries.
	IsVWallWithin(pos Position) bool

	// IsHWallWithin checks if a given horizontal wall position is within the board boundaries.
	IsHWallWithin(pos Position) bool

	// IsValid checks if the board's wall positions are each within the board boundaries.
	IsValid() error

	// Rotate90cw returns a new Board instance that is rotated 90 degrees clockwise.
	// Generally useful for building larger boards from smaller panels.
	Rotate90cw() Board
}

func NewBoardFromProto(bp *pb.Board) Board {
	vWalls := make([]Position, len(bp.VWalls))
	for i, vp := range bp.VWalls {
		vWalls[i] = NewPositionFromProto(vp)
	}
	hWalls := make([]Position, len(bp.HWalls))
	for i, hp := range bp.HWalls {
		hWalls[i] = NewPositionFromProto(hp)
	}
	return NewBoard(BoardDim(bp.Size), vWalls, hWalls)
}

func NewBoard(size BoardDim, vWalls, hWalls []Position) Board {
	return &board{size: size, vWallPos: vWalls, hWallPos: hWalls, isPanel: false}
}

func NewPanel(size BoardDim, vWalls, hWalls []Position) Board {
	return &board{size: size, vWallPos: vWalls, hWallPos: hWalls, isPanel: true}
}

type board struct {
	// Length of one side of the square board.
	size BoardDim

	vWallPos []Position // Vertical walls between (X,Y) and (X+1,Y)
	hWallPos []Position // Horizontal walls between (X,Y) and (X,Y+1)

	// Whether this board is a panel (for rendering purposes, as panels don't have
	// implicit walls on their right and bottom edges)
	isPanel bool
}

func (b *board) ToProto() *pb.Board {
	vWalls := make([]*pb.Position, len(b.VWalls()))
	for i, vp := range b.VWalls() {
		vWalls[i] = vp.ToProto()
	}
	hWalls := make([]*pb.Position, len(b.HWalls()))
	for i, hp := range b.HWalls() {
		hWalls[i] = hp.ToProto()
	}
	return &pb.Board{
		Size:   int32(b.Size()),
		VWalls: vWalls,
		HWalls: hWalls,
	}
}

func (b *board) Size() BoardDim {
	return b.size
}

func (b *board) String() string {
	return renderBoard(b)
}

func (b *board) HWalls() []Position {
	// Make a copy to prevent external modification
	result := make([]Position, len(b.hWallPos))
	copy(result, b.hWallPos)
	return result
}

func (b *board) VWalls() []Position {
	// Make a copy to prevent external modification
	result := make([]Position, len(b.vWallPos))
	copy(result, b.vWallPos)
	return result
}

func (b *board) IsBotWithin(pos Position) bool {
	return pos.X >= 0 && pos.X < b.size && pos.Y >= 0 && pos.Y < b.size
}

func (b *board) IsVWallWithin(pos Position) bool {
	xsize := b.size
	if b.isPanel {
		// Panels can have vwalls on the right edge
		xsize++
	}
	return pos.X >= 0 && pos.X < xsize-1 &&
		pos.Y >= 0 && pos.Y < b.size
}

func (b *board) IsHWallWithin(pos Position) bool {
	ysize := b.size
	if b.isPanel {
		// Panels can have hwalls on the bottom edge
		ysize++
	}
	return pos.X >= 0 && pos.X < b.size &&
		pos.Y >= 0 && pos.Y < ysize-1
}

func (b *board) IsValid() error {
	for _, wallPos := range b.vWallPos {
		if !b.IsVWallWithin(wallPos) {
			return fmt.Errorf("vertical wall position %v is out of board boundaries for board of size %d", wallPos, b.Size())
		}
	}
	for _, wallPos := range b.hWallPos {
		if !b.IsHWallWithin(wallPos) {
			return fmt.Errorf("horizontal wall position %v is out of board boundaries for board of size %d", wallPos, b.Size())
		}
	}
	return nil
}

func (b *board) ValidateBotWithin(pos Position) error {
	if !b.IsBotWithin(pos) {
		return fmt.Errorf("pos %v is out of board boundaries for board of size %d", pos, b.Size())
	}
	return nil
}

func (b *board) HasVWallAt(pos Position) bool {
	return pos.X == -1 || (!b.isPanel && pos.X == b.size-1) || slices.Contains(b.vWallPos, pos)
}

func (b *board) HasHWallAt(pos Position) bool {
	return pos.Y == -1 || (!b.isPanel && pos.Y == b.size-1) || slices.Contains(b.hWallPos, pos)
}

func (b *board) Rotate90cw() Board {
	newVWalls := make([]Position, len(b.hWallPos))
	for i, pos := range b.hWallPos {
		newVWalls[i] = Position{X: b.size - 2 - pos.Y, Y: pos.X}
	}
	newHWalls := make([]Position, len(b.vWallPos))
	for i, pos := range b.vWallPos {
		newHWalls[i] = Position{X: b.size - 1 - pos.Y, Y: pos.X}
	}
	if b.isPanel {
		return NewPanel(b.size, newVWalls, newHWalls)
	}
	return NewBoard(b.size, newVWalls, newHWalls)
}
