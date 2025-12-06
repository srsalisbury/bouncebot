package model

import "testing"

func TestBuildBoardFromPanels(t *testing.T) {
	tests := []struct {
		name     string
		panelStr string
		wantStr  string
	}{
		{
			"Case 1",
			`
			+----+----+----+
			|    |         |
			+    +    +    +
			|              |
			+    +----+    +
			|              |
			+----+----+----+
			`,
			`
			+----+----+----+----+----+----+
			|    |                        |
			+    +    +    +    +    +----+
			|                   |         |
			+    +----+    +    +    +    +
			|                             |
			+    +    +    +    +    +    +
			|                             |
			+    +    +    +    +----+    +
			|         |                   |
			+----+    +    +    +    +    +
			|                        |    |
			+----+----+----+----+----+----+
			`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			panel, err := ParseBoardString(tc.panelStr)
			if err != nil {
				t.Errorf("Unexpected error parsing panel string: %v", err)
			}
			wantBoard, err := ParseBoardString(tc.wantStr)
			if err != nil {
				t.Errorf("Unexpected error parsing board string: %v", err)
			}
			gotBoard := BuildBoardFromPanels(panel, panel, panel, panel)
			// Compare string forms because it normalizes wall order.
			if gotBoard.String() != wantBoard.String() {
				t.Errorf("BoardFromPanels()\nGot:\n%+v\nWant:\n%+v", gotBoard, wantBoard)
			}
		})
	}
}
