package main

import (
	"fmt"

	"github.com/srsalisbury/bouncebot/model"
)

func main() {
	game := model.NewRandomGame()
	fmt.Println(game.String())
	fmt.Println(game.Board.String())
}
