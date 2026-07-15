import type { Game } from './gen/bouncebot_pb'

// See planning/BOARD_SHARING_DESIGN.md for the full byte layout. This must
// stay byte-for-byte compatible with model/share_code.go's Go implementation
// - shareCode.test.ts checks both against the same fixed vector to catch
// drift between the two.

const SHARE_CODE_VERSION = 1
const SHARE_CODE_MAX_COORD = 15

export class ShareCodeError extends Error {}

interface PlainPosition {
  x: number
  y: number
}

/** The result of decoding a share code - a plain, proto-independent shape. */
export interface DecodedShareCode {
  size: number
  vWalls: PlainPosition[]
  hWalls: PlainPosition[]
  possibleTargets: PlainPosition[]
  targetBotId: number
  targetPos: PlainPosition
  bots: PlainPosition[] // indexed by bot id 0-3
}

/**
 * Encodes a game's board (walls and possible targets), robot positions, and
 * target into a short, URL-safe token that the server's DecodeShareCode
 * (model/share_code.go) can turn back into an identical game.
 */
export function encodeShareCode(game: Game): string {
  const board = game.board
  if (!board) throw new ShareCodeError('game has no board')
  if (board.size <= 0 || board.size > SHARE_CODE_MAX_COORD + 1) {
    throw new ShareCodeError(`board size ${board.size} out of range for share codes (max ${SHARE_CODE_MAX_COORD + 1})`)
  }

  const { vWalls, hWalls, possibleTargets } = board
  if (vWalls.length > 255 || hWalls.length > 255 || possibleTargets.length > 255) {
    throw new ShareCodeError('board has too many walls/targets to encode as a share code')
  }

  const target = game.target
  if (!target || !target.pos) throw new ShareCodeError('game has no target')
  if (target.id < 0 || target.id > 3) {
    throw new ShareCodeError(`target bot id ${target.id} out of range`)
  }

  const bytes: number[] = [SHARE_CODE_VERSION, board.size]

  appendPositions(bytes, vWalls)
  appendPositions(bytes, hWalls)
  appendPositions(bytes, possibleTargets)

  bytes.push(target.id)
  bytes.push(packPosition(target.pos))

  for (let id = 0; id < 4; id++) {
    const bot = game.bots.find(b => b.id === id)
    if (!bot || !bot.pos) throw new ShareCodeError(`missing bot ${id}`)
    bytes.push(packPosition(bot.pos))
  }

  bytes.push(checksum(bytes))

  return toBase64Url(new Uint8Array(bytes))
}

/**
 * Reverses encodeShareCode. Not used by the production app (the server
 * decodes when consuming a share link) - kept so tests can verify this
 * implementation round-trips and agrees with the Go one on the same fixed
 * vectors.
 */
export function decodeShareCode(code: string): DecodedShareCode {
  const buf = fromBase64Url(code)
  if (buf.length < 1) {
    throw new ShareCodeError('share code too short')
  }

  const payload = buf.slice(0, buf.length - 1)
  const gotSum = buf[buf.length - 1]
  if (checksum(Array.from(payload)) !== gotSum) {
    throw new ShareCodeError('share code failed checksum validation')
  }

  const r = new ByteReader(payload)

  const version = r.readByte()
  if (version !== SHARE_CODE_VERSION) {
    throw new ShareCodeError(`unknown share code version ${version}`)
  }

  const size = r.readByte()
  if (size === 0) {
    throw new ShareCodeError('invalid board size 0')
  }

  const vWalls = r.readPositions()
  const hWalls = r.readPositions()
  const possibleTargets = r.readPositions()

  const targetBotId = r.readByte()
  if (targetBotId > 3) {
    throw new ShareCodeError(`target bot id ${targetBotId} out of range`)
  }
  const targetPos = r.readPosition()

  const bots: PlainPosition[] = []
  for (let id = 0; id < 4; id++) {
    bots.push(r.readPosition())
  }

  if (!r.atEnd()) {
    throw new ShareCodeError('share code has unexpected trailing data')
  }

  return { size, vWalls, hWalls, possibleTargets, targetBotId, targetPos, bots }
}

function appendPositions(bytes: number[], positions: PlainPosition[]): void {
  bytes.push(positions.length)
  for (const p of positions) {
    bytes.push(packPosition(p))
  }
}

function packPosition(p: PlainPosition): number {
  if (p.x < 0 || p.x > SHARE_CODE_MAX_COORD || p.y < 0 || p.y > SHARE_CODE_MAX_COORD) {
    throw new ShareCodeError(`position (${p.x}, ${p.y}) out of range for share codes (max ${SHARE_CODE_MAX_COORD})`)
  }
  return ((p.x & 0x0f) << 4) | (p.y & 0x0f)
}

function unpackPosition(b: number): PlainPosition {
  return { x: (b >> 4) & 0x0f, y: b & 0x0f }
}

function checksum(bytes: number[]): number {
  let sum = 0
  for (const b of bytes) sum = (sum + b) & 0xff
  return sum
}

class ByteReader {
  private pos = 0
  constructor(private buf: Uint8Array) {}

  readByte(): number {
    if (this.pos >= this.buf.length) {
      throw new ShareCodeError('share code truncated')
    }
    return this.buf[this.pos++]!
  }

  readPosition(): PlainPosition {
    return unpackPosition(this.readByte())
  }

  readPositions(): PlainPosition[] {
    const count = this.readByte()
    const result: PlainPosition[] = []
    for (let i = 0; i < count; i++) {
      result.push(this.readPosition())
    }
    return result
  }

  atEnd(): boolean {
    return this.pos === this.buf.length
  }
}

function toBase64Url(bytes: Uint8Array): string {
  let binary = ''
  for (const b of bytes) binary += String.fromCharCode(b)
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

function fromBase64Url(str: string): Uint8Array {
  const base64 = str.replace(/-/g, '+').replace(/_/g, '/')
  const padded = base64 + '='.repeat((4 - (base64.length % 4)) % 4)
  let binary: string
  try {
    binary = atob(padded)
  } catch {
    throw new ShareCodeError('invalid share code encoding')
  }
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return bytes
}
