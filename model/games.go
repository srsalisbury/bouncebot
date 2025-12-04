package model

import "fmt"

// Tools for building boards and games.

func (bp *Board) rotate90cw() *Board {
	newVWalls := make([]Position, len(bp.HWallPos))
	for i, pos := range bp.HWallPos {
		newVWalls[i] = Position{X: bp.Size - 2 - pos.Y, Y: pos.X}
	}
	newHWalls := make([]Position, len(bp.VWallPos))
	for i, pos := range bp.VWallPos {
		newHWalls[i] = Position{X: bp.Size - 1 - pos.Y, Y: pos.X}
	}
	return NewBoard(bp.Size, newVWalls, newHWalls)
}

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
func BuildBoardFromPanels(a, b, c, d *Board) *Board {
	if a.Size != b.Size || a.Size != c.Size || a.Size != d.Size {
		panic("all Panels must have the same Size")
	}
	vWalls := make([]Position, 0)
	hWalls := make([]Position, 0)

	appendPanelWalls := func(p *Board, xOffset, yOffset BoardDim) {
		for _, pos := range p.VWallPos {
			vWalls = append(vWalls, Position{X: pos.X + xOffset, Y: pos.Y + yOffset})
		}
		for _, pos := range p.HWallPos {
			hWalls = append(hWalls, Position{X: pos.X + xOffset, Y: pos.Y + yOffset})
		}
	}
	appendPanelWalls(a, 0, 0)
	appendPanelWalls(b.rotate90cw(), a.Size, 0)
	appendPanelWalls(c.rotate90cw().rotate90cw(), a.Size, a.Size)
	appendPanelWalls(d.rotate90cw().rotate90cw().rotate90cw(), 0, a.Size)

	return NewBoard(a.Size*2, vWalls, hWalls)
}

// Returns a sample panel.
func Panel1() *Board {
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
func Panel2() *Board {
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
func Panel3() *Board {
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
func Panel4() *Board {
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
func BuildBoard(panel1, panel2, panel3, panel4 int) *Board {
	makePanel := func(id int) *Board {
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
func Board1() *Board {
	return BuildBoard(1, 2, 3, 4)
}

// mustBuildNewGame is like NewGame but panics on error.
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
		1: {X: 10, Y: 12},
		2: {X: 3, Y: 9},
		3: {X: 12, Y: 4},
	}
	target := BotPosition{Id: 0, Pos: Position{X: 5, Y: 13}}
	return mustBuildNewGame(board, bots, target)
}
