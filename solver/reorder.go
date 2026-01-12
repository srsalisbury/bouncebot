package solver

import (
	model "github.com/srsalisbury/bouncebot/model"
)

// ReorderSolution reorders the moves in the solution to prioritize grouping moves by bot.
func ReorderSolution(game *model.Game, solution []model.BotPosition) []model.BotPosition {
	if len(solution) <= 1 {
		return solution
	}

	type MoveList = []model.BotPosition

	// Initialize map to hold moves by bot
	movesByBot := make(map[model.BotId]MoveList)
	currentIndices := make(map[model.BotId]int)

	// Group moves by bot
	for _, move := range solution {
		movesByBot[move.Id] = append(movesByBot[move.Id], move)
	}

	// Generate reordered moves
	var interleavings []MoveList
	var current = make(MoveList, 0)

	var interleave func()
	interleave = func() {
		// Check if all moves are exhausted
		if len(current) == len(solution) {
			interleavings = append(interleavings, append(MoveList{}, current...))
			return
		}
		for botId := range movesByBot {
			var idx = currentIndices[botId]
			if idx < len(movesByBot[botId]) {
				// Choose move
				current = append(current, movesByBot[botId][idx])
				currentIndices[botId] = idx + 1

				// Recurse
				interleave()

				// Backtrack
				current = current[:len(current)-1]
				currentIndices[botId] = idx
			}
		}
	}

	interleave()
	// Select the interleaving with the least number of bot switches
	countGroups := func(moves MoveList) int {
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

	minPossibleGroups := len(movesByBot)
	bestInterleaving := solution
	bestGroupCount := countGroups(bestInterleaving)
	for _, il := range interleavings {
		if bestGroupCount == minPossibleGroups {
			break
		}
		gc := countGroups(il)
		if gc < bestGroupCount {
			isValid, _ := game.CheckSolution(il)
			if isValid {
				bestGroupCount = gc
				bestInterleaving = il
			}
		}
	}
	return bestInterleaving
}
