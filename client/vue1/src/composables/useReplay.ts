import { ref, onUnmounted, type Ref } from 'vue'
import type { PlayerSolution, BotPos } from '../gen/bouncebot_pb'
import type { Direction } from '../constants'

export interface Robot {
  id: number
  x: number
  y: number
}

export interface ReplayCallbacks {
  resetBoard: () => void
  clearCommittedMoves: () => void
  applyReplayMove: (robotId: number, x: number, y: number) => void
  unwindReplayMove: (robotId: number, x: number, y: number) => void
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
  const isUnwinding = ref(false)
  const replayMoveIndex = ref(0)
  const unwindMoveIndex = ref(0)
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
    isUnwinding.value = false
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
    }, 600)
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
    }, 600)
  }

  function unwindStep(moves: BotPos[], positionHistory: Map<number, { x: number; y: number }>[], solutions: PlayerSolution[]) {
    if (unwindMoveIndex.value < 0) {
      isUnwinding.value = false
      displayedSolutionIndex.value = -1
      startReplayWithDelay(solutions)
      return
    }

    const move = moves[unwindMoveIndex.value]
    if (!move) {
      isUnwinding.value = false
      displayedSolutionIndex.value = -1
      startReplayWithDelay(solutions)
      return
    }

    const beforePositions = positionHistory[unwindMoveIndex.value]
    const beforePos = beforePositions?.get(move.id)

    if (beforePos) {
      callbacks.unwindReplayMove(move.id, beforePos.x, beforePos.y)
    }

    unwindMoveIndex.value--

    replayTimeout.value = window.setTimeout(() => {
      unwindStep(moves, positionHistory, solutions)
    }, 150)
  }

  function unwindThenReplay(solutions: PlayerSolution[]) {
    if (displayedSolutionIndex.value < 0 || !solutions[displayedSolutionIndex.value]) {
      callbacks.resetBoard()
      callbacks.clearCommittedMoves()
      startReplayWithDelay(solutions)
      return
    }

    const displayedSolution = solutions[displayedSolutionIndex.value]
    if (!displayedSolution || displayedSolution.moves.length === 0) {
      callbacks.resetBoard()
      callbacks.clearCommittedMoves()
      startReplayWithDelay(solutions)
      return
    }

    isUnwinding.value = true
    unwindMoveIndex.value = displayedSolution.moves.length - 1

    const positionHistory: Map<number, { x: number; y: number }>[] = []

    const initialPositions = new Map<number, { x: number; y: number }>()
    for (const robot of initialRobots.value) {
      initialPositions.set(robot.id, { x: robot.x, y: robot.y })
    }
    positionHistory.push(new Map(initialPositions))

    let currentPositions = new Map(initialPositions)
    for (const move of displayedSolution.moves) {
      currentPositions = new Map(currentPositions)
      currentPositions.set(move.id, { x: move.pos?.x ?? 0, y: move.pos?.y ?? 0 })
      positionHistory.push(new Map(currentPositions))
    }

    replayRobotPositions.value = positionHistory[unwindMoveIndex.value] ?? new Map()

    unwindStep(displayedSolution.moves, positionHistory, solutions)
  }

  function switchToPlayerSolution(index: number, solutions: PlayerSolution[]) {
    if (!solutions || index < 0 || index >= solutions.length) return
    if (index === activePlayerSolutionIndex.value && !isReplaying.value && !isUnwinding.value) return

    stopReplay()
    activePlayerSolutionIndex.value = index
    unwindThenReplay(solutions)
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
    isUnwinding,
    replayMoveIndex,
    stopReplay,
    switchToPlayerSolution,
    startInitialReplay,
    getPlayerSolutionMoves,
  }
}
