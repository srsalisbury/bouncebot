# Board Sharing Design

## Overview

Let a player share a specific board they've found interesting — via a URL or a
QR code — so someone else (or themselves, later) can jump straight into solo
play on that exact board: same walls, same starting robot positions, same
target.

## Goals

1. **Shareable**: a URL that works when pasted into text/chat/email, and a QR
   code that works when scanned on a phone, both leading to the same place.
2. **Durable**: unlike a normal room (garbage-collected after 24h of
   inactivity), a share link should keep working indefinitely — it's meant to
   be bookmarked ("save it for future me"), not just passed around in the
   moment.
3. **Transparent**: the URL encodes the full board state directly. No
   server-side lookup table, no expiring IDs. Anyone who understands the
   format can hand-construct one — this is intentionally a small step toward
   a future board editor, whose "export" step would produce exactly this kind
   of link.
4. **Available anywhere**: from solo play or a multiplayer room, since the
   thing being shared (the board) is independent of who's playing it or how
   many players there are.

## Design decisions (from discussion)

- **Landing experience**: opening a share link creates a real solo `Room`
  server-side (not a lightweight, Daily-Challenge-style client-only view).
  The recipient gets the full normal experience — BBot solver comparison,
  room settings, continuing into a new random puzzle afterward via the
  existing "Next Puzzle" flow.
- **Snapshot state**: a share link always captures the board's fresh,
  round-start state — wall layout, original robot positions, target — never
  wherever the sharer's own robots happen to be mid-solve. This falls out for
  free: `Room.CurrentGame` is never mutated by player moves during a round
  (see below), so there's nothing to "reset" before encoding.
- **Encoding scope**: the raw wall/target layout is encoded directly (see
  below), not which of the 12 canonical panels produced it — so this covers
  any board shape from day one, including whatever a future board editor
  might produce, with no separate v2 format needed later.
- **Where "Share" appears**: any active game, solo or multiplayer. Since the
  underlying data (`Room.CurrentGame`) has the same shape either way, this
  costs nothing extra to support.

## Why this is simpler than it first looks

`Room.CurrentGame` is already the "fresh" state, always. Player moves during
a round are client-side only; `SubmitSolution` records a solution's move
list without mutating `room.CurrentGame.Bots`. The board only actually
changes at the *start* of the next round (`GameLifecycle.NewGameSource`/
`CommitNewGame`, added for the min-solution-length feature). So "the
original puzzle" and "what the server currently has recorded" are the same
thing, for the whole round, in both solo and multiplayer rooms — there's
nothing to reconstruct or reset before encoding.

That means the entire feature is: encode the board's walls/targets + 4 robot
positions + 1 target into a token, and a new endpoint that turns a token back
into a `*model.Game` (via the existing `model.NewBoardWithTargets` +
`model.NewGame`, both already used for exactly this kind of reconstruction —
see `NewBoardFromProto`) and commits it as a brand-new room's current game —
reusing `CommitNewGame`, which already exists for exactly this "here's the
game, make it the room's current one" job.

## Encoding format (v1)

