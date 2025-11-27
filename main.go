package main

import (
	"fmt"

	game "salisburyclan.com/bouncebot/game"
)

func main() {
	vW := []game.Position{
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
	hW := []game.Position{
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
	bP := []game.BotPosition{
		{Id: 0, Pos: game.Position{X: 5, Y: 4}},
		{Id: 2, Pos: game.Position{X: 10, Y: 12}},
		{Id: 4, Pos: game.Position{X: 3, Y: 9}},
		{Id: 6, Pos: game.Position{X: 12, Y: 4}}}
	board := game.Board{Size: 16, VWallPos: vW, HWallPos: hW}
	goal := game.BotPosition{Id: 0, Pos: game.Position{X: 5, Y: 13}}
	fmt.Println(game.Render(board, goal, bP))
}
