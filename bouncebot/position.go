package bouncebot

import "fmt"

type BoardDim int8
type Position struct {
	X BoardDim
	Y BoardDim
}

func (p Position) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}
