package model

import (
	"fmt"
	"math/rand"
)

// Tools for building boards and games.

// BuildBoardFromPanels constructs a full Board from four Board panels in clockwise order:
// topLeft, topRight, bottomRight, bottomLeft.
// Each panel should have the same Size and represents a quarter of a full Board,
// with exterior walls on the top and left edges.
/*
   +---- +---- +---- +----     +--------+
   | a   | b   | c   | d    -> | a    b |
   |     |     |     |         |        |
                               |        |
  	                           | d    c |
  	                           +--------+
*/
func BuildBoardFromPanels(a, b, c, d Board) Board {
	if a.Size() != b.Size() || a.Size() != c.Size() || a.Size() != d.Size() {
		panic("all Panels must have the same Size")
	}
	size := a.Size()
	vWalls := make([]Position, 0)
	hWalls := make([]Position, 0)
	possibleTargets := make([]Position, 0)

	appendPanelData := func(p Board, xOffset, yOffset BoardDim) {
		for _, pos := range p.VWalls() {
			vWalls = append(vWalls, Position{X: pos.X + xOffset, Y: pos.Y + yOffset})
		}
		for _, pos := range p.HWalls() {
			hWalls = append(hWalls, Position{X: pos.X + xOffset, Y: pos.Y + yOffset})
		}
		for _, pos := range p.PossibleTargets() {
			possibleTargets = append(possibleTargets, Position{X: pos.X + xOffset, Y: pos.Y + yOffset})
		}
	}
	appendPanelData(a, 0, 0)
	appendPanelData(b.Rotate90cw(), size, 0)
	appendPanelData(c.Rotate90cw().Rotate90cw(), size, size)
	appendPanelData(d.Rotate90cw().Rotate90cw().Rotate90cw(), 0, size)

	return NewBoardWithTargets(size*2, vWalls, hWalls, possibleTargets)
}

// Returns a sample panel.
func Panel1() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|         |                              
		+    +    +    +    +----+    +    +    +
		|                   | []                 
		+    +----+    +    +    +    +    +    +
		|      [] |                              
		+    +    +    +    +    +    +    +    +
		|                               [] |     
		+    +    +    +    +    +    +----+    +
		|                                        
		+    +    +    +    +    +    +    +    +
		|                                        
		+----+    +    +    +    +    +    +    +
		|              | []                      
		+    +    +    +----+    +    +    +----+
		|                                  |     
		+    +    +    +    +    +    +    +    +
	`)
}

// Returns a sample panel.
func Panel2() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|                        |               
		+    +    +----+    +    +    +    +    +
		|         | []                           
		+    +    +    +    +    +    +    +    +
		|                                        
		+    +    +    +    +    +    +    +    +
		|                             | []       
		+    +    +    +    +    +    +----+    +
		|                                        
		+----+    +    +    +----+    +    +    +
		|                     [] |               
		+    +    +    +    +    +    +    +    +
		|      [] |                              
		+    +----+    +    +    +    +    +----+
		|                                  |     
		+    +    +    +    +    +    +    +    +
	`)
}

// Returns a sample panel.
func Panel3() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|                   |                    
		+    +    +    +    +    +    +    +    +
		|    | []                                
		+    +----+    +    +    +    +----+    +
		|                               [] |     
		+    +    +    +    +    +    +    +    +
		|                                        
		+    +    +    +    +    +    +    +    +
		|           [] |                         
		+    +    +----+    +    +    +    +----+
		|                                  | []  
		+----+    +    +    +    +    +    +    +
		|                                        
		+    +    +    +    +    +    +    +----+
		|                                  |     
		+    +    +    +    +    +    +    +    +
	`)
}

// Returns a sample panel.
func Panel4() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|                   |                    
		+    +    +    +    +    +    +    +    +
		|                             | []       
		+    +    +    +    +    +    +----+    +
		|                                        
		+    +----+    +    +    +    +    +    +
		|      [] |                              
		+    +    +    +    +    +----+    +    +
		|                        | []            
		+    +    +    +    +    +    +    +    +
		|           [] |                     [] |
		+    +    +----+    +    +    +    +----+
		|                                        
		+----+    +    +    +    +    +    +----+
		|                                  |     
		+    +    +    +    +    +    +    +    +
	`)
}

// BuildBoard constructs a full Board from four panel IDs in clockwise order.
func BuildBoard(panel1, panel2, panel3, panel4 int) Board {
	makePanel := func(id int) Board {
		switch id {
		case 1:
			return Panel1()
		case 2:
			return Panel2()
		case 3:
			return Panel3()
		case 4:
			return Panel4()
		default:
			panic(fmt.Sprintf("unknown panel id: %d", id))
		}
	}
	return BuildBoardFromPanels(
		makePanel(panel1),
		makePanel(panel2),
		makePanel(panel3),
		makePanel(panel4),
	)
}

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
