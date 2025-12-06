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
	if got != want {
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
