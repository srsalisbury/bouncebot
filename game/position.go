package bouncebot

type Position struct {
	X int8
	Y int8
}

type BotPosition struct {
	Id  int8
	Pos Position
}

type BotPositionMap map[int8]Position
