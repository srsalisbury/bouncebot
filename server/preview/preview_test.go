package preview

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRender_ValidPNG(t *testing.T) {
	data, err := Render("ABCD")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("output is not a valid PNG: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != width || bounds.Dy() != height {
		t.Errorf("expected %dx%d image, got %dx%d", width, height, bounds.Dx(), bounds.Dy())
	}
}

func TestRender_DiffersByRoomID(t *testing.T) {
	a, err := Render("AAAA")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	b, err := Render("ZZZZ")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if bytes.Equal(a, b) {
		t.Error("expected different room IDs to produce different images")
	}
}

func TestRender_DoesNotMutateBaseTemplate(t *testing.T) {
	if _, err := Render("FIRST"); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	first := make([]byte, len(baseTemplate.Pix))
	copy(first, baseTemplate.Pix)

	if _, err := Render("SECOND"); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if !bytes.Equal(first, baseTemplate.Pix) {
		t.Error("Render mutated the shared base template - concurrent calls would corrupt each other's output")
	}
}
