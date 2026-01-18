package model

// Game1 returns a sample full-sized game with fixed configuration.
// This is intended for testing only to provide a deterministic game.
//
// Board layout (B0-B3 = bots, T0 = target for bot 0):
//
//	+----+----+----+----+----+----+----+----+----+----+----+----+----+----+----+----+
//	|         |                                            |                        |
//	+    +    +    +    +----+    +    +    +    +    +    +    +    +    +    +    +
//	|                   |                        |                                  |
//	+    +----+    +    +    +    +    +    +    +----+    +    +    +    +----+    +
//	|         |                                                                |    |
//	+    +    +    +    +    +    +    +    +    +    +    +    +    +    +    +    +
//	|                                  |                                            |
//	+    +    +    +    +    +    +----+    +    +    +    +    +    +    +    +    +
//	|                          B0                          |      B3                |
//	+    +    +    +    +    +    +    +    +    +    +----+    +    +    +    +----+
//	|                                                                               |
//	+----+    +    +    +    +    +    +    +    +    +    +    +----+    +    +    +
//	|              |                                            |                   |
//	+    +    +    +----+    +    +    +----+----+    +    +    +    +    +    +    +
//	|                                  |         |                                  |
//	+    +    +    +    +    +    +    +    +    +    +    +    +    +    +    +    +
//	|                        |         |         |                                  |
//	+    +    +----+    +    +----+    +----+----+    +    +    +    +    +    +----+
//	|         |      B2                                                             |
//	+    +    +    +    +    +    +    +    +----+    +    +    +    +----+    +    +
//	|                                       |                        |              |
//	+    +    +    +    +    +    +    +    +    +    +    +    +    +    +    +    +
//	|                                                      |                        |
//	+----+    +    +    +    +    +    +    +    +    +----+    +    +    +    +    +
//	|                                                   B1                |         |
//	+    +    +    +    +----+    +    +    +    +    +    +    +    +    +----+    +
//	|                     T0 |                                                      |
//	+    +    +    +    +    +    +    +    +    +----+    +    +    +    +    +    +
//	|         |                                       |                             |
//	+    +----+    +    +    +    +    +    +    +    +    +    +    +    +    +    +
//	|                             |                             |                   |
//	+----+----+----+----+----+----+----+----+----+----+----+----+----+----+----+----+
func Game1() *Game {
	board := BuildBoard(1, 2, 3, 4)
	bots := map[BotId]Position{
		0: {X: 5, Y: 4},
		1: {X: 10, Y: 12},
		2: {X: 3, Y: 9},
		3: {X: 12, Y: 4},
	}
	target := BotPosition{Id: 0, Pos: Position{X: 4, Y: 13}}
	return mustBuildNewGame(board, bots, target)
}

// Game1OptimalSolution returns a valid 10-move solution for Game1.
// Target is bot 0 at (4, 13), starting at (5, 4).
//
// Solution steps:
//
//	 1. B0: (5,4)   -> (5,0)   up
//	 2. B0: (5,0)   -> (2,0)   left
//	 3. B0: (2,0)   -> (2,8)   down
//	 4. B0: (2,8)   -> (4,8)   right
//	 5. B0: (4,8)   -> (4,12)  down
//	 6. B0: (4,12)  -> (0,12)  left
//	 7. B0: (0,12)  -> (0,15)  down
//	 8. B1: (10,12) -> (0,12)  left
//	 9. B0: (0,15)  -> (0,13)  up
//	10. B0: (0,13)  -> (4,13)  right (reaches target T0)
func Game1OptimalSolution() []BotPosition {
	return []BotPosition{
		{Id: 0, Pos: Position{X: 5, Y: 0}},
		{Id: 0, Pos: Position{X: 2, Y: 0}},
		{Id: 0, Pos: Position{X: 2, Y: 8}},
		{Id: 0, Pos: Position{X: 4, Y: 8}},
		{Id: 0, Pos: Position{X: 4, Y: 12}},
		{Id: 0, Pos: Position{X: 0, Y: 12}},
		{Id: 0, Pos: Position{X: 0, Y: 15}},
		{Id: 1, Pos: Position{X: 0, Y: 12}},
		{Id: 0, Pos: Position{X: 0, Y: 13}},
		{Id: 0, Pos: Position{X: 4, Y: 13}},
	}
}
