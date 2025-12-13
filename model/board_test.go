package model

import "testing"

func TestBoard_String(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	board := NewBoard(3, vW, hW)
	want := dedentBoardString(`
		+----+----+----+
		|              |
		+    +    +    +
		|    |         |
		+    +----+    +
		|              |
		+----+----+----+
		`)
	got := board.String()
	if got != want {
		t.Errorf(`String()
got:
%v

want:
%v`, got, want)
	}
}

func TestPanel_String(t *testing.T) {
	vW := []Position{{0, 1}, {2, 2}}
	hW := []Position{{1, 1}, {2, 2}}
	board := NewPanel(3, vW, hW)
	want := dedentBoardString(`
		+----+----+----+
		|
		+    +    +    +
		|    |
		+    +----+    +
		|              |
		+    +    +----+
		`)
	got := board.String()
	if !boardStringsEqual(got, want) {
		t.Errorf(`String()
got:
%v

want:
%v`, got, want)
	}
}

func TestRotate90cw(t *testing.T) {
	tests := []struct {
		name     string
		startStr string
		wantStr  string
	}{
		{
			// Panels still render top and left borders even though there are no explicit walls there.
			"Case 1",
			`
			+----+----+----+
			|
			+    +    +----+
			|    |         |
			+    +----+    +
			|    |
			+----+    +----+
			`,
			`
			+----+----+----+
			|
			+----+----+    +
			|    |
			+    +    +    +
			|         |
			+    +----+    +
			`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			startPanel, err := ParsePanelString(tc.startStr)
			if err != nil {
				t.Errorf("Unexpected error parsing panel string: %v", err)
			}
			wantPanel, err := ParsePanelString(tc.wantStr)
			if err != nil {
				t.Errorf("Unexpected error parsing panel string: %v", err)
			}
			gotPanel := startPanel.Rotate90cw()
			// Compare string forms because it normalizes wall order.
			if gotPanel.String() != wantPanel.String() {
				t.Errorf("Rotated panel does not match expected panel\nGot:\n%+v\nWant:\n%+v", gotPanel, wantPanel)
			}
		})
	}
}

func TestParsePanelString_PossibleTargets(t *testing.T) {
	panelStr := `
		+----+----+----+
		| []
		+    +    +    +
		|    |
		+    +----+    +
		|           []
		+    +    +    +
		`
	panel, err := ParsePanelString(panelStr)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	targets := panel.PossibleTargets()
	if len(targets) != 2 {
		t.Errorf("Expected 2 possible targets, got %d", len(targets))
	}
	// Check positions (0,0) and (2,2)
	expected := []Position{{0, 0}, {2, 2}}
	for i, exp := range expected {
		if i >= len(targets) || targets[i] != exp {
			t.Errorf("Target %d: expected %v, got %v", i, exp, targets)
		}
	}
}

func TestRotate90cw_PossibleTargets(t *testing.T) {
	// Panel with target at (0,0) - top-left
	panelStr := `
		+----+----+----+
		| []            
		+    +    +    +
		|               
		+    +    +    +
		|               
		+    +    +    +
		`
	panel, err := ParsePanelString(panelStr)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// After 90cw rotation, (0,0) should become (2,0) - top-right
	rotated := panel.Rotate90cw()
	targets := rotated.PossibleTargets()
	if len(targets) != 1 {
		t.Fatalf("Expected 1 target, got %d", len(targets))
	}
	expected := Position{2, 0}
	if targets[0] != expected {
		t.Errorf("Expected target at %v after rotation, got %v", expected, targets[0])
	}
}

func TestBuildBoardFromPanels_PossibleTargets(t *testing.T) {
	// Create a panel with one target
	panelStr := `
		+----+----+
		| []       
		+    +    +
		|          
		+    +    +
		`
	panel, err := ParsePanelString(panelStr)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Build board from 4 identical panels
	board := BuildBoardFromPanels(panel, panel, panel, panel)
	targets := board.PossibleTargets()

	// Each panel contributes 1 target, rotated to different positions
	if len(targets) != 4 {
		t.Errorf("Expected 4 targets from 4 panels, got %d: %v", len(targets), targets)
	}
}
