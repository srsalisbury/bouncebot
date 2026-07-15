package model

import (
	"encoding/base64"
	"testing"
)

// game1ShareCode is a fixed vector: the exact expected encoding of Game1().
// The TypeScript implementation has a test asserting this same string, to
// guarantee the two implementations agree byte-for-byte.
const game1ShareCode = "ARAZEDESYyZn4ragpIGHv57cq8p6iB4ZTUhfaBlAEWMFNnb04cWkkYad7KvZifiICx4oTFh4EUESYzbixqSRnuyr2ooeKU1YAE1UrDnEkw"

func TestEncodeShareCode_FixedVector(t *testing.T) {
	code, err := EncodeShareCode(Game1())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != game1ShareCode {
		t.Errorf("EncodeShareCode(Game1()) = %q, want %q (this must match the TypeScript implementation's fixed vector - if this changed intentionally, update both)", code, game1ShareCode)
	}
}

func TestDecodeShareCode_FixedVector(t *testing.T) {
	game, err := DecodeShareCode(game1ShareCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !game.Equals(Game1()) {
		t.Errorf("DecodeShareCode(game1ShareCode) did not reproduce Game1():\ngot:\n%s\nwant:\n%s", game.String(), Game1().String())
	}
}

func TestShareCode_RoundTrip(t *testing.T) {
	games := []*Game{Game1()}
	for i := 0; i < 10; i++ {
		games = append(games, NewRandomGame())
	}

	for i, original := range games {
		code, err := EncodeShareCode(original)
		if err != nil {
			t.Fatalf("game %d: EncodeShareCode failed: %v", i, err)
		}
		decoded, err := DecodeShareCode(code)
		if err != nil {
			t.Fatalf("game %d: DecodeShareCode failed: %v", i, err)
		}
		if !decoded.Equals(original) {
			t.Errorf("game %d: round-trip mismatch:\noriginal:\n%s\ndecoded:\n%s", i, original.String(), decoded.String())
		}
	}
}

func TestShareCode_ContinuationAfterDecode(t *testing.T) {
	// Regression guard for the exact bug fixed earlier this session
	// (NewContinuationGame panicking on a board with no possible targets) -
	// a decoded shared board must be just as usable for a next round as any
	// other game.
	code, err := EncodeShareCode(NewRandomGame())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	decoded, err := DecodeShareCode(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decoded.Board.PossibleTargets()) == 0 {
		t.Fatal("decoded board has no possible targets - continuation would silently fall back to a random game")
	}
	// Must not panic.
	NewContinuationGame(decoded)
}

func TestDecodeShareCode_BadChecksum(t *testing.T) {
	code, err := EncodeShareCode(Game1())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Flip the last character (the checksum byte's encoding) to corrupt it.
	last := code[len(code)-1]
	replacement := byte('A')
	if last == 'A' {
		replacement = 'B'
	}
	corrupted := code[:len(code)-1] + string(replacement)

	if _, err := DecodeShareCode(corrupted); err == nil {
		t.Error("expected an error for a corrupted checksum, got nil")
	}
}

func TestDecodeShareCode_Truncated(t *testing.T) {
	code, err := EncodeShareCode(Game1())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	truncated := code[:len(code)/2]

	if _, err := DecodeShareCode(truncated); err == nil {
		t.Error("expected an error for a truncated code, got nil")
	}
}

func TestDecodeShareCode_InvalidBase64(t *testing.T) {
	if _, err := DecodeShareCode("not valid base64url!!!"); err == nil {
		t.Error("expected an error for invalid base64, got nil")
	}
}

func TestDecodeShareCode_EmptyString(t *testing.T) {
	if _, err := DecodeShareCode(""); err == nil {
		t.Error("expected an error for an empty code, got nil")
	}
}

func TestDecodeShareCode_UnknownVersion(t *testing.T) {
	// Build a minimal valid-shape payload but with a bogus version byte, and
	// a correct checksum so only the version check can reject it.
	payload := []byte{99, 16, 0, 0, 0, 0, 0, 0, 0, 0}
	payload = append(payload, checksum(payload))
	code := base64.RawURLEncoding.EncodeToString(payload)

	if _, err := DecodeShareCode(code); err == nil {
		t.Error("expected an error for an unknown format version, got nil")
	}
}

func TestDecodeShareCode_TargetBotIdOutOfRange(t *testing.T) {
	// version=1, size=16, 0 vWalls, 0 hWalls, 0 targets, targetBotId=7 (invalid)
	payload := []byte{1, 16, 0, 0, 0, 7}
	payload = append(payload, checksum(payload))
	code := base64.RawURLEncoding.EncodeToString(payload)

	if _, err := DecodeShareCode(code); err == nil {
		t.Error("expected an error for an out-of-range target bot id, got nil")
	}
}

func TestEncodeShareCode_RejectsOutOfRangePosition(t *testing.T) {
	board := NewBoardWithTargets(20, nil, nil, []Position{{X: 19, Y: 19}})
	bots := map[BotId]Position{0: {X: 0, Y: 0}, 1: {X: 1, Y: 0}, 2: {X: 2, Y: 0}, 3: {X: 3, Y: 0}}
	game := &Game{Board: board, Bots: bots, Target: BotPosition{Id: 0, Pos: Position{X: 19, Y: 19}}}

	if _, err := EncodeShareCode(game); err == nil {
		t.Error("expected an error encoding a board with coordinates >15, got nil")
	}
}
