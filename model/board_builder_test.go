package model

import "testing"

func TestAllPanelsAreValid(t *testing.T) {
	for id := 1; id <= 12; id++ {
		t.Run(panelName(id), func(t *testing.T) {
			panel := makePanel(id)

			// Check size is 8x8
			if panel.Size() != 8 {
				t.Errorf("panel %d has size %d, expected 8", id, panel.Size())
			}

			// Check target count is 4 or 5
			targets := panel.PossibleTargets()
			if len(targets) != 4 && len(targets) != 5 {
				t.Errorf("panel %d has %d targets, expected 4 or 5", id, len(targets))
			}

			// Check walls around lower right cell (7,7)
			if !panel.HasVWallAt(Position{X: 6, Y: 7}) {
				t.Errorf("panel %d missing vertical wall at (6,7) for lower-right cell", id)
			}
			if !panel.HasHWallAt(Position{X: 7, Y: 6}) {
				t.Errorf("panel %d missing horizontal wall at (7,6) for lower-right cell", id)
			}
		})
	}
}

func panelName(id int) string {
	names := map[int]string{
		1: "Panel1A", 2: "Panel2A", 3: "Panel3A", 4: "Panel4A",
		5: "Panel1B", 6: "Panel2B", 7: "Panel3B", 8: "Panel4B",
		9: "Panel1C", 10: "Panel2C", 11: "Panel3C", 12: "Panel4C",
	}
	return names[id]
}
