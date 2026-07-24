// Package preview generates the "join this game" link-preview image shown
// by chat apps and social platforms (iMessage, WhatsApp, Slack, etc.) when a
// room link is shared. Those crawlers fetch the image directly and don't run
// JavaScript, so the room ID has to be baked into actual pixels server-side.
package preview

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	stddraw "image/draw"
	"image/png"
	"sync"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed assets/logo_dark.png
var logoPNG []byte

//go:embed assets/ConthraxRg-Bold.ttf
var conthraxTTF []byte

const (
	width, height = 1200, 630

	logoSize = 240
	logoY    = 40

	titleFontSize = 74
	titleY        = 370

	roomFontSize = 52
	roomLabelGap = 18 // pixels between "ROOM" and the room ID
	roomY        = 530
)

var (
	bgColor    = color.RGBA{0x1a, 0x1a, 0x1a, 0xff}
	whiteColor = color.RGBA{0xff, 0xff, 0xff, 0xff}
	blueColor  = color.RGBA{0x1e, 0x88, 0xe5, 0xff}
	roomColor  = color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
)

var (
	baseTemplate *image.RGBA
	titleFace    font.Face
	roomFace     font.Face
	buildOnce    sync.Once
	buildErr     error
)

func build() {
	f, err := opentype.Parse(conthraxTTF)
	if err != nil {
		buildErr = err
		return
	}
	titleFace, buildErr = opentype.NewFace(f, &opentype.FaceOptions{Size: titleFontSize, DPI: 72})
	if buildErr != nil {
		return
	}
	roomFace, buildErr = opentype.NewFace(f, &opentype.FaceOptions{Size: roomFontSize, DPI: 72})
	if buildErr != nil {
		return
	}

	logoSrc, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		buildErr = err
		return
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	stddraw.Draw(img, img.Bounds(), image.NewUniform(bgColor), image.Point{}, stddraw.Src)

	logoX := (width - logoSize) / 2
	dstRect := image.Rect(logoX, logoY, logoX+logoSize, logoY+logoSize)
	xdraw.CatmullRom.Scale(img, dstRect, logoSrc, logoSrc.Bounds(), xdraw.Over, nil)

	// Two-tone wordmark: white "BOUNCE" + blue "BOT", matching the homepage.
	dWhite := &font.Drawer{Dst: img, Src: image.NewUniform(whiteColor), Face: titleFace}
	bounceWidth := dWhite.MeasureString("BOUNCE")
	dBlue := &font.Drawer{Dst: img, Src: image.NewUniform(blueColor), Face: titleFace}
	botWidth := dBlue.MeasureString("BOT")
	totalWidth := int(bounceWidth>>6) + int(botWidth>>6)
	startX := (width - totalWidth) / 2

	dWhite.Dot = fixed.P(startX, titleY)
	dWhite.DrawString("BOUNCE")
	dBlue.Dot = fixed.P(startX+int(bounceWidth>>6), titleY)
	dBlue.DrawString("BOT")

	baseTemplate = img
}

// Render draws roomID onto a copy of the cached base template and returns the
// result PNG-encoded. Safe for concurrent use.
func Render(roomID string) ([]byte, error) {
	buildOnce.Do(build)
	if buildErr != nil {
		return nil, buildErr
	}

	img := image.NewRGBA(baseTemplate.Bounds())
	stddraw.Draw(img, img.Bounds(), baseTemplate, image.Point{}, stddraw.Src)

	drawRoomLine(img, roomID)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// drawRoomLine draws "ROOM" and the room ID as one centered unit, with a
// fixed pixel gap between them (rather than relying on a literal space
// character's glyph width, which varies by font).
func drawRoomLine(img stddraw.Image, roomID string) {
	label := "ROOM"
	dLabel := &font.Drawer{Dst: img, Src: image.NewUniform(roomColor), Face: roomFace}
	labelWidth := int(dLabel.MeasureString(label) >> 6)

	dID := &font.Drawer{Dst: img, Src: image.NewUniform(roomColor), Face: roomFace}
	idWidth := int(dID.MeasureString(roomID) >> 6)

	totalWidth := labelWidth + roomLabelGap + idWidth
	startX := (width - totalWidth) / 2

	dLabel.Dot = fixed.P(startX, roomY)
	dLabel.DrawString(label)
	dID.Dot = fixed.P(startX+labelWidth+roomLabelGap, roomY)
	dID.DrawString(roomID)
}
