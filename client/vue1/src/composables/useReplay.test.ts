import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref } from 'vue'
import { useReplay, type Robot, type ReplayCallbacks } from './useReplay'
import type { PlayerSolution } from '../gen/bouncebot_pb'

// Helper to create mock callbacks
function createMockCallbacks(): ReplayCallbacks {
  return {
    resetBoard: vi.fn(),
    clearCommittedMoves: vi.fn(),
    applyReplayMove: vi.fn(),
    unwindReplayMove: vi.fn(),
  }
}

// Helper to create a mock solution
function createSolution(moves: { id: number; x: number; y: number }[]): PlayerSolution {
  return {
    playerId: 'player1',
    moves: moves.map(m => ({
      id: m.id,
      pos: { x: m.x, y: m.y },
    })),
    $typeName: 'bouncebot.PlayerSolution',
    $unknown: undefined,
  } as PlayerSolution
}

describe('useReplay', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe('getPlayerSolutionMoves', () => {
    it('returns empty array for undefined solution', () => {
      const robots = ref<Robot[]>([{ id: 0, x: 5, y: 5 }])
      const { getPlayerSolutionMoves } = useReplay(robots, createMockCallbacks())

      expect(getPlayerSolutionMoves(undefined)).toEqual([])
    })

    it('computes direction from position changes', () => {
      const robots = ref<Robot[]>([
        { id: 0, x: 5, y: 5 },
        { id: 1, x: 10, y: 10 },
      ])
      const { getPlayerSolutionMoves } = useReplay(robots, createMockCallbacks())

      const solution = createSolution([
        { id: 0, x: 5, y: 0 },   // up (y decreases)
        { id: 0, x: 10, y: 0 },  // right (x increases)
        { id: 0, x: 10, y: 5 },  // down (y increases)
        { id: 0, x: 0, y: 5 },   // left (x decreases)
      ])

      const moves = getPlayerSolutionMoves(solution)

      expect(moves).toHaveLength(4)
      expect(moves[0]).toEqual({ robotId: 0, direction: 'up', toX: 5, toY: 0 })
      expect(moves[1]).toEqual({ robotId: 0, direction: 'right', toX: 10, toY: 0 })
      expect(moves[2]).toEqual({ robotId: 0, direction: 'down', toX: 10, toY: 5 })
      expect(moves[3]).toEqual({ robotId: 0, direction: 'left', toX: 0, toY: 5 })
    })

    it('tracks position through multiple moves', () => {
      const robots = ref<Robot[]>([{ id: 0, x: 0, y: 0 }])
      const { getPlayerSolutionMoves } = useReplay(robots, createMockCallbacks())

      // Robot starts at (0,0), moves right to (5,0), then down to (5,5)
      const solution = createSolution([
        { id: 0, x: 5, y: 0 },
        { id: 0, x: 5, y: 5 },
      ])

      const moves = getPlayerSolutionMoves(solution)

      expect(moves[0]?.direction).toBe('right')
      expect(moves[1]?.direction).toBe('down')
    })

    it('handles multiple robots', () => {
      const robots = ref<Robot[]>([
        { id: 0, x: 0, y: 0 },
        { id: 1, x: 10, y: 10 },
      ])
      const { getPlayerSolutionMoves } = useReplay(robots, createMockCallbacks())

      const solution = createSolution([
        { id: 0, x: 5, y: 0 },  // Robot 0 moves right
        { id: 1, x: 10, y: 0 }, // Robot 1 moves up
        { id: 0, x: 5, y: 5 },  // Robot 0 moves down
      ])

      const moves = getPlayerSolutionMoves(solution)

      expect(moves[0]).toEqual({ robotId: 0, direction: 'right', toX: 5, toY: 0 })
      expect(moves[1]).toEqual({ robotId: 1, direction: 'up', toX: 10, toY: 0 })
      expect(moves[2]).toEqual({ robotId: 0, direction: 'down', toX: 5, toY: 5 })
    })
  })

  describe('initial state', () => {
    it('starts with correct default values', () => {
      const robots = ref<Robot[]>([])
      const {
        activePlayerSolutionIndex,
        displayedSolutionIndex,
        isReplaying,
        isUnwinding,
      } = useReplay(robots, createMockCallbacks())

      expect(activePlayerSolutionIndex.value).toBe(0)
      expect(displayedSolutionIndex.value).toBe(-1)
      expect(isReplaying.value).toBe(false)
      expect(isUnwinding.value).toBe(false)
    })
  })

  describe('startInitialReplay', () => {
    it('does nothing with empty solutions', () => {
      const robots = ref<Robot[]>([{ id: 0, x: 5, y: 5 }])
      const callbacks = createMockCallbacks()
      const { startInitialReplay } = useReplay(robots, callbacks)

      startInitialReplay([])

      expect(callbacks.resetBoard).not.toHaveBeenCalled()
    })

    it('resets board and starts replay', () => {
      const robots = ref<Robot[]>([{ id: 0, x: 5, y: 5 }])
      const callbacks = createMockCallbacks()
      const { startInitialReplay, isReplaying } = useReplay(robots, callbacks)

      const solutions = [createSolution([{ id: 0, x: 5, y: 0 }])]
      startInitialReplay(solutions)

      expect(callbacks.resetBoard).toHaveBeenCalled()
      expect(callbacks.clearCommittedMoves).toHaveBeenCalled()
      expect(isReplaying.value).toBe(true)
    })

    it('applies moves after delay', () => {
      const robots = ref<Robot[]>([{ id: 0, x: 5, y: 5 }])
      const callbacks = createMockCallbacks()
      const { startInitialReplay } = useReplay(robots, callbacks)

      const solutions = [createSolution([{ id: 0, x: 5, y: 0 }])]
      startInitialReplay(solutions)

      // Move not applied yet
      expect(callbacks.applyReplayMove).not.toHaveBeenCalled()

      // After REPLAY_DELAY (600ms)
      vi.advanceTimersByTime(600)
      expect(callbacks.applyReplayMove).toHaveBeenCalledWith(0, 5, 0)
    })
  })

  describe('stopReplay', () => {
    it('clears replay state', () => {
      const robots = ref<Robot[]>([{ id: 0, x: 5, y: 5 }])
      const callbacks = createMockCallbacks()
      const { startInitialReplay, stopReplay, isReplaying } = useReplay(robots, callbacks)

      const solutions = [createSolution([{ id: 0, x: 5, y: 0 }])]
      startInitialReplay(solutions)
      expect(isReplaying.value).toBe(true)

      stopReplay()
      expect(isReplaying.value).toBe(false)
    })

    it('prevents further move applications', () => {
      const robots = ref<Robot[]>([{ id: 0, x: 5, y: 5 }])
      const callbacks = createMockCallbacks()
      const { startInitialReplay, stopReplay } = useReplay(robots, callbacks)

      const solutions = [createSolution([
        { id: 0, x: 5, y: 0 },
        { id: 0, x: 10, y: 0 },
      ])]
      startInitialReplay(solutions)

      // Apply first move
      vi.advanceTimersByTime(600)
      expect(callbacks.applyReplayMove).toHaveBeenCalledTimes(1)

      // Stop before second move
      stopReplay()
      vi.advanceTimersByTime(600)

      // Second move should not be applied
      expect(callbacks.applyReplayMove).toHaveBeenCalledTimes(1)
    })
  })

  describe('switchToPlayerSolution', () => {
    it('ignores invalid index', () => {
      const robots = ref<Robot[]>([{ id: 0, x: 5, y: 5 }])
      const callbacks = createMockCallbacks()
      const { switchToPlayerSolution, activePlayerSolutionIndex } = useReplay(robots, callbacks)

      const solutions = [createSolution([{ id: 0, x: 5, y: 0 }])]

      switchToPlayerSolution(-1, solutions)
      expect(activePlayerSolutionIndex.value).toBe(0)

      switchToPlayerSolution(5, solutions)
      expect(activePlayerSolutionIndex.value).toBe(0)
    })

    it('updates active index and triggers replay', () => {
      const robots = ref<Robot[]>([{ id: 0, x: 5, y: 5 }])
      const callbacks = createMockCallbacks()
      const { switchToPlayerSolution, activePlayerSolutionIndex, displayedSolutionIndex } = useReplay(robots, callbacks)

      // Set displayed solution to simulate previous replay completed
      displayedSolutionIndex.value = 0

      const solutions = [
        createSolution([{ id: 0, x: 5, y: 0 }]),
        createSolution([{ id: 0, x: 0, y: 5 }]),
      ]

      switchToPlayerSolution(1, solutions)

      expect(activePlayerSolutionIndex.value).toBe(1)
    })
  })
})
