package bouncebot

import "testing"

func TestRender(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	botStart := []BotPosition{
		{Id: 0, Pos: Position{2, 2}},
		{Id: 1, Pos: Position{0, 0}}}
	goal := BotPosition{0, Position{1, 1}}
	board := Board{Size: 3, VWallPos: vW, HWallPos: hW}
	want := `+----+----+----+
| B1           |
+    +    +    +
|    | T0      |
+    +----+    +
|           B0 |
+----+----+----+`
	got := Render(board, goal, botStart)
	if got != want {
		t.Errorf(`Render(board)
got:
%v

want:
%v`, got, want)
	}
}
