import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'
import { useRoomConnection } from './useRoomConnection'
import { bounceBotClient } from '../services/connectClient'
import type { Room } from '../gen/bouncebot_pb'

vi.mock('../services/connectClient', () => ({
  bounceBotClient: {
    getRoom: vi.fn(),
  },
}))

// Avoids pulling in the real Pinia store, which reads localStorage at setup
// time (not available/functional in this test environment) - this test is
// only exercising loadRoom's response-ordering guard, not room store state.
vi.mock('../stores/roomStore', () => ({
  useRoomStore: () => ({
    currentPlayerId: null,
    currentSessionToken: null,
    lastRoomId: null,
    setLastRoom: vi.fn(),
    clearRoom: vi.fn(),
  }),
}))

function fakeRoom(overrides: Partial<Room>): Room {
  return {
    id: 'ABCD',
    players: [],
    pendingPlayers: [],
    solutions: [],
    scores: [],
    finishedSolving: [],
    readyForNext: [],
    solverResults: [],
    gamesPlayed: 0,
    currentGame: undefined,
    ...overrides,
  } as unknown as Room
}

describe('useRoomConnection', () => {
  beforeEach(() => {
    vi.mocked(bounceBotClient.getRoom).mockReset()
  })

  it('discards a loadRoom response that resolves after a newer one was already applied', async () => {
    // Simulates two events firing in quick succession (e.g. "ready for next"
    // followed shortly by "game started"), each triggering their own
    // loadRoom call, where the network resolves them out of order.
    const resolvers: Array<(room: Room) => void> = []
    vi.mocked(bounceBotClient.getRoom).mockImplementation(
      () => new Promise((resolve) => { resolvers.push(resolve as (room: Room) => void) })
    )

    const { room, loadRoom } = useRoomConnection({ roomId: ref('ABCD') })

    const staleCall = loadRoom() // issued first (e.g. from "ready for next")
    const freshCall = loadRoom(true) // issued shortly after (e.g. from "game started")

    // The second (fresher) call's response arrives first.
    resolvers[1](fakeRoom({ players: [{ id: 'p1' } as never], pendingPlayers: [] }))
    await freshCall
    expect(room.value?.pendingPlayers).toEqual([])

    // The first (now-stale) call's response arrives late - it must not
    // overwrite the fresher state already applied above.
    resolvers[0](fakeRoom({ players: [], pendingPlayers: [{ id: 'p1' } as never] }))
    await staleCall
    expect(room.value?.pendingPlayers).toEqual([])
  })

  it('still applies responses that resolve in issue order', async () => {
    const rooms = [
      fakeRoom({ gamesPlayed: 0 }),
      fakeRoom({ gamesPlayed: 1 }),
    ]
    vi.mocked(bounceBotClient.getRoom)
      .mockResolvedValueOnce(rooms[0])
      .mockResolvedValueOnce(rooms[1])

    const { room, loadRoom } = useRoomConnection({ roomId: ref('ABCD') })

    await loadRoom()
    expect(room.value?.gamesPlayed).toBe(0)

    await loadRoom()
    expect(room.value?.gamesPlayed).toBe(1)
  })
})
