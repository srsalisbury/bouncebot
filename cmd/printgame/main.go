package main

import (
	"fmt"

	"github.com/srsalisbury/bouncebot/model"
)

func main() {
	game := model.Game1()
	fmt.Println(game.String())
}
