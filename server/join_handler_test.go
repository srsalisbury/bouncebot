package main

import (
	"bytes"
	"image/png"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/srsalisbury/bouncebot/server/room"
)

func TestHandleJoinPage_ExistingRoom(t *testing.T) {
	svc := room.NewRoomService()
	createdRoom, _, _ := svc.Create("Mike", false)

	handler := handleJoinPage(svc, "https://client.example.com")
	req := httptest.NewRequest("GET", "https://api.example.com/join/"+createdRoom.ID, nil)
	req.SetPathValue("roomId", createdRoom.ID)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()

	roomURL := "https://client.example.com/room/" + createdRoom.ID
	if !strings.Contains(body, `content="0;url=`+roomURL+`"`) {
		t.Errorf("expected meta-refresh redirect to %s, got body: %s", roomURL, body)
	}
	if !strings.Contains(body, `location.replace("`+roomURL+`")`) {
		t.Errorf("expected JS redirect to %s, got body: %s", roomURL, body)
	}

	imageURL := "https://api.example.com/join/" + createdRoom.ID + "/preview.png"
	if !strings.Contains(body, `property="og:image" content="`+imageURL+`"`) {
		t.Errorf("expected og:image pointing at %s, got body: %s", imageURL, body)
	}
	if !strings.Contains(body, `property="og:title"`) {
		t.Error("expected an og:title tag")
	}
	if !strings.Contains(body, `property="og:description"`) {
		t.Error("expected an og:description tag")
	}
	if !strings.Contains(body, `name="description"`) {
		t.Error("expected a fallback <meta name=\"description\"> tag")
	}
	if !strings.Contains(body, createdRoom.ID) {
		t.Error("expected the room ID to appear in the page somewhere")
	}
}

func TestHandleJoinPage_NonexistentRoom(t *testing.T) {
	svc := room.NewRoomService()
	handler := handleJoinPage(svc, "https://client.example.com")

	req := httptest.NewRequest("GET", "https://api.example.com/join/NOPE", nil)
	req.SetPathValue("roomId", "NOPE")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != 404 {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleJoinPreviewImage_ExistingRoom(t *testing.T) {
	svc := room.NewRoomService()
	createdRoom, _, _ := svc.Create("Mike", false)

	handler := handleJoinPreviewImage(svc)
	req := httptest.NewRequest("GET", "https://api.example.com/join/"+createdRoom.ID+"/preview.png", nil)
	req.SetPathValue("roomId", createdRoom.ID)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("expected Content-Type image/png, got %q", ct)
	}
	if _, err := png.Decode(bytes.NewReader(rec.Body.Bytes())); err != nil {
		t.Errorf("response body is not a valid PNG: %v", err)
	}
}

func TestHandleJoinPreviewImage_NonexistentRoom(t *testing.T) {
	svc := room.NewRoomService()
	handler := handleJoinPreviewImage(svc)

	req := httptest.NewRequest("GET", "https://api.example.com/join/NOPE/preview.png", nil)
	req.SetPathValue("roomId", "NOPE")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != 404 {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
