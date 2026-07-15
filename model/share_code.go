package model

import (
	"encoding/base64"
	"fmt"
)

// shareCodeVersion is the current share code format version. See
// planning/BOARD_SHARING_DESIGN.md for the full byte layout.
const shareCodeVersion = 1

// shareCodeMaxCoord is the largest coordinate value (in either dimension)
// that can be packed into a share code, since each coordinate is packed into
// a 4-bit nibble.
const shareCodeMaxCoord = 15

// EncodeShareCode encodes a game's board (walls and possible targets), robot
// positions, and target into a short, URL-safe token that DecodeShareCode
// can turn back into an identical game.
func EncodeShareCode(game *Game) (string, error) {
	size := game.Board.Size()
	if size <= 0 || size > shareCodeMaxCoord+1 {
		return "", fmt.Errorf("board size %d out of range for share codes (max %d)", size, shareCodeMaxCoord+1)
	}

	vWalls := game.Board.VWalls()
	hWalls := game.Board.HWalls()
	targets := game.Board.PossibleTargets()
	if len(vWalls) > 255 || len(hWalls) > 255 || len(targets) > 255 {
		return "", fmt.Errorf("board has too many walls/targets to encode as a share code")
	}

	if game.Target.Id < 0 || game.Target.Id > 3 {
		return "", fmt.Errorf("target bot id %d out of range", game.Target.Id)
	}

	buf := make([]byte, 0, 96)
	buf = append(buf, shareCodeVersion, byte(size))

	var err error
	if buf, err = appendPositions(buf, vWalls); err != nil {
		return "", err
	}
	if buf, err = appendPositions(buf, hWalls); err != nil {
		return "", err
	}
	if buf, err = appendPositions(buf, targets); err != nil {
		return "", err
	}

	buf = append(buf, byte(game.Target.Id))
	targetByte, err := packPosition(game.Target.Pos)
	if err != nil {
		return "", err
	}
	buf = append(buf, targetByte)

	for id := BotId(0); id < 4; id++ {
		pos, ok := game.Bots[id]
		if !ok {
			return "", fmt.Errorf("missing bot %d", id)
		}
		b, err := packPosition(pos)
		if err != nil {
			return "", err
		}
		buf = append(buf, b)
	}

	buf = append(buf, checksum(buf))

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// DecodeShareCode reverses EncodeShareCode. The resulting game is validated
// the same way any other game is (see NewGame) - malformed or corrupted
// input never panics, only returns an error.
func DecodeShareCode(code string) (*Game, error) {
	buf, err := base64.RawURLEncoding.DecodeString(code)
	if err != nil {
		return nil, fmt.Errorf("invalid share code: %w", err)
	}
	if len(buf) < 1 {
		return nil, fmt.Errorf("share code too short")
	}

	payload, gotSum := buf[:len(buf)-1], buf[len(buf)-1]
	if checksum(payload) != gotSum {
		return nil, fmt.Errorf("share code failed checksum validation")
	}

	r := &byteReader{buf: payload}

	version, err := r.readByte()
	if err != nil {
		return nil, err
	}
	if version != shareCodeVersion {
		return nil, fmt.Errorf("unknown share code version %d", version)
	}

	sizeByte, err := r.readByte()
	if err != nil {
		return nil, err
	}
	if sizeByte == 0 {
		return nil, fmt.Errorf("invalid board size 0")
	}
	size := BoardDim(sizeByte)

	vWalls, err := r.readPositions()
	if err != nil {
		return nil, err
	}
	hWalls, err := r.readPositions()
	if err != nil {
		return nil, err
	}
	targets, err := r.readPositions()
	if err != nil {
		return nil, err
	}

	targetIDByte, err := r.readByte()
	if err != nil {
		return nil, err
	}
	if targetIDByte > 3 {
		return nil, fmt.Errorf("target bot id %d out of range", targetIDByte)
	}
	targetPos, err := r.readPosition()
	if err != nil {
		return nil, err
	}

	bots := make(map[BotId]Position, 4)
	for id := BotId(0); id < 4; id++ {
		pos, err := r.readPosition()
		if err != nil {
			return nil, err
		}
		bots[id] = pos
	}

	if !r.atEnd() {
		return nil, fmt.Errorf("share code has unexpected trailing data")
	}

	board := NewBoardWithTargets(size, vWalls, hWalls, targets)
	target := BotPosition{Id: BotId(targetIDByte), Pos: targetPos}

	return NewGame(board, bots, target)
}

func appendPositions(buf []byte, positions []Position) ([]byte, error) {
	buf = append(buf, byte(len(positions)))
	for _, p := range positions {
		b, err := packPosition(p)
		if err != nil {
			return nil, err
		}
		buf = append(buf, b)
	}
	return buf, nil
}

func packPosition(p Position) (byte, error) {
	if p.X < 0 || p.X > shareCodeMaxCoord || p.Y < 0 || p.Y > shareCodeMaxCoord {
		return 0, fmt.Errorf("position %v out of range for share codes (max %d)", p, shareCodeMaxCoord)
	}
	return byte(p.X)<<4 | byte(p.Y), nil
}

func unpackPosition(b byte) Position {
	return Position{X: BoardDim(b >> 4), Y: BoardDim(b & 0x0F)}
}

func checksum(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return sum
}

// byteReader is a small cursor over a byte slice used by DecodeShareCode.
// Every read is bounds-checked, so a truncated/corrupted token always
// produces an error rather than a panic.
type byteReader struct {
	buf []byte
	pos int
}

func (r *byteReader) readByte() (byte, error) {
	if r.pos >= len(r.buf) {
		return 0, fmt.Errorf("share code truncated")
	}
	b := r.buf[r.pos]
	r.pos++
	return b, nil
}

func (r *byteReader) readPosition() (Position, error) {
	b, err := r.readByte()
	if err != nil {
		return Position{}, err
	}
	return unpackPosition(b), nil
}

func (r *byteReader) readPositions() ([]Position, error) {
	count, err := r.readByte()
	if err != nil {
		return nil, err
	}
	result := make([]Position, count)
	for i := range result {
		pos, err := r.readPosition()
		if err != nil {
			return nil, err
		}
		result[i] = pos
	}
	return result, nil
}

func (r *byteReader) atEnd() bool {
	return r.pos == len(r.buf)
}
