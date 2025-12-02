package bouncebot

import (
	"strings"
	"testing"

	"github.com/lithammer/dedent"
)

// Dedent and remove leading/trailing blank lines for easier comparison in tests.
func dedentTestString(s string) string {
	return strings.TrimSpace(dedent.Dedent(s))
}

func TestRender(t *testing.T) {
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	botsStart := BotPositionMap{
		0: Position{2, 2},
		1: Position{0, 0}}
	goal := BotPosition{0, Position{1, 1}}
	board := &Board{Size: 3, VWallPos: vW, HWallPos: hW}
	want := dedentTestString(`
		+----+----+----+
		| B1           |
		+    +    +    +
		|    | T0      |
		+    +----+    +
		|           B0 |
		+----+----+----+
		`)
	got := Render(board, goal, botsStart)
	if got != want {
		t.Errorf(`Render(board)
got:
%v

want:
%v`, got, want)
	}
}
