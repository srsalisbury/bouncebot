package bouncebot

type Position struct {
	X int8
	Y int8
}

type Board struct {
	// Length of one side of board
	Size int8

	// Vertical and horizontal wall positions
	VWallPos []Position
	HWallPos []Position
}

type BotPosition struct {
	Id  int8
	Pos Position
}
