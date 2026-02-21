package solver

import (
	"testing"
	"time"

	"github.com/srsalisbury/bouncebot/model"
)

// Helper to count bot switches in a solution
func countSwitches(solution []model.BotPosition) int {
	if len(solution) <= 1 {
		return 0
	}
	switches := 0
	for i := 1; i < len(solution); i++ {
		if solution[i].Id != solution[i-1].Id {
			switches++
		}
	}
	return switches
}

// Helper to extract per-bot move sequences (to verify ordering preserved)
func extractBotSequences(solution []model.BotPosition) map[model.BotId][]model.Position {
	result := make(map[model.BotId][]model.Position)
	for _, m := range solution {
		result[m.Id] = append(result[m.Id], m.Pos)
	}
	return result
}

// Helper to check if two position slices are equal
func positionsEqual(a, b []model.Position) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Tests for validity checking - reorder should only return valid solutions

func TestReorderSolution_ReturnsValidSolution(t *testing.T) {
	// Use Game1 which has a known valid solution
	game := model.Game1()
	solution := model.Game1OptimalSolution()

	result := ReorderSolution(game, solution)

	// Result must be a valid solution
	isValid, _ := game.CheckSolution(result)
	if !isValid {
		t.Errorf("ReorderSolution returned an invalid solution: %v", result)
	}
}

func TestReorderSolution_KeepsValidOrderWhenReorderingWouldInvalidate(t *testing.T) {
	// Game1OptimalSolution: Bot 1 must move first to clear the path for Bot 0
	// Original: [Bot1, Bot0, Bot0, Bot0, Bot0, Bot0, Bot0] - 1 switch
	// If we tried [Bot0, Bot0, Bot0, Bot0, Bot0, Bot0, Bot1] - still 1 switch but likely invalid
	game := model.Game1()
	solution := model.Game1OptimalSolution()

	// Verify original solution is valid
	isValid, _ := game.CheckSolution(solution)
	if !isValid {
		t.Fatalf("Original Game1OptimalSolution should be valid")
	}

	result := ReorderSolution(game, solution)

	// Result must still be valid
	isValid, _ = game.CheckSolution(result)
	if !isValid {
		t.Errorf("ReorderSolution returned an invalid solution")
	}

	// Verify per-bot ordering is preserved
	originalSequences := extractBotSequences(solution)
	resultSequences := extractBotSequences(result)
	for botId, originalMoves := range originalSequences {
		if !positionsEqual(originalMoves, resultSequences[botId]) {
			t.Errorf("Bot %d ordering not preserved. Original: %v, Result: %v",
				botId, originalMoves, resultSequences[botId])
		}
	}
}

func TestReorderSolution_PrefersFewerSwitchesWhenBothValid(t *testing.T) {
	// Create a scenario where multiple orderings are valid
	// but we want the one with fewer switches
	game := model.Game1()
	solution := model.Game1OptimalSolution()

	result := ReorderSolution(game, solution)

	// Count switches in result
	switches := countSwitches(result)

	// The result should have minimal switches among valid solutions
	// For Game1OptimalSolution with 1 Bot1 move and 9 Bot0 moves, optimal is 2 switches
	if switches > 2 {
		t.Errorf("Expected at most 2 switches for Game1OptimalSolution, got %d", switches)
	}
}

func TestReorderSolution_ValidityCheck_OrderDependent(t *testing.T) {
	// This test verifies that when move order matters for validity,
	// the reorder function respects that constraint
	game := model.Game1()
	solution := model.Game1OptimalSolution()

	// Try reversing the solution (putting all Bot0 moves before Bot1)
	// This should be invalid because Bot1 needs to move first
	reversed := make([]model.BotPosition, len(solution))
	bot0Moves := []model.BotPosition{}
	bot1Moves := []model.BotPosition{}
	for _, m := range solution {
		if m.Id == 0 {
			bot0Moves = append(bot0Moves, m)
		} else {
			bot1Moves = append(bot1Moves, m)
		}
	}
	// Put Bot0 moves first, then Bot1
	copy(reversed, bot0Moves)
	copy(reversed[len(bot0Moves):], bot1Moves)

	// Verify the reversed order is actually invalid
	isReversedValid, _ := game.CheckSolution(reversed)

	// Now run ReorderSolution - it should NOT return the invalid ordering
	// even if that ordering has fewer switches
	result := ReorderSolution(game, solution)
	isResultValid, _ := game.CheckSolution(result)

	if !isResultValid {
		t.Errorf("ReorderSolution returned invalid solution: %v", result)
	}

	// If reversed was invalid, make sure we didn't return something equivalent
	if !isReversedValid {
		// Check that result doesn't start with all Bot0 moves
		bot0Count := 0
		for i, m := range result {
			if m.Id == 0 {
				bot0Count++
			} else {
				// Found a non-Bot0 move
				if bot0Count == len(bot0Moves) {
					t.Errorf("Result has all Bot0 moves first (invalid order): %v", result)
				}
				break
			}
			if i == len(result)-1 && bot0Count == len(bot0Moves) {
				// All Bot0 moves came first (only possible if no other bots)
				// This check handles edge case
			}
		}
	}
}

