import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useGameInput, type GameInputCallbacks, type GameInputOptions } from './useGameInput'

// Helper to create mock callbacks
function createMockCallbacks(): GameInputCallbacks {
  return {
    onMove: vi.fn(),
    onUndo: vi.fn(),
    onDelete: vi.fn(),
    onNewSolution: vi.fn(),
    onSelectRobot: vi.fn(),
    onSwitchSolution: vi.fn(),
    onSwitchPlayerSolution: vi.fn(),
    onToggleHelp: vi.fn(),
    onCloseHelp: vi.fn(),
  }
}

// Helper to create default options
function createOptions(overrides: Partial<GameInputOptions> = {}): GameInputOptions {
  return {
    inputBlocked: ref(false),
    gameEnded: ref(false),
    helpOpen: ref(false),
    canStartNewSolution: ref(true),
    selectedRobotId: ref(0),
    robotCount: ref(4),
    ...overrides,
  }
}

// Helper to create KeyboardEvent
function keyEvent(key: string, options: { shiftKey?: boolean } = {}): KeyboardEvent {
  return new KeyboardEvent('keydown', { key, ...options })
}

describe('useGameInput', () => {
  describe('movement keys', () => {
    it('calls onMove with correct direction for arrow keys', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions())

      handleKeydown(keyEvent('ArrowUp'))
      expect(callbacks.onMove).toHaveBeenCalledWith('up')

      handleKeydown(keyEvent('ArrowDown'))
      expect(callbacks.onMove).toHaveBeenCalledWith('down')

      handleKeydown(keyEvent('ArrowLeft'))
      expect(callbacks.onMove).toHaveBeenCalledWith('left')

      handleKeydown(keyEvent('ArrowRight'))
      expect(callbacks.onMove).toHaveBeenCalledWith('right')
    })

    it('calls onMove with correct direction for WASD', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions())

      handleKeydown(keyEvent('w'))
      expect(callbacks.onMove).toHaveBeenCalledWith('up')

      handleKeydown(keyEvent('s'))
      expect(callbacks.onMove).toHaveBeenCalledWith('down')

      handleKeydown(keyEvent('a'))
      expect(callbacks.onMove).toHaveBeenCalledWith('left')

      handleKeydown(keyEvent('d'))
      expect(callbacks.onMove).toHaveBeenCalledWith('right')
    })

    it('does not call onMove when no robot selected', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions({
        selectedRobotId: ref(null),
      }))

      handleKeydown(keyEvent('ArrowUp'))
      expect(callbacks.onMove).not.toHaveBeenCalled()
    })
  })

  describe('undo', () => {
    it.each(['z', 'u', 'Escape'])('calls onUndo for key "%s"', (key) => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions())

      handleKeydown(keyEvent(key))
      expect(callbacks.onUndo).toHaveBeenCalled()
    })
  })

  describe('robot selection', () => {
    it('selects robot by number key (1-indexed to 0-indexed)', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions())

      handleKeydown(keyEvent('1'))
      expect(callbacks.onSelectRobot).toHaveBeenCalledWith(0)

      handleKeydown(keyEvent('2'))
      expect(callbacks.onSelectRobot).toHaveBeenCalledWith(1)

      handleKeydown(keyEvent('3'))
      expect(callbacks.onSelectRobot).toHaveBeenCalledWith(2)

      handleKeydown(keyEvent('4'))
      expect(callbacks.onSelectRobot).toHaveBeenCalledWith(3)
    })

    it('ignores numbers beyond robot count', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions({
        robotCount: ref(2),
      }))

      handleKeydown(keyEvent('3'))
      expect(callbacks.onSelectRobot).not.toHaveBeenCalled()

      handleKeydown(keyEvent('4'))
      expect(callbacks.onSelectRobot).not.toHaveBeenCalled()
    })
  })

  describe('new solution', () => {
    it.each(['n', '+'])('calls onNewSolution for key "%s"', (key) => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions())

      handleKeydown(keyEvent(key))
      expect(callbacks.onNewSolution).toHaveBeenCalled()
    })

    it('does not create new solution when at max', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions({
        canStartNewSolution: ref(false),
      }))

      handleKeydown(keyEvent('n'))
      expect(callbacks.onNewSolution).not.toHaveBeenCalled()
    })
  })

  describe('shift commands', () => {
    it('Shift+D calls onDelete', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions())

      handleKeydown(keyEvent('D', { shiftKey: true }))
      expect(callbacks.onDelete).toHaveBeenCalled()
    })

    it('Shift+Arrow switches solutions', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions())

      handleKeydown(keyEvent('ArrowLeft', { shiftKey: true }))
      expect(callbacks.onSwitchSolution).toHaveBeenCalledWith(-1)

      handleKeydown(keyEvent('ArrowRight', { shiftKey: true }))
      expect(callbacks.onSwitchSolution).toHaveBeenCalledWith(1)
    })
  })

  describe('help', () => {
    it('? toggles help', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions())

      handleKeydown(keyEvent('?'))
      expect(callbacks.onToggleHelp).toHaveBeenCalled()
    })

    it('Escape closes help when open', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions({
        helpOpen: ref(true),
      }))

      handleKeydown(keyEvent('Escape'))
      expect(callbacks.onCloseHelp).toHaveBeenCalled()
      expect(callbacks.onUndo).not.toHaveBeenCalled()
    })
  })

  describe('input blocking', () => {
    it('ignores movement when blocked', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions({
        inputBlocked: ref(true),
      }))

      handleKeydown(keyEvent('ArrowUp'))
      handleKeydown(keyEvent('z'))
      handleKeydown(keyEvent('1'))

      expect(callbacks.onMove).not.toHaveBeenCalled()
      expect(callbacks.onUndo).not.toHaveBeenCalled()
      expect(callbacks.onSelectRobot).not.toHaveBeenCalled()
    })

    it('still allows help toggle when blocked', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions({
        inputBlocked: ref(true),
      }))

      handleKeydown(keyEvent('?'))
      expect(callbacks.onToggleHelp).toHaveBeenCalled()
    })
  })

  describe('game ended mode', () => {
    it('ignores regular movement keys', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions({
        gameEnded: ref(true),
      }))

      handleKeydown(keyEvent('ArrowUp'))
      handleKeydown(keyEvent('w'))
      handleKeydown(keyEvent('z'))

      expect(callbacks.onMove).not.toHaveBeenCalled()
      expect(callbacks.onUndo).not.toHaveBeenCalled()
    })

    it('Shift+Arrow switches player solutions', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions({
        gameEnded: ref(true),
      }))

      handleKeydown(keyEvent('ArrowLeft', { shiftKey: true }))
      expect(callbacks.onSwitchPlayerSolution).toHaveBeenCalledWith(-1)

      handleKeydown(keyEvent('ArrowRight', { shiftKey: true }))
      expect(callbacks.onSwitchPlayerSolution).toHaveBeenCalledWith(1)
    })

    it('still allows help toggle', () => {
      const callbacks = createMockCallbacks()
      const { handleKeydown } = useGameInput(callbacks, createOptions({
        gameEnded: ref(true),
      }))

      handleKeydown(keyEvent('?'))
      expect(callbacks.onToggleHelp).toHaveBeenCalled()
    })
  })
})
