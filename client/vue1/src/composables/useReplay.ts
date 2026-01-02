import { ref, onUnmounted, type Ref } from 'vue'
import type { PlayerSolution } from '../gen/bouncebot_pb'
import type { Direction } from '../constants'
import { ANIMATION_TIMING } from '../services/AnimationService'

export interface Robot {
  id: number
  x: number
  y: number
}

export interface ReplayCallbacks {
  resetBoard: () => void
  clearCommittedMoves: () => void
  applyReplayMove: (robotId: number, x: number, y: number) => void
}

export interface MoveWithDirection {
  robotId: number
  direction: Direction
  toX: number
  toY: number
}

export function useReplay(
  initialRobots: Ref<Robot[]>,
  callbacks: ReplayCallbacks
) {
  const activePlayerSolutionIndex = ref(0)
  const displayedSolutionIndex = ref(-1)
  const isReplaying = ref(false)
  const replayMoveIndex = ref(0)
  const replayTimeout = ref<number | null>(null)
  const replayRobotPositions = ref<Map<number, { x: number; y: number }>>(new Map())

  function computeDirection(fromX: number, fromY: number, toX: number, toY: number): Direction | null {
    if (toX > fromX) return 'right'
    if (toX < fromX) return 'left'
    if (toY > fromY) return 'down'
    if (toY < fromY) return 'up'
    return null
  }

  function getPlayerSolutionMoves(solution: PlayerSolution | undefined): MoveWithDirection[] {
    if (!solution) return []

    const positions = new Map<number, { x: number; y: number }>()
    for (const robot of initialRobots.value) {
      positions.set(robot.id, { x: robot.x, y: robot.y })
    }

    return solution.moves.map(move => {
      const robotId = move.id
      const toX = move.pos?.x ?? 0
      const toY = move.pos?.y ?? 0
      const from = positions.get(robotId) ?? { x: 0, y: 0 }
      const direction = computeDirection(from.x, from.y, toX, toY)

      positions.set(robotId, { x: toX, y: toY })

      return {
        robotId,
        direction: direction ?? 'right' as Direction,
        toX,
        toY,
      }
    })
  }

  function stopReplay() {
    if (replayTimeout.value) {
      clearTimeout(replayTimeout.value)
      replayTimeout.value = null
    }
    isReplaying.value = false
    replayMoveIndex.value = 0
  }

  function startReplayWithDelay(solutions: PlayerSolution[]) {
    if (!solutions.length) return
    const solution = solutions[activePlayerSolutionIndex.value]
    if (!solution || !solution.moves.length) return

    isReplaying.value = true
    replayMoveIndex.value = 0

    replayRobotPositions.value = new Map()
    for (const robot of initialRobots.value) {
      replayRobotPositions.value.set(robot.id, { x: robot.x, y: robot.y })
    }

    replayTimeout.value = window.setTimeout(() => {
      stepReplay(solutions)
    }, ANIMATION_TIMING.REPLAY_DELAY)
  }

  function stepReplay(solutions: PlayerSolution[]) {
    if (!solutions.length) {
      isReplaying.value = false
      displayedSolutionIndex.value = activePlayerSolutionIndex.value
      return
    }
    const solution = solutions[activePlayerSolutionIndex.value]
    if (!solution || replayMoveIndex.value >= solution.moves.length) {
      isReplaying.value = false
      displayedSolutionIndex.value = activePlayerSolutionIndex.value
      return
    }

    const move = solution.moves[replayMoveIndex.value]
    if (!move || !move.pos) {
      isReplaying.value = false
      displayedSolutionIndex.value = activePlayerSolutionIndex.value
      return
    }

    callbacks.applyReplayMove(move.id, move.pos.x, move.pos.y)
    replayMoveIndex.value++

    replayTimeout.value = window.setTimeout(() => {
      stepReplay(solutions)
    }, ANIMATION_TIMING.REPLAY_DELAY)
  }

  function resetAndReplay(solutions: PlayerSolution[]) {
    // Reset board to initial positions immediately
    callbacks.resetBoard()
    callbacks.clearCommittedMoves()
    displayedSolutionIndex.value = -1

    // Start replay after a delay
    startReplayWithDelay(solutions)
  }

  function switchToPlayerSolution(index: number, solutions: PlayerSolution[]) {
    if (!solutions || index < 0 || index >= solutions.length) return
    if (index === activePlayerSolutionIndex.value && !isReplaying.value) return

    stopReplay()
    activePlayerSolutionIndex.value = index
    resetAndReplay(solutions)
  }

  function startInitialReplay(solutions: PlayerSolution[]) {
    if (solutions.length) {
      activePlayerSolutionIndex.value = 0
      callbacks.resetBoard()
      callbacks.clearCommittedMoves()
      startReplayWithDelay(solutions)
    }
  }

  onUnmounted(() => {
    stopReplay()
  })

  return {
    activePlayerSolutionIndex,
    displayedSolutionIndex,
    isReplaying,
    replayMoveIndex,
    stopReplay,
    switchToPlayerSolution,
    startInitialReplay,
    getPlayerSolutionMoves,
  }
}
