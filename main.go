package main

import (
	"fmt"

	"salisburyclan.com/bouncebot/bouncebot"
)

func main() {
	game := bouncebot.Game1()
	fmt.Println(game.String())
}
