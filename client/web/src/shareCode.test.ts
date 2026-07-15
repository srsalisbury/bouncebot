import { describe, it, expect } from 'vitest'
import { create } from '@bufbuild/protobuf'
import { GameSchema, BoardSchema, BotPosSchema, PositionSchema, type Game } from './gen/bouncebot_pb'
import { encodeShareCode, decodeShareCode, ShareCodeError } from './shareCode'

// Mirrors model.Game1() in the Go codebase exactly (same board, bots,
// target) - dumped directly from a running instance of that function, so
// this file and model/share_code_test.go can share the same fixed vector.
const GAME1_SHARE_CODE = 'ARAZEDESYyZn4ragpIGHv57cq8p6iB4ZTUhfaBlAEWMFNnb04cWkkYad7KvZifiICx4oTFh4EUESYzbixqSRnuyr2ooeKU1YAE1UrDnEkw'

function pos(x: number, y: number) {
  return create(PositionSchema, { x, y })
}

function game1(): Game {
  const vWalls = [
    [1, 0], [3, 1], [1, 2], [6, 3], [2, 6], [6, 7], [14, 2], [11, 6], [10, 0],
    [10, 4], [8, 1], [8, 7], [11, 15], [9, 14], [13, 12], [10, 11], [12, 10],
    [7, 10], [8, 8], [1, 14], [1, 9], [4, 13], [4, 8], [5, 15], [6, 8],
  ].map(([x, y]) => pos(x!, y!))

  const hWalls = [
    [4, 0], [1, 1], [6, 3], [0, 5], [3, 6], [7, 6], [15, 4], [14, 1], [12, 5],
    [10, 4], [9, 1], [8, 6], [9, 13], [14, 12], [10, 11], [13, 9], [8, 9],
    [15, 8], [8, 8], [0, 11], [1, 14], [2, 8], [4, 12], [5, 8], [7, 8],
  ].map(([x, y]) => pos(x!, y!))

  const possibleTargets = [
    [4, 1], [1, 2], [6, 3], [3, 6], [14, 2], [12, 6], [10, 4], [9, 1], [9, 14],
    [14, 12], [10, 11], [13, 10], [8, 10], [1, 14], [2, 9], [4, 13], [5, 8],
  ].map(([x, y]) => pos(x!, y!))

  const board = create(BoardSchema, { size: 16, vWalls, hWalls, possibleTargets })

  const bots = [
    create(BotPosSchema, { id: 0, pos: pos(5, 4) }),
    create(BotPosSchema, { id: 1, pos: pos(10, 12) }),
    create(BotPosSchema, { id: 2, pos: pos(3, 9) }),
    create(BotPosSchema, { id: 3, pos: pos(12, 4) }),
  ]

  const target = create(BotPosSchema, { id: 0, pos: pos(4, 13) })

  return create(GameSchema, { board, bots, target })
}

describe('shareCode', () => {
  it('encodes Game1 to the same fixed vector as the Go implementation', () => {
    expect(encodeShareCode(game1())).toBe(GAME1_SHARE_CODE)
  })

  it('decodes the fixed vector back to Game1\'s data', () => {
    const decoded = decodeShareCode(GAME1_SHARE_CODE)
    expect(decoded.size).toBe(16)
    expect(decoded.vWalls).toHaveLength(25)
    expect(decoded.hWalls).toHaveLength(25)
    expect(decoded.possibleTargets).toHaveLength(17)
    expect(decoded.targetBotId).toBe(0)
    expect(decoded.targetPos).toEqual({ x: 4, y: 13 })
    expect(decoded.bots).toEqual([
      { x: 5, y: 4 },
      { x: 10, y: 12 },
      { x: 3, y: 9 },
      { x: 12, y: 4 },
    ])
  })

  it('round-trips a real game', () => {
    const code = encodeShareCode(game1())
    const decoded = decodeShareCode(code)
    expect(decoded.vWalls).toHaveLength(25)
    expect(decoded.hWalls).toHaveLength(25)
    expect(decoded.possibleTargets).toHaveLength(17)
    expect(decoded.targetBotId).toBe(0)
    expect(decoded.targetPos).toEqual({ x: 4, y: 13 })
  })

  it('rejects a truncated code', () => {
    const code = encodeShareCode(game1())
    expect(() => decodeShareCode(code.slice(0, code.length / 2))).toThrow(ShareCodeError)
  })

  it('rejects a corrupted checksum', () => {
    const code = encodeShareCode(game1())
    const lastChar = code[code.length - 1]
    const replacement = lastChar === 'A' ? 'B' : 'A'
    const corrupted = code.slice(0, -1) + replacement
    expect(() => decodeShareCode(corrupted)).toThrow(ShareCodeError)
  })

  it('rejects invalid base64', () => {
    expect(() => decodeShareCode('not valid base64url!!!')).toThrow(ShareCodeError)
  })

  it('rejects an empty code', () => {
    expect(() => decodeShareCode('')).toThrow(ShareCodeError)
  })

  it('rejects a missing target', () => {
    const game = create(GameSchema, {
      board: create(BoardSchema, { size: 16, vWalls: [], hWalls: [], possibleTargets: [] }),
      bots: [
        create(BotPosSchema, { id: 0, pos: pos(0, 0) }),
        create(BotPosSchema, { id: 1, pos: pos(1, 0) }),
        create(BotPosSchema, { id: 2, pos: pos(2, 0) }),
        create(BotPosSchema, { id: 3, pos: pos(3, 0) }),
      ],
    })
    expect(() => encodeShareCode(game)).toThrow(ShareCodeError)
  })

  it('rejects out-of-range coordinates', () => {
    const game = create(GameSchema, {
      board: create(BoardSchema, { size: 20, vWalls: [], hWalls: [], possibleTargets: [pos(19, 19)] }),
      bots: [
        create(BotPosSchema, { id: 0, pos: pos(0, 0) }),
        create(BotPosSchema, { id: 1, pos: pos(1, 0) }),
        create(BotPosSchema, { id: 2, pos: pos(2, 0) }),
        create(BotPosSchema, { id: 3, pos: pos(3, 0) }),
      ],
      target: create(BotPosSchema, { id: 0, pos: pos(19, 19) }),
    })
    expect(() => encodeShareCode(game)).toThrow(ShareCodeError)
  })
})
