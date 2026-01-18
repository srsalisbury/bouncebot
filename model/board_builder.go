package model

import (
	"fmt"
	"math/rand"
)

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

// Panel** functions each return a sample panel that can be composed to make a full board..
func Panel1A() Board {
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

func Panel2A() Board {
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

func Panel3A() Board {
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

func Panel4A() Board {
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

func Panel1B() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|                        |
		+    +    +    +    +    +    +    +    +
		|                               [] |
		+    +----+    +    +    +    +----+    +
		|    | []  
		+    +    +    +    +    +    +    +    +
		|                                   
		+    +    +    +    +    +    +    +    +
		|
		+    +    +    +    +    +    +----+    +
		|                               [] |
		+----+    +    +    +    +    +    +    +
		|              | []
		+    +    +    +----+    +    +    +----+
		|                                  |
		+    +    +    +    +    +    +    +    +
	`)
}

func Panel2B() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|                   |     
		+    +    +    +    +    +    +    +    +
		|                          [] |
		+    +    +    +    +    +----+    +    +
		|    | []
		+    +----+    +    +    +    +    +    +
		|                                 
		+----+    +    +    +    +    +----+    +
		|                             | []
		+    +    +    +    +    +    +    +    +
		|                         
		+    +    +----+    +    +    +    +    +
		|           [] |
		+    +    +    +    +    +    +    +----+
		|                                  |
		+    +    +    +    +    +    +    +    +
	`)
}

func Panel3B() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|                        |
		+    +    +    +    +    +    +    +    +
		|           [] |                  
		+    +    +----+    +    +    +    +    +
		|
		+    +    +    +    +    +    +    +    +
		|    | []  
		+    +----+    +    +    +    +----+    +
		|                             | []
		+----+    +    +    +    +    +    +    +
		|                                        
		+    +    +    +    +    +----+    +    +
		|                          [] |
		+    +    +    +    +    +    +    +----+
		|                [] |              |
		+    +    +    +----+    +    +    +    +
	`)
}

func Panel4B() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|                   |
		+    +    +    +    +    +    +    +    +
		|        
		+    +    +    +    +    +    +    +    +
		|                          [] |     
		+    +    +    +    +    +----+    +    +
		|
		+    +    +----+    +    +    +    +    +
		|           [] |
		+----+    +    +    +    +    +    +    +
		|                                  | []
		+    +----+    +    +    +    +    +----+
		|    | []
		+    +    +    +    +    +    +    +----+
		|                                  |
		+    +    +    +    +    +    +    +    +
	`)
}
func Panel1C() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|         |               
		+    +    +    +----+    +    +    +    +
		|              | []                 
		+    +    +    +    +    +    +    +    +
		|          
		+    +    +    +    +    +    +    +    +
		|                               [] |
		+    +    +    +    +    +    +----+    +
		|    | []
		+    +----+    +    +    +    +    +    +
		|                                   
		+    +    +    +    +----+    +    +    +
		|                     [] |
		+----+    +    +    +    +    +    +----+
		|                                  |
		+    +    +    +    +    +    +    +    +
	`)
}

func Panel2C() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|                             |
		+    +    +    +    +    +    +    +    +
		|                              
		+    +    +    +----+    +    +    +    +
		|              | []
		+    +    +    +    +    +    +    +    +
		|                        | []      
		+----+    +----+    +    +----+    +    +
		|           [] |                  
		+    +    +    +    +    +    +    +    +
		|                     [] |
		+    +    +    +    +----+    +    +    +
		|               
		+    +    +    +    +    +    +    +----+
		|                                  |
		+    +    +    +    +    +    +    +    +
	`)
}

func Panel3C() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|              |          
		+    +    +    +    +    +    +    +    +
		|                        | []
		+    +    +    +    +    +----+    +    +
		|                                    [] |
		+    +    +    +    +    +    +    +----+
		|          
		+----+    +    +    +    +    +    +    +
		|                [] |             
		+    +    +    +----+    +    +----+    +
		|                             | []       
		+    +----+    +    +    +    +    +    +
		|      [] |                    
		+    +    +    +    +    +    +    +----+
		|                                  |
		+    +    +    +    +    +    +    +    +
	`)
}

func Panel4C() Board {
	return MustParsePanelString(`
		+----+----+----+----+----+----+----+----+
		|         |          
		+    +    +    +    +----+    +    +    +
		|                     [] |
		+    +    +    +    +    +    +    +    +
		|                                   
		+    +    +    +    +    +    +    +    +
		|    | []
		+    +----+    +    +    +    +    +    +
		|               
		+    +    +    +    +    +----+    +    +
		|                        | []          
		+----+    +    +    +    +    +    +    +
		|                [] |
		+    +    +    +----+    +    +    +----+
		|                                  |
		+    +    +    +    +    +    +    +    +
	`)
}

// makePanel returns a panel for the given panel ID (1..12).
func makePanel(id int) Board {
	switch id {
	case 1:
		return Panel1A()
	case 2:
		return Panel2A()
	case 3:
		return Panel3A()
	case 4:
		return Panel4A()
	case 5:
		return Panel1B()
	case 6:
		return Panel2B()
	case 7:
		return Panel3B()
	case 8:
		return Panel4B()
	case 9:
		return Panel1C()
	case 10:
		return Panel2C()
	case 11:
		return Panel3C()
	case 12:
		return Panel4C()
	default:
		panic(fmt.Sprintf("unknown panel id: %d", id))
	}
}

func NewRandomBoard() Board {
	// Shuffle panels 1-12 into random positions and choose the first four.
	panels := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	rand.Shuffle(len(panels), func(i, j int) {
		panels[i], panels[j] = panels[j], panels[i]
	})
	return BuildBoard(panels[0], panels[1], panels[2], panels[3])
}

// BuildBoard constructs a full Board from four panel IDs in clockwise order.
func BuildBoard(panel1, panel2, panel3, panel4 int) Board {
	return BuildBoardFromPanels(
		makePanel(panel1),
		makePanel(panel2),
		makePanel(panel3),
		makePanel(panel4),
	)
}
