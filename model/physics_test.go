package model

import (
	"encoding/json"
	"os"
	"testing"
)

// PhysicsCase represents a single test case from the shared fixtures.
type PhysicsCase struct {
	Name      string `json:"name"`
	BoardSize int    `json:"boardSize"`
	VWalls    []struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"vWalls"`
	HWalls []struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"hWalls"`
	Robots []struct {
		ID int `json:"id"`
		X  int `json:"x"`
		Y  int `json:"y"`
	} `json:"robots"`
	Move struct {
		RobotID   int    `json:"robotId"`
		Direction string `json:"direction"`
	} `json:"move"`
	Expected struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"expected"`
}

type PhysicsCases struct {
	Cases []PhysicsCase `json:"cases"`
}

func TestPhysicsMatchesClient(t *testing.T) {
	data, err := os.ReadFile("../tests/physics_cases.json")
	if err != nil {
		t.Fatalf("failed to read test cases: %v", err)
	}

	var cases PhysicsCases
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("failed to parse test cases: %v", err)
	}

	for _, tc := range cases.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Build walls
			vWalls := make([]Position, len(tc.VWalls))
			for i, w := range tc.VWalls {
				vWalls[i] = Position{X: BoardDim(w.X), Y: BoardDim(w.Y)}
			}
			hWalls := make([]Position, len(tc.HWalls))
			for i, w := range tc.HWalls {
				hWalls[i] = Position{X: BoardDim(w.X), Y: BoardDim(w.Y)}
			}
			board := NewBoard(BoardDim(tc.BoardSize), vWalls, hWalls)

			// Build bots map
			bots := make(map[BotId]Position)
			for _, r := range tc.Robots {
				bots[BotId(r.ID)] = Position{X: BoardDim(r.X), Y: BoardDim(r.Y)}
			}

			// Create game (target doesn't matter for physics)
			game := &Game{
				Board:  board,
				Bots:   bots,
				Target: BotPosition{Id: 0, Pos: Position{X: 0, Y: 0}},
			}

			// Compute destination
			result, err := game.ComputeDestination(BotId(tc.Move.RobotID), Direction(tc.Move.Direction))
			if err != nil {
				t.Fatalf("ComputeDestination error: %v", err)
			}

			if result.X != BoardDim(tc.Expected.X) || result.Y != BoardDim(tc.Expected.Y) {
				t.Errorf("got (%d,%d), want (%d,%d)", result.X, result.Y, tc.Expected.X, tc.Expected.Y)
			}
		})
	}
}
