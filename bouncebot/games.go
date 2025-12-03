package bouncebot

// Returns a sample full-sized board.
func Board1() *Board {
	vW := []Position{
		{X: 1, Y: 0}, {X: 10, Y: 0},
		{X: 3, Y: 1}, {X: 8, Y: 1},
		{X: 1, Y: 2}, {X: 14, Y: 2},
		{X: 6, Y: 3},
		{X: 10, Y: 4},
		{X: 2, Y: 6}, {X: 11, Y: 6},
		{X: 6, Y: 7}, {X: 8, Y: 7},
		{X: 5, Y: 8}, {X: 6, Y: 8}, {X: 8, Y: 8},
		{X: 1, Y: 9},
		{X: 3, Y: 10}, {X: 8, Y: 10},
		{X: 12, Y: 11},
		{X: 5, Y: 13}, {X: 8, Y: 13},
		{X: 2, Y: 14}, {X: 14, Y: 14},
		{X: 6, Y: 15}, {X: 11, Y: 15}}
	hW := []Position{
		{X: 4, Y: 0},
		{X: 1, Y: 1}, {X: 9, Y: 1}, {X: 14, Y: 1},
		{X: 6, Y: 3},
		{X: 10, Y: 4}, {X: 15, Y: 4},
		{X: 0, Y: 5}, {X: 12, Y: 5},
		{X: 3, Y: 6}, {X: 7, Y: 6}, {X: 8, Y: 6},
		{X: 5, Y: 7},
		{X: 7, Y: 8}, {X: 8, Y: 8},
		{X: 1, Y: 9}, {X: 15, Y: 9},
		{X: 4, Y: 10}, {X: 8, Y: 10}, {X: 13, Y: 10},
		{X: 0, Y: 11},
		{X: 5, Y: 12},
		{X: 3, Y: 13}, {X: 9, Y: 13}, {X: 14, Y: 13}}
	return &Board{Size: 16, VWallPos: vW, HWallPos: hW}
}

// MustBuildNewGame is like NewGame but panics on error.
func mustBuildNewGame(board *Board, bots map[BotId]Position, botTarget BotPosition) *Game {
	game, err := NewGame(board, bots, botTarget)
	if err != nil {
		panic(err)
	}
	return game
}

// Returns a sample full-sized game.
func Game1() *Game {
	board := Board1()
	bots := map[BotId]Position{
		0: {X: 5, Y: 4},
		2: {X: 10, Y: 12},
		4: {X: 3, Y: 9},
		6: {X: 12, Y: 4},
	}
	target := BotPosition{Id: 0, Pos: Position{X: 5, Y: 13}}
	return mustBuildNewGame(board, bots, target)
}
