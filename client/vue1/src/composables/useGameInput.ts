import { onMounted, onUnmounted, type Ref } from 'vue'
import type { Direction } from '../constants'

// Key mappings for movement directions
const MOVEMENT_KEYS: Record<string, Direction> = {
  ArrowUp: 'up',
  ArrowDown: 'down',
  ArrowLeft: 'left',
  ArrowRight: 'right',
  w: 'up',
  s: 'down',
  a: 'left',
  d: 'right',
}

export interface GameInputCallbacks {
  onMove: (direction: Direction) => void
  onUndo: () => void
  onDelete: () => void
  onNewSolution: () => void
  onSelectRobot: (index: number) => void
  onSwitchSolution: (delta: number) => void
  onSwitchPlayerSolution: (delta: number) => void
  onToggleHelp: () => void
  onCloseHelp: () => void
}

export interface GameInputOptions {
  inputBlocked: Ref<boolean>
  gameEnded: Ref<boolean>
  helpOpen: Ref<boolean>
  canStartNewSolution: Ref<boolean>
  selectedRobotId: Ref<number | null>
  robotCount: Ref<number>
}

export function useGameInput(callbacks: GameInputCallbacks, options: GameInputOptions) {
  const {
    inputBlocked,
    gameEnded,
    helpOpen,
    canStartNewSolution,
    selectedRobotId,
    robotCount,
  } = options

  function handleKeydown(event: KeyboardEvent) {
    const { key, shiftKey } = event

    // Help toggle (works in all modes)
    if (key === '?') {
      callbacks.onToggleHelp()
      return
    }

    // Close help modal with Escape
    if (key === 'Escape' && helpOpen.value) {
      callbacks.onCloseHelp()
      return
    }

    // Block all other input when a modal is open
    if (inputBlocked.value) {
      return
    }

    // Game ended mode - only allow navigating player solutions
    if (gameEnded.value) {
      if (shiftKey) {
        if (key === 'ArrowLeft') {
          event.preventDefault()
          callbacks.onSwitchPlayerSolution(-1)
          return
        }
        if (key === 'ArrowRight') {
          event.preventDefault()
          callbacks.onSwitchPlayerSolution(1)
          return
        }
      }
      return // Ignore other keys in game-ended mode
    }

    // Normal game mode below

    // Undo
    if (key === 'z' || key === 'u' || key === 'Escape') {
      callbacks.onUndo()
      return
    }

    // Shift commands
    if (shiftKey) {
      if (key === 'D') {
        callbacks.onDelete()
        return
      }
      if (key === 'ArrowLeft') {
        event.preventDefault()
        callbacks.onSwitchSolution(-1)
        return
      }
      if (key === 'ArrowRight') {
        event.preventDefault()
        callbacks.onSwitchSolution(1)
        return
      }
    }

    // New solution
    if ((key === 'n' || key === '+') && canStartNewSolution.value) {
      callbacks.onNewSolution()
      return
    }

    // Robot selection (1-4)
    const num = parseInt(key)
    if (num >= 1 && num <= robotCount.value) {
      callbacks.onSelectRobot(num - 1)
      return
    }

    // Movement
    const direction = MOVEMENT_KEYS[key]
    if (direction && selectedRobotId.value !== null) {
      event.preventDefault()
      callbacks.onMove(direction)
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeydown)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown)
  })

  return {
    handleKeydown,
  }
}
