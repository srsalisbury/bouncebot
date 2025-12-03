package model

import (
	"fmt"
	"slices"
)

// Board represents the static bits of the a game board.
// The board is square with Size x Size cells.
// Bots will occupy cells, while walls exist between cells.
// Implicit walls exist around the edges of the board.
type Board struct {
	// Length of one side of the square board.
	Size BoardDim

	VWallPos []Position // Vertical walls between (X,Y) and (X+1,Y)
	HWallPos []Position // Horizontal walls between (X,Y) and (X,Y+1)
}

// IsBotWithin checks if a given bot position is within the board boundaries.
func (b *Board) IsBotWithin(pos Position) bool {
	return pos.X >= 0 && pos.X < b.Size && pos.Y >= 0 && pos.Y < b.Size
}

// IsVWallWithin checks if a given vertical wall position is within the board boundaries.
func (b *Board) IsVWallWithin(pos Position) bool {
	return pos.X >= 0 && pos.X < b.Size-1 &&
		pos.Y >= 0 && pos.Y < b.Size
}

// IsHWallWithin checks if a given horizontal wall position is within the board boundaries.
func (b *Board) IsHWallWithin(pos Position) bool {
	return pos.X >= 0 && pos.X < b.Size &&
		pos.Y >= 0 && pos.Y < b.Size-1
}

// IsValid checks if the board's wall positions are each within the board boundaries.
func (b *Board) IsValid() error {
	for _, wallPos := range b.VWallPos {
		if !b.IsVWallWithin(wallPos) {
			return fmt.Errorf("vertical wall position %v is out of board boundaries for board of size %d", wallPos, b.Size)
		}
	}
	for _, wallPos := range b.HWallPos {
		if !b.IsHWallWithin(wallPos) {
			return fmt.Errorf("horizontal wall position %v is out of board boundaries for board of size %d", wallPos, b.Size)
		}
	}
	return nil
}

// ValidateBotWithin validates if a given bot position is within the board boundaries.
func (b *Board) ValidateBotWithin(pos Position) error {
	if !b.IsBotWithin(pos) {
		return fmt.Errorf("pos %v is out of board boundaries for board of size %d", pos, b.Size)
	}
	return nil
}

// Checks if there is a vertical wall at the given position
func (b *Board) HasVWallAt(pos Position) bool {
	return pos.X == -1 || pos.X == b.Size-1 || slices.Contains(b.VWallPos, pos)
}

// Checks if there is a horizontal wall at the given position
func (b *Board) HasHWallAt(pos Position) bool {
	return pos.Y == -1 || pos.Y == b.Size-1 || slices.Contains(b.HWallPos, pos)
}