func TestReorderSolution_PreservesTotalMoves(t *testing.T) {
	game := model.Game1()
	solution := model.Game1OptimalSolution()

	result := ReorderSolution(game, solution)

	if len(result) != len(solution) {
		t.Errorf("Expected %d moves, got %d", len(solution), len(result))
	}
}

func TestReorderSolution_ContainsSameMoves(t *testing.T) {
	game := model.Game1()
	solution := model.Game1OptimalSolution()

	result := ReorderSolution(game, solution)

	// Count occurrences of each move in input
	inputCounts := make(map[model.BotPosition]int)
	for _, m := range solution {
		inputCounts[m]++
	}

	// Count occurrences in result
	resultCounts := make(map[model.BotPosition]int)
	for _, m := range result {
		resultCounts[m]++
	}

	// Compare
	for m, count := range inputCounts {
		if resultCounts[m] != count {
			t.Errorf("Move %v appears %d times in input but %d times in result", m, count, resultCounts[m])
		}
	}
}

func TestReorderSolution_EmptyInput(t *testing.T) {
	game := model.Game1()

	result := ReorderSolution(game, nil)
	if result != nil {
		t.Errorf("Expected nil for nil input, got %v", result)
	}

	result = ReorderSolution(game, []model.BotPosition{})
	if len(result) != 0 {
		t.Errorf("Expected empty for empty input, got %v", result)
	}
}

func TestReorderSolution_SingleBot(t *testing.T) {
	game := model.Game1()

	// Solution with only bot 0 moves (first 7 moves of Game1OptimalSolution)
	solution := model.Game1OptimalSolution()[:7]
	result := ReorderSolution(game, solution)

	// Single bot means no reordering needed, result should equal input
	if len(result) != len(solution) {
		t.Fatalf("Expected %d moves, got %d", len(solution), len(result))
	}
	for i := range result {
		if result[i] != solution[i] {
			t.Errorf("Move %d differs: expected %v, got %v", i, solution[i], result[i])
		}
	}
}

