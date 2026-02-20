package solver

import (
	model "github.com/srsalisbury/bouncebot/model"
)

// ReorderSolution reorders the moves in the solution to prioritize grouping moves by bot.
// Uses a greedy approach: tries starting with each bot, then continues with the current bot
// until exhausted before switching. Validates the result and picks the ordering with the
// fewest bot switches.
func ReorderSolution(game *model.Game, solution []model.BotPosition) []model.BotPosition {
	if len(solution) <= 1 {
		return solution
	}

	// Group moves by bot, preserving per-bot order
	movesByBot := make(map[model.BotId][]model.BotPosition)
	for _, move := range solution {
		movesByBot[move.Id] = append(movesByBot[move.Id], move)
	}

	if len(movesByBot) <= 1 {
		return solution // Only one bot, already optimal
	}

	bestResult := solution
	bestGroups := countGroups(solution)
	minPossible := len(movesByBot)

	// Try starting with each bot and greedily build the solution
	for startBot := range movesByBot {
		candidate := buildGreedy(movesByBot, startBot)
		groups := countGroups(candidate)
		if groups < bestGroups {
			if isValid, _ := game.CheckSolution(candidate); isValid {
				bestGroups = groups
				bestResult = candidate
				if bestGroups <= minPossible {
					return bestResult
				}
			}
		}
	}

	return bestResult
}

// buildGreedy constructs a move ordering starting with startBot, then continuing
// with the current bot until its moves are exhausted before switching to the bot
// with the most remaining moves.
func buildGreedy(movesByBot map[model.BotId][]model.BotPosition, startBot model.BotId) []model.BotPosition {
	indices := make(map[model.BotId]int)
	total := 0
	for _, moves := range movesByBot {
		total += len(moves)
	}

	result := make([]model.BotPosition, 0, total)
	currentBot := startBot

	for len(result) < total {
		// Try to continue with current bot
		if idx := indices[currentBot]; idx < len(movesByBot[currentBot]) {
			result = append(result, movesByBot[currentBot][idx])
			indices[currentBot]++
			continue
		}

		// Current bot exhausted, switch to bot with most remaining moves
		bestBot := model.BotId(-1)
		bestRemaining := 0
		for botId, moves := range movesByBot {
			remaining := len(moves) - indices[botId]
			if remaining > 0 && remaining > bestRemaining {
				bestRemaining = remaining
				bestBot = botId
			}
		}

		if bestBot < 0 {
			break
		}

		currentBot = bestBot
		result = append(result, movesByBot[currentBot][indices[currentBot]])
		indices[currentBot]++
	}

	return result
}

// countGroups counts the number of contiguous bot groups in a move sequence.
func countGroups(moves []model.BotPosition) int {
	if len(moves) <= 1 {
		return 1
	}
	groups := 1
	for i := 1; i < len(moves); i++ {
		if moves[i].Id != moves[i-1].Id {
			groups++
		}
	}
	return groups
}
