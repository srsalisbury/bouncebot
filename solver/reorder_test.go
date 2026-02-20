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