func TestReorderSolution_FindsBetterThanGreedy(t *testing.T) {
	// Construct a scenario with circular dependencies where greedy fails but DFS
	// finds a better ordering.
	//
	// 8x8 board, no internal walls. 3 bots forming a blocking cycle:
	//   Bot 0 (A) at (0,0), Bot 1 (B) at (0,4), Bot 2 (C) at (4,4)
	//
	// Solution moves (6 moves, each bot has 2):
	//   A1: A moves down, blocked by B -> (0,3)
	//   B1: B moves right, blocked by C -> (3,4)
	//   C1: C moves up -> (4,0)
	//   A2: A moves down (B gone) -> (0,7)
	//   B2: B moves right (C gone) -> (7,4)
	//   C2: C moves left (A gone) -> (0,0)
	//
	// Dependencies form a cycle:
	//   A1 must precede B1 (A1 needs B at (0,4) to block)
	//   B1 must precede C1 (B1 needs C at (4,4) to block)
	//   A2 needs B1 done, B2 needs C1 done, C2 needs A1 done
	//
	// Greedy: all starting-bot candidates try to exhaust one bot first, which
	// violates the circular deps. All candidates fail validation, so greedy
	// returns the original 6-group interleaving.
	//
	// DFS finds: [A1, B1, C1, C2, A2, B2] = 5 groups (groups C1+C2 together).

	board := model.NewBoard(8, nil, nil)
	bots := map[model.BotId]model.Position{
		0: {X: 0, Y: 0},
		1: {X: 0, Y: 4},
		2: {X: 4, Y: 4},
	}
	target := model.BotPosition{Id: 0, Pos: model.Position{X: 0, Y: 7}}
	game, err := model.NewGame(board, bots, target)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Original solution: maximally interleaved (A, B, C, A, B, C) = 6 groups
	solution := []model.BotPosition{
		{Id: 0, Pos: model.Position{X: 0, Y: 3}}, // A1: down, blocked by B
		{Id: 1, Pos: model.Position{X: 3, Y: 4}}, // B1: right, blocked by C
		{Id: 2, Pos: model.Position{X: 4, Y: 0}}, // C1: up to edge
		{Id: 0, Pos: model.Position{X: 0, Y: 7}}, // A2: down, B moved
		{Id: 1, Pos: model.Position{X: 7, Y: 4}}, // B2: right, C moved
		{Id: 2, Pos: model.Position{X: 0, Y: 0}}, // C2: left, A moved
	}

	// Verify the original solution is valid
	isValid, _ := game.CheckSolution(solution)
	if !isValid {
		t.Fatalf("Original solution should be valid")
	}

	inputGroups := countGroups(solution)
	if inputGroups != 6 {
		t.Fatalf("Expected 6 groups in input, got %d", inputGroups)
	}

	result := ReorderSolution(game, solution)

	// DFS should find a reordering with fewer groups than the original
	resultGroups := countGroups(result)
	if resultGroups >= inputGroups {
		t.Errorf("DFS should find fewer groups than input (%d), got %d. Result: %v",
			inputGroups, resultGroups, result)
	}

	// Verify the result is a valid solution
	isValid, _ = game.CheckSolution(result)
	if !isValid {
		t.Errorf("Reordered solution is invalid: %v", result)
	}

	// Verify per-bot ordering is preserved
	origSeq := extractBotSequences(solution)
	resultSeq := extractBotSequences(result)
	for botId, origMoves := range origSeq {
		if !positionsEqual(origMoves, resultSeq[botId]) {
			t.Errorf("Bot %d ordering not preserved. Original: %v, Result: %v",
				botId, origMoves, resultSeq[botId])
		}
	}

	t.Logf("Input: %d groups, Result: %d groups, Ordering: %v", inputGroups, resultGroups, result)
}

func TestReorderSolution_LargeInput_CompletesQuickly(t *testing.T) {
	// Build a large synthetic solution with 5 bots and many interleaved moves.
	// The old exhaustive approach would take factorial time on this input;
	// the greedy approach should complete in well under a second.
	game := model.Game1()

	// Create a heavily interleaved sequence: 5 bots, 4 moves each = 20 moves
	// (The multinomial coefficient for 5 groups of 4 is 20!/(4!^5) = 11,732,745,024)
	bots := []model.BotId{0, 1, 2, 3}
	movesPerBot := 5
	var solution []model.BotPosition
	for i := 0; i < movesPerBot; i++ {
		for _, botId := range bots {
			solution = append(solution, model.BotPosition{
				Id:  botId,
				Pos: model.Position{X: model.BoardDim(i + 1), Y: model.BoardDim(int(botId) + 1)},
			})
		}
	}

	start := time.Now()
	result := ReorderSolution(game, solution)
	elapsed := time.Since(start)

	// The greedy approach should complete nearly instantly
	if elapsed > 1*time.Second {
		t.Errorf("ReorderSolution took %v, expected < 1s for greedy approach", elapsed)
	}

	// Result should have the same total moves
	if len(result) != len(solution) {
		t.Errorf("Expected %d moves, got %d", len(solution), len(result))
	}

	// Result should have fewer or equal groups compared to the interleaved input
	inputGroups := countGroups(solution)
	resultGroups := countGroups(result)
	if resultGroups > inputGroups {
		t.Errorf("Result has more groups (%d) than input (%d)", resultGroups, inputGroups)
	}

	t.Logf("Input: %d groups, Result: %d groups, Time: %v", inputGroups, resultGroups, elapsed)
}