Encodes the raw wall/target data directly (`Board.VWalls()`/`HWalls()`/
`PossibleTargets()`, all already exposed on the `Board` interface and, as of
the persistence bug fix earlier, already on the proto `Board` message too) —
not which panels produced it. This works for *any* board shape, not just
combinations of the 12 canonical panels, so there's no v1/v2 split and no
extra plumbing needed to recover "which panels was this built from" (which
isn't tracked anywhere today).

A count-prefixed sparse list per field, one byte per position (x and y each
fit a nibble, since the board is 16×16 → 0-15):

| Bytes | Field | Notes |
|---|---|---|
| 0 | Format version | `1` |
| 1 | Board size | `16` today; encoded rather than assumed |
| 2 | vWall count (N1) | |
| 3..3+N1-1 | vWall positions | 1 byte each, `(x<<4)\|y` |
| next 1 | hWall count (N2) | |
| next N2 | hWall positions | 1 byte each |
| next 1 | Possible-target count (N3) | |
| next N3 | Possible-target positions | 1 byte each |
| next 1 | Target bot ID | 0-3 |
| next 1 | Target position | packed |
| next 4 | Bot 0-3 positions | 1 byte each, packed |
| last 1 | Checksum | sum of all preceding bytes, mod 256 |

For a typical board (25 vWalls, 25 hWalls, 17 possible targets — checked
against a real generated board), that's 12 + 25 + 25 + 17 = **79 bytes**,
base64url-encoded (RFC 4648 §5, no padding) → **~106 characters**. Longer
than a hypothetical panel-only encoding (~23 characters), but sparse lists
still beat a dense presence-bitmask for this data (walls/targets occupy only
~7-10% of their possible slots — a fixed bitmask covering every possible slot
would run ~92-96 bytes). ~106 characters is still trivially fine for a URL
path segment, an SMS, or a QR code, with no percent-encoding needed.

The checksum is defense against a mistyped/truncated link (unlikely for a
scanned QR code, more plausible for a hand-copied URL) producing a
*different but still structurally valid* board instead of an obvious error.

Decoding validates: version and board size are known/sane, all positions are
in bounds, and the resulting board/bots/target pass `model.NewGame`'s
existing validation (no overlapping bots, target bot ID matches a real bot)
— this is the same validation every other game already goes through, just
reached from a new, less-trusted input source (a client-constructed token,
rather than server-generated data). Any failure is a decode error, not a
panic.

Both Go (`model`) and TypeScript (client) need an implementation of this
encode/decode scheme, since creating a link is done client-side (the client
already has the full board in `room.value.currentGame`, so no server round
trip is needed just to *generate* a link) while consuming one is server-side
(a new endpoint reconstructs and validates the game). These two
implementations need to agree exactly — worth a handful of shared fixed test
vectors (same input bytes/fields, checked against the same expected token in
both `model`'s tests and the client's).

## Architecture

### 1. Creating a link (client-side, no server call)

- New `shareCode.ts` (client) with `encode(game: Game): string` mirroring the
  format above, reading directly from `room.value.currentGame`.
- New Share button placed in the "GAME N" title row (`GameBoard.vue`'s
  `.title`, an `<h1>` that already spans the board's own grid column and
  isn't gated behind `!gameEnded` — a single, state-independent location
  needing no separate post-game placement). Positioned to the right of the
  "GAME N" text, pinned to that row's right edge so it lines up with the
  board's own right edge at any screen width. Blue icon-only button, white
  three-dot share glyph.
  - Implementation note: `.title` currently centers "GAME N" via
    `text-align: center` on the `<h1>` itself. The button must be taken out
    of that text flow (`.title` gets `position: relative`, the button gets
    `position: absolute; right: 0` and vertical centering) rather than added
    as a plain third inline child — otherwise the centering would center the
    text+button as a group and visibly shift "GAME N" off the board's true
    center.
- New `ShareModal.vue` (Teleport-based, matching `SettingsModal.vue`'s
  `.modal-backdrop`/`.modal` pattern), titled "Share Board": a QR code of the
  share URL in a white box (needed for scan reliability regardless of the
  app's dark/light theme), then a "Copy Link" button below it — both
  represent the same URL, and the URL itself is never shown as text (it's
  an opaque-looking token with nothing meaningful to read at a glance).
  QR generation is client-side only (a small library rendering to
  canvas/SVG from the URL string) — no server involvement for either half of
  link creation.
  - Reuses the existing clipboard-copy pattern from `RoomView.vue`'s
    `copyShareUrl` (`navigator.clipboard.writeText` with an
    `execCommand('copy')` fallback for older browsers) and its exact "Copy
    Link" label — but copies the constructed `/share/:code` URL, not
    `window.location.href` (the current page), since the share link is a
    different URL than the room being viewed.
  - Unlike the existing `copyShareUrl` button (which gives no feedback on
    click), this one swaps its label to "Copied!" for ~1.5s after a
    successful copy, then reverts.

### 2. Consuming a link (new route + new RPC)

- New route `/share/:code` (`router.ts`) → new `ShareLandingView.vue`.
- On mount: call a new RPC, e.g. `CreateRoomFromSharedBoard(player_name,
  encoded_board)`, mirroring today's solo-play bootstrap
  (`HomeView.vue`'s `startSoloGame`) — no name prompt, same default/saved
  player name behavior — then `router.push` to `/room/:newRoomId`, same as
  normal solo play.
- On decode/validation failure, show a simple error state ("This link looks
  broken or came from a newer version of BounceBot") with a way back to Home,
  rather than a silent redirect or crash.

### 3. Server (`server/room`, `model`)

- New `model/share_code.go`: `EncodeShareCode(game *Game) (string, error)` /
  `DecodeShareCode(code string) (*Game, error)`, implementing the format
  above and its validation.
- New `RoomService` method, e.g. `CreateRoomFromBoard(playerName string, game
  *model.Game) (*Room, string, string, error)`: creates a room the same way
  `Create` does today, then commits the given game directly via the existing
  `GameLifecycle.CommitNewGame(room, game)` — **reused as-is**, no new
  board-selection logic needed, since the board is already fully determined.
  Note this deliberately bypasses `MinSolutionLength`/`generateBoard`'s retry
  loop entirely: there's no "board" to search for or reject, it's fully
  specified by the link.
- Still fires the normal post-commit `onGameStart` hook, so the recipient
  gets the same BBot solver comparison as any other round.
- New RPC `CreateRoomFromSharedBoard` in `bouncebot.proto`, handled in
  `bouncebotserver.go`: decode + validate the token, call
  `RoomService.CreateRoomFromBoard`, return the same shape as
  `CreateRoomResponse` (room, player_id, session_token) so the client can
  reuse existing post-creation logic.

## Error handling

- **Malformed/corrupted token** (bad checksum, wrong version, out-of-range
  position, fails `model.NewGame` validation): RPC returns `InvalidArgument`;
  client shows a friendly "this link looks broken" state.
- **Unknown format version** (a link from a future version of the app):
  treated the same as malformed for now — no forward-compatible partial
  decoding in v1.

## Future considerations (post-v1)

- **Board editor**: place walls/robots/target freely, "export" as a share
  link. The v1 format already encodes arbitrary wall/target layouts, so no
  new format version should be needed for this — just a UI to produce one.
- **"Beat my score"**: optionally embed the sharer's move count in the link
  or in accompanying share text, shown to the recipient as a target to beat.
- **Social link previews**: Open Graph tags for a nice preview card when
  pasted into Slack/iMessage/etc. Needs some form of server-rendered meta
  tags for an otherwise pure client-side SPA — out of scope for v1.

## Work breakdown

1. **`model/share_code.go`**: encode/decode + validation, with unit tests
   including the shared fixed-vector round-trip cases.
2. **Server**: `RoomService.CreateRoomFromBoard`, `CreateRoomFromSharedBoard`
   RPC + proto changes, `bouncebotserver.go` handler.
3. **Client encode + Share UI**: `shareCode.ts`, `ShareModal.vue`, Share
   button wiring in `RoomView.vue`.
4. **Client landing route**: `/share/:code` route, `ShareLandingView.vue`.
5. **QR generation**: pick a small client-side library, wire into
   `ShareModal.vue`.
