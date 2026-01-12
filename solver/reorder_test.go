package solver

import (
	"testing"

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
	solution := model.Game1Solution()

	result := ReorderSolution(game, solution)

	// Result must be a valid solution
	isValid, _ := game.CheckSolution(result)
	if !isValid {
		t.Errorf("ReorderSolution returned an invalid solution: %v", result)
	}
}

func TestReorderSolution_KeepsValidOrderWhenReorderingWouldInvalidate(t *testing.T) {
	// Game1Solution: Bot 1 must move first to clear the path for Bot 0
	// Original: [Bot1, Bot0, Bot0, Bot0, Bot0, Bot0, Bot0] - 1 switch
	// If we tried [Bot0, Bot0, Bot0, Bot0, Bot0, Bot0, Bot1] - still 1 switch but likely invalid
	game := model.Game1()
	solution := model.Game1Solution()

	// Verify original solution is valid
	isValid, _ := game.CheckSolution(solution)
	if !isValid {
		t.Fatalf("Original Game1Solution should be valid")
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
	solution := model.Game1Solution()

	result := ReorderSolution(game, solution)

	// Count switches in result
	switches := countSwitches(result)

	// The result should have minimal switches among valid solutions
	// For Game1Solution with 1 Bot1 move and 6 Bot0 moves, optimal is 1 switch
	if switches > 1 {
		t.Errorf("Expected at most 1 switch for Game1Solution, got %d", switches)
	}
}

func TestReorderSolution_ValidityCheck_OrderDependent(t *testing.T) {
	// This test verifies that when move order matters for validity,
	// the reorder function respects that constraint
	game := model.Game1()
	solution := model.Game1Solution()

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
	solution := model.Game1Solution()

	result := ReorderSolution(game, solution)

	if len(result) != len(solution) {
		t.Errorf("Expected %d moves, got %d", len(solution), len(result))
	}
}

func TestReorderSolution_ContainsSameMoves(t *testing.T) {
	game := model.Game1()
	solution := model.Game1Solution()

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
