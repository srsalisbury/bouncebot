package model

import "fmt"

// ComputeDestination calculates where a bot will end up when sliding in a direction.
// The bot slides until it hits a wall, board edge, or another bot.
func (g *Game) ComputeDestination(botId BotId, dir Direction) (Position, error) {
	pos, ok := g.Bots[botId]
	if !ok {
		return Position{}, fmt.Errorf("bot with id %d not found", botId)
	}

	var dx, dy BoardDim
	switch dir {
	case Up:
		dy = -1
	case Down:
		dy = 1
	case Left:
		dx = -1
	case Right:
		dx = 1
	default:
		return Position{}, fmt.Errorf("invalid direction: %s", dir)
	}

	// Slide until hitting an obstacle
	for {
		// Check for wall blocking movement
		if g.hasWallBlocking(pos, dir) {
			break
		}

		nextPos := Position{X: pos.X + dx, Y: pos.Y + dy}

		// Check for other bots
		blocked := false
		for otherId, otherPos := range g.Bots {
			if otherId != botId && otherPos == nextPos {
				blocked = true
				break
			}
		}
		if blocked {
			break
		}

		pos = nextPos
	}

	return pos, nil
}

// hasWallBlocking checks if there's a wall or board edge blocking movement from pos in dir.
func (g *Game) hasWallBlocking(pos Position, dir Direction) bool {
	switch dir {
	case Up:
		if pos.Y == 0 {
			return true
		}
		for _, w := range g.Board.HWalls() {
			if w.X == pos.X && w.Y == pos.Y-1 {
				return true
			}
		}
	case Down:
		if pos.Y == g.Board.Size()-1 {
			return true
		}
		for _, w := range g.Board.HWalls() {
			if w.X == pos.X && w.Y == pos.Y {
				return true
			}
		}
	case Left:
		if pos.X == 0 {
			return true
		}
		for _, w := range g.Board.VWalls() {
			if w.X == pos.X-1 && w.Y == pos.Y {
				return true
			}
		}
	case Right:
		if pos.X == g.Board.Size()-1 {
			return true
		}
		for _, w := range g.Board.VWalls() {
			if w.X == pos.X && w.Y == pos.Y {
				return true
			}
		}
	}
	return false
}

// ValidateMove checks if a bot's intended move is valid based on the game rules.
// The inputs are the board state, the starting positions of all bots, and a bot's intended end position.
func (g *Game) ValidateMove(botId BotId, botEndPos Position) error {
	botPos, ok := g.Bots[botId]
	if !ok {
		return fmt.Errorf("bot with id %d not found", botId)
	}

	if botEndPos == botPos {
		return fmt.Errorf("end position %v is the same as start position", botEndPos)
	}

	// Determine direction from start to end position
	var dir Direction
	switch {
	case botEndPos.X == botPos.X && botEndPos.Y < botPos.Y:
		dir = Up
	case botEndPos.X == botPos.X && botEndPos.Y > botPos.Y:
		dir = Down
	case botEndPos.Y == botPos.Y && botEndPos.X < botPos.X:
		dir = Left
	case botEndPos.Y == botPos.Y && botEndPos.X > botPos.X:
		dir = Right
	default:
		return fmt.Errorf("move from %v to %v is not in a straight line", botPos, botEndPos)
	}

	// Compute where the bot would actually end up
	actualEnd, err := g.ComputeDestination(botId, dir)
	if err != nil {
		return err
	}

	if actualEnd != botEndPos {
		return fmt.Errorf("move to %v is invalid; bot would end at %v", botEndPos, actualEnd)
	}

	return nil
}

// MoveBot returns a new Game with the given bot moved to the given position,
// or an error if the move is invalid.
func (g *Game) MoveBot(id BotId, pos Position) (*Game, error) {
	err := g.ValidateMove(id, pos)
	if err != nil {
		return nil, fmt.Errorf("invalid move for bot %d to position %v: %v", id, pos, err)
	}

	// Create new Bots map with updated position for the moved bot
	newBots := make(map[BotId]Position)
	for botId, botPos := range g.Bots {
		if botId == id {
			newBots[botId] = pos
		} else {
			newBots[botId] = botPos
		}
	}

	return NewGame(g.Board, newBots, g.Target)
}

// IsWin returns true if the target bot is at the target position.
func (g *Game) IsWin() bool {
	targetPos, ok := g.Bots[g.Target.Id]
	if !ok {
		return false
	}
	return targetPos == g.Target.Pos
}

// CheckSolution returns isValid and the resulting game after applying the given moves.
func (g *Game) CheckSolution(moves []BotPosition) (bool, *Game) {
	currentGame := g
	for _, move := range moves {
		var err error
		currentGame, err = currentGame.MoveBot(move.Id, move.Pos)
		if err != nil {
			return false, nil
		}
	}
	if !currentGame.IsWin() {
		return false, nil
	}
	return true, currentGame
}
