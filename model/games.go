package model

import "math/rand"

// mustBuildNewGame is like NewGame but panics on error.
func mustBuildNewGame(board Board, bots map[BotId]Position, botTarget BotPosition) *Game {
	game, err := NewGame(board, bots, botTarget)
	if err != nil {
		panic(err)
	}
	return game
}

// NewRandomGame generates a new game with random configuration:
// - Random permutation of panels 1-4
// - Random target from possible target locations
// - Random robot placement (avoiding each other, target, and center cells)
func NewRandomGame() *Game {
	// Shuffle panels 1-4 into random positions
	panels := []int{1, 2, 3, 4}
	rand.Shuffle(len(panels), func(i, j int) {
		panels[i], panels[j] = panels[j], panels[i]
	})
	board := BuildBoard(panels[0], panels[1], panels[2], panels[3])

	// Pick a random target from possible targets
	possibleTargets := board.PossibleTargets()
	if len(possibleTargets) == 0 {
		panic("board has no possible targets")
	}
	targetPos := possibleTargets[rand.Intn(len(possibleTargets))]
	targetBotId := BotId(rand.Intn(4))
	target := BotPosition{Id: targetBotId, Pos: targetPos}

	// Place robots randomly, avoiding:
	// - Each other
	// - The target position
	// - The center 4 cells (for a 16x16 board: (7,7), (8,7), (7,8), (8,8))
	size := board.Size()
	centerCells := []Position{
		{X: size/2 - 1, Y: size/2 - 1},
		{X: size / 2, Y: size/2 - 1},
		{X: size/2 - 1, Y: size / 2},
		{X: size / 2, Y: size / 2},
	}

	isOccupied := func(pos Position, placedBots map[BotId]Position) bool {
		// Check if position is the target
		if pos == targetPos {
			return true
		}
		// Check if position is a center cell
		for _, center := range centerCells {
			if pos == center {
				return true
			}
		}
		// Check if position is already occupied by another bot
		for _, botPos := range placedBots {
			if pos == botPos {
				return true
			}
		}
		return false
	}

	bots := make(map[BotId]Position)
	for botId := BotId(0); botId < 4; botId++ {
		// Find a random unoccupied position
		for {
			pos := Position{
				X: BoardDim(rand.Intn(int(size))),
				Y: BoardDim(rand.Intn(int(size))),
			}
			if !isOccupied(pos, bots) {
				bots[botId] = pos
				break
			}
		}
	}

	return mustBuildNewGame(board, bots, target)
}

// NewContinuationGame creates a new game continuing from the previous game:
// - Same board configuration
// - Same robot positions (keeps robots where they ended up)
// - New random target position and robot
func NewContinuationGame(prev *Game) *Game {
	if prev == nil {
		return NewRandomGame()
	}

	// Copy the bot positions
	bots := make(map[BotId]Position)
	for id, pos := range prev.Bots {
		bots[id] = pos
	}

	// Pick a new random target from possible targets
	// Avoid placing target where a robot already is
	possibleTargets := prev.Board.PossibleTargets()
	if len(possibleTargets) == 0 {
		panic("board has no possible targets")
	}

	// Filter out positions occupied by robots
	availableTargets := make([]Position, 0, len(possibleTargets))
	for _, pos := range possibleTargets {
		occupied := false
		for _, botPos := range bots {
			if pos == botPos {
				occupied = true
				break
			}
		}
		if !occupied {
			availableTargets = append(availableTargets, pos)
		}
	}

	// If all targets are occupied (unlikely), fall back to all possible targets
	if len(availableTargets) == 0 {
		availableTargets = possibleTargets
	}

	targetPos := availableTargets[rand.Intn(len(availableTargets))]
	targetBotId := BotId(rand.Intn(4))
	target := BotPosition{Id: targetBotId, Pos: targetPos}

	return mustBuildNewGame(prev.Board, bots, target)
}
