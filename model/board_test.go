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
	vW := []Position{{0, 1}}
	hW := []Position{{1, 1}}
	board := NewPanel(3, vW, hW)
	want := dedentBoardString(`
		+----+----+----+
		|               
		+    +    +    +
		|    |          
		+    +----+    +
		|               
		+    +    +    +
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
