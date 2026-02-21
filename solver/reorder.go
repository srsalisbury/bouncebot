package solver

import (
	"sort"

	model "github.com/srsalisbury/bouncebot/model"
)

// dfsState holds shared state for the DFS reorder search.
type dfsState struct {
	game       *model.Game
	movesByBot map[model.BotId][]model.BotPosition
	indices    map[model.BotId]int
	botIds     []model.BotId // sorted for deterministic iteration
	totalMoves int
	result     []model.BotPosition // current path
	bestResult []model.BotPosition
	bestGroups int
	nodeCount  int
	nodeLimit  int
}

// ReorderSolution reorders the moves in the solution to prioritize grouping moves by bot.
// Uses DFS with branch-and-bound to find the ordering with the fewest bot switches
// among all valid interleavings that preserve per-bot move order.
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

	// Run greedy to get initial bound
	bestResult := solution
	bestGroups := countGroups(solution)

	for startBot := range movesByBot {
		candidate := buildGreedy(movesByBot, startBot)
		groups := countGroups(candidate)
		if groups < bestGroups {
			if isValid, _ := game.CheckSolution(candidate); isValid {
				bestGroups = groups
				bestResult = candidate
			}
		}
	}

	// If greedy already achieved the minimum possible (one group per bot), return
	minPossible := len(movesByBot)
	if bestGroups <= minPossible {
		return bestResult
	}

	// Collect and sort bot IDs for deterministic iteration
	botIds := make([]model.BotId, 0, len(movesByBot))
	for botId := range movesByBot {
		botIds = append(botIds, botId)
	}
	sort.Slice(botIds, func(i, j int) bool { return botIds[i] < botIds[j] })

	// Save original bot positions, then restore after DFS
	originalBots := make(map[model.BotId]model.Position, len(game.Bots))
	for id, pos := range game.Bots {
		originalBots[id] = pos
	}
	defer func() {
		for id, pos := range originalBots {
			game.Bots[id] = pos
		}
	}()

	// Initialize DFS state
	state := &dfsState{
		game:       game,
		movesByBot: movesByBot,
		indices:    make(map[model.BotId]int, len(movesByBot)),
		botIds:     botIds,
		totalMoves: len(solution),
		result:     make([]model.BotPosition, len(solution)),
		bestResult: bestResult,
		bestGroups: bestGroups,
		nodeCount:  0,
		nodeLimit:  100_000,
	}

	// Try starting with each bot
	for _, startBot := range botIds {
		if state.bestGroups <= minPossible {
			break
		}
		move := movesByBot[startBot][0]
		if game.ValidateMove(move.Id, move.Pos) != nil {
			continue
		}

		// Apply first move
		oldPos := game.Bots[move.Id]
		game.Bots[move.Id] = move.Pos
		state.indices[startBot] = 1
		state.result[0] = move

		dfsReorder(state, 1, startBot, 1) // 1 group for the starting bot

		// Restore
		game.Bots[move.Id] = oldPos
		state.indices[startBot] = 0
	}

	return state.bestResult
}

// dfsReorder recursively explores valid move interleavings.
func dfsReorder(s *dfsState, depth int, currentBot model.BotId, currentGroups int) {
	if depth == s.totalMoves {
		if currentGroups < s.bestGroups {
			s.bestGroups = currentGroups
			s.bestResult = make([]model.BotPosition, s.totalMoves)
			copy(s.bestResult, s.result)
		}
		return
	}

	s.nodeCount++
	if s.nodeCount > s.nodeLimit {
		return
	}

	// Pruning: count remaining distinct bots (excluding currentBot) that still have moves.
	// Each needs at least one new group.
	remainingOtherBots := 0
	for _, botId := range s.botIds {
		if botId != currentBot && s.indices[botId] < len(s.movesByBot[botId]) {
			remainingOtherBots++
		}
	}
	if currentGroups+remainingOtherBots >= s.bestGroups {
		return
	}

	// Try current bot first (no group increment), then others in sorted order.
	// This finds good solutions early, tightening the bound quickly.
	tryBot := func(botId model.BotId) {
		if s.nodeCount > s.nodeLimit {
			return
		}

		idx := s.indices[botId]
		if idx >= len(s.movesByBot[botId]) {
			return
		}

		move := s.movesByBot[botId][idx]

		if s.game.ValidateMove(move.Id, move.Pos) != nil {
			return
		}

		// Apply move in-place
		oldPos := s.game.Bots[move.Id]
		s.game.Bots[move.Id] = move.Pos
		s.indices[botId] = idx + 1
		s.result[depth] = move

		newGroups := currentGroups
		if botId != currentBot {
			newGroups++
		}

		dfsReorder(s, depth+1, botId, newGroups)

		// Restore
		s.game.Bots[move.Id] = oldPos
		s.indices[botId] = idx
	}

	// Current bot first
	tryBot(currentBot)

	// Then all other bots
	for _, botId := range s.botIds {
		if botId == currentBot {
			continue
		}
		tryBot(botId)
	}
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
