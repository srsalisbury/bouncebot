package bouncebot

import "testing"

/*
Starting Board:
+ -- + -- + -- +
|              |
+    +    +    +
| B2 |         |
+    + -- +    +
| B1        B0 |
+ -- + -- + -- +
*/

func TestValidate(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	botStart := []BotPosition{
		{Id: 0, Pos: Position{2, 2}},
		{Id: 1, Pos: Position{0, 2}},
		{Id: 2, Pos: Position{0, 1}}}
	board := Board{Size: 3, VWallPos: vW, HWallPos: hW}

	var tests = []struct {
		name   string
		botEnd BotPosition // Bot's intended end position
		want   bool
	}{
		{"Valid - bot 0 upwards to border", BotPosition{0, Position{2, 0}}, true},
		{"Valid - left to bot 1", BotPosition{0, Position{1, 2}}, true},
		{"Valid - bot 2 upwards to border", BotPosition{2, Position{0, 0}}, true},
		{"Valid - right to bot 0", BotPosition{1, Position{1, 2}}, true},
		{"Invalid - same position", BotPosition{0, Position{2, 2}}, false},
		{"Invalid - out of bounds", BotPosition{0, Position{-1, 2}}, false},
		{"Invalid - through bot 1", BotPosition{0, Position{0, 2}}, false},
		{"Invalid - diagonal move", BotPosition{0, Position{0, 1}}, false},
		{"Invalid - through wall", BotPosition{2, Position{1, 1}}, false},
		{"Invalid - through bot 2", BotPosition{1, Position{0, 0}}, false},
		{"Invalid - not against wall/border/bot", BotPosition{0, Position{2, 1}}, false},
	}

	for _, tc := range tests {
		got := ValidateMove(board, botStart, tc.botEnd)
		if got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
