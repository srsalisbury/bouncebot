package main

import (
	"fmt"
	"html"
	"net/http"

	"github.com/srsalisbury/bouncebot/server/preview"
	"github.com/srsalisbury/bouncebot/server/room"
)

// requestOrigin reconstructs this server's own externally-visible origin
// from the incoming request, so the og:image URL is absolute (required by
// most link-preview crawlers) without needing a separate config value for
// something the request already tells us.
func requestOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + r.Host
}

const joinPageTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>%[1]s</title>
<meta property="og:title" content="%[1]s">
<meta property="og:description" content="%[2]s">
<meta property="og:image" content="%[3]s">
<meta property="og:image:width" content="1200">
<meta property="og:image:height" content="630">
<meta name="description" content="%[2]s">
<meta http-equiv="refresh" content="0;url=%[4]s">
<script>location.replace(%[5]q);</script>
</head>
<body>
Redirecting to BounceBot&hellip; <a href="%[4]s">click here if you're not redirected</a>.
</body>
</html>
`

const joinPageNotFoundTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Room not found - BounceBot</title>
</head>
<body>
This BounceBot room doesn't exist anymore. <a href="%[1]s">Go to BounceBot</a>.
</body>
</html>
`

// handleJoinPage serves a small server-rendered HTML page at /join/{roomId}
// with real Open Graph tags, so link-preview crawlers (which don't run
// JavaScript) see a proper preview. Real browsers land here for an instant
// and get forwarded into the actual SPA room route.
func handleJoinPage(rooms *room.RoomService, publicClientURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := r.PathValue("roomId")
		roomURL := fmt.Sprintf("%s/room/%s", publicClientURL, roomID)

		if _, err := rooms.Get(roomID); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, joinPageNotFoundTemplate, html.EscapeString(publicClientURL))
			return
		}

		title := "Join a BounceBot game!"
		description := fmt.Sprintf("Room %s is waiting - tap to jump in.", roomID)
		imageURL := fmt.Sprintf("%s/join/%s/preview.png", requestOrigin(r), roomID)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, joinPageTemplate,
			html.EscapeString(title),
			html.EscapeString(description),
			html.EscapeString(imageURL),
			html.EscapeString(roomURL),
			roomURL,
		)
	}
}

// handleJoinPreviewImage serves the dynamically-generated preview PNG at
// /join/{roomId}/preview.png.
func handleJoinPreviewImage(rooms *room.RoomService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := r.PathValue("roomId")

		if _, err := rooms.Get(roomID); err != nil {
			http.NotFound(w, r)
			return
		}

		data, err := preview.Render(roomID)
		if err != nil {
			http.Error(w, "failed to render preview image", http.StatusInternalServerError)
			return
		}

		// The image content is stable for a room's lifetime (only the room
		// ID is drawn, and that never changes), so it's safe for crawlers to
		// cache for a while rather than re-fetching on every share.
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("Content-Type", "image/png")
		w.Write(data)
	}
}
