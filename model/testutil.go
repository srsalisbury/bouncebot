package model

// Game1 returns a sample full-sized game with fixed configuration.
// This is intended for testing only to provide a deterministic game.
func Game1() *Game {
	board := BuildBoard(1, 2, 3, 4)
	bots := map[BotId]Position{
		0: {X: 5, Y: 4},
		1: {X: 10, Y: 12},
		2: {X: 3, Y: 9},
		3: {X: 12, Y: 4},
	}
	target := BotPosition{Id: 0, Pos: Position{X: 5, Y: 13}}
	return mustBuildNewGame(board, bots, target)
}

// Game1Solution returns a valid 7-move solution for Game1.
// Target is bot 0 at (5, 13), starting at (5, 4).
// Moves: Bot 1 left, then Bot 0: up, left, down, left, up, right
func Game1Solution() []BotPosition {
	return []BotPosition{
		{Id: 1, Pos: Position{X: 0, Y: 12}},
		{Id: 0, Pos: Position{X: 5, Y: 0}},
		{Id: 0, Pos: Position{X: 2, Y: 0}},
		{Id: 0, Pos: Position{X: 2, Y: 15}},
		{Id: 0, Pos: Position{X: 0, Y: 15}},
		{Id: 0, Pos: Position{X: 0, Y: 13}},
		{Id: 0, Pos: Position{X: 5, Y: 13}},
	}
}
