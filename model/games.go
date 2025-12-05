package model

import "fmt"

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

	appendPanelWalls := func(p Board, xOffset, yOffset BoardDim) {
		for _, pos := range p.VWalls() {
			vWalls = append(vWalls, Position{X: pos.X + xOffset, Y: pos.Y + yOffset})
		}
		for _, pos := range p.HWalls() {
			hWalls = append(hWalls, Position{X: pos.X + xOffset, Y: pos.Y + yOffset})
		}
	}
	appendPanelWalls(a, 0, 0)
	appendPanelWalls(b.Rotate90cw(), size, 0)
	appendPanelWalls(c.Rotate90cw().Rotate90cw(), size, size)
	appendPanelWalls(d.Rotate90cw().Rotate90cw().Rotate90cw(), 0, size)

	return NewBoard(size*2, vWalls, hWalls)
}

// Returns a sample panel.
func Panel1() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|         |                              
		+    +    +    +    +----+    +    +    +
		|                   |                    
		+    +----+    +    +    +    +    +    +
		|         |                              
		+    +    +    +    +    +    +    +    +
		|                                  |     
		+    +    +    +    +    +    +----+    +
		|                                        
		+    +    +    +    +    +    +    +    +
		|                                        
		+----+    +    +    +    +    +    +    +
		|              |                         
		+    +    +    +----+    +    +    +----+
		|                                  |     
		+    +    +    +    +    +----+    +    +
	`)
}

// Returns a sample panel.
func Panel2() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|                        |               
		+    +    +----+    +    +    +    +    +
		|         |                              
		+    +    +    +    +    +    +    +    +
		|                                        
		+    +    +    +    +    +    +    +    +
		|                             |          
		+    +    +    +    +    +    +----+    +
		|                                        
		+----+    +    +    +----+    +    +    +
		|                        |               
		+    +    +    +    +    +    +    +    +
		|         |                              
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
		|    |                                   
		+    +----+    +    +    +    +----+    +
		|                                  |     
		+    +    +    +    +    +    +    +    +
		|                                        
		+    +    +    +    +    +    +    +    +
		|              |                         
		+    +    +----+    +    +    +    +----+
		|                                  |     
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
		|                             |          
		+    +    +    +    +    +    +----+    +
		|                                        
		+    +----+    +    +    +    +    +    +
		|         |                              
		+    +    +    +    +    +----+    +    +
		|                        |               
		+    +    +    +    +    +    +    +    +
		|              |                        |
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

// Returns a sample full-sized board.
func Board1() Board {
	return BuildBoard(1, 2, 3, 4)
}

// mustBuildNewGame is like NewGame but panics on error.
func mustBuildNewGame(board Board, bots map[BotId]Position, botTarget BotPosition) *Game {
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
		1: {X: 10, Y: 12},
		2: {X: 3, Y: 9},
		3: {X: 12, Y: 4},
	}
	target := BotPosition{Id: 0, Pos: Position{X: 5, Y: 13}}
	return mustBuildNewGame(board, bots, target)
}
