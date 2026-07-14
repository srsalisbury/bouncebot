import { defineStore } from 'pinia'
import { ref, computed, watch, toRaw } from 'vue'
import type { Game } from '../gen/bouncebot_pb'
import { MAX_SOLUTIONS, OPPOSITE_DIRECTION, SOLUTION_SWITCH_DELAY_MS, type Direction } from '../constants'
import { calculateDestination } from '../gamePhysics'
import { ANIMATION_TIMING } from '../services/AnimationService'

export type Robot = {
  id: number
  x: number
  y: number
}

export type Wall = {
  x: number
  y: number
}

export type Target = {
  robotId: number
  x: number
  y: number
}

export type Move = {
  robotId: number
  direction: Direction
  fromX: number
  fromY: number
  toX: number
  toY: number
}

export type Solution = {
  moves: Move[]
  isSolved: boolean
}

// Persisted solution state
type PersistedSolutions = {
  gameNumber: number
  solutions: Solution[]
  activeSolutionIndex: number
}

const STORAGE_KEY_PREFIX = 'bouncebot_solutions_'

function getSolutionsStorageKey(roomId: string): string {
  return `${STORAGE_KEY_PREFIX}${roomId}`
}

function loadPersistedSolutions(roomId: string): PersistedSolutions | null {
  try {
    const stored = localStorage.getItem(getSolutionsStorageKey(roomId))
    if (!stored) return null
    return JSON.parse(stored) as PersistedSolutions
  } catch {
    return null
  }
}

function savePersistedSolutions(roomId: string, data: PersistedSolutions): void {
  try {
    localStorage.setItem(getSolutionsStorageKey(roomId), JSON.stringify(data))
  } catch {
    // Ignore storage errors (quota exceeded, etc.)
  }
}

function clearPersistedSolutions(roomId: string): void {
  try {
    localStorage.removeItem(getSolutionsStorageKey(roomId))
  } catch {
    // Ignore errors
  }
}

export const useGameStore = defineStore('game', () => {
  // Loading state
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Initial state (stored for reset)
  const initialRobots = ref<Robot[]>([])

  // Current game/room tracking for persistence
  const currentRoomId = ref<string | null>(null)
  const currentGameNumber = ref<number>(0)

  // Game state
  const robots = ref<Robot[]>([])
  const vWalls = ref<Wall[]>([])
  const hWalls = ref<Wall[]>([])
  const target = ref<Target>({
    robotId: 0,
    x: 0,
    y: 0,
  })

  // Selected robot state
  const selectedRobotId = ref<number | null>(null)

  // Solutions tracking
  const solutions = ref<Solution[]>([{ moves: [], isSolved: false }])
  const activeSolutionIndex = ref(0)
  const animatingMoveIndex = ref<number | null>(null)

  // Committed moves (for history dots - delayed to match animation)
  const committedMoves = ref<Move[]>([])

  // Track pending committed move timeouts so we can cancel them on forgiveness
  const pendingMoveTimeouts = new Map<Move, number>()

  // Track pending replay timeouts and animation frames so we can cancel them
  const pendingTimeouts = ref<number[]>([])
  const pendingAnimationFrame = ref<number | null>(null)

  // Persist solutions when they change
  function persistSolutions() {
    if (!currentRoomId.value || currentGameNumber.value === 0) return
    savePersistedSolutions(currentRoomId.value, {
      gameNumber: currentGameNumber.value,
      solutions: solutions.value,
      activeSolutionIndex: activeSolutionIndex.value,
    })
  }

  // Watch for solution changes and persist
  watch(
    [solutions, activeSolutionIndex],
    () => persistSolutions(),
    { deep: true }
  )

  // Helper to find robot by ID
  function findRobotById(id: number): Robot | undefined {
    return robots.value.find(r => r.id === id)
  }

  // Computed
  const activeSolution = computed(() => solutions.value[activeSolutionIndex.value]!)
  const moves = computed(() => activeSolution.value.moves)
  const moveCount = computed(() => moves.value.length)

  const isSolved = computed(() => {
    const targetRobot = findRobotById(target.value.robotId)
    if (!targetRobot) return false
    return targetRobot.x === target.value.x && targetRobot.y === target.value.y
  })

  const hasAnySolvedSolution = computed(() => {
    return solutions.value.some(s => s.isSolved)
  })

  const canStartNewSolution = computed(() => {
    return solutions.value.length < MAX_SOLUTIONS
  })

  // Find the index of the worst solution to delete when making room for a new one.
  // bestSubmittedIndex is the index of the best solution that has been submitted to the server (if any).
  // Returns -1 if no solution should be deleted.
  //
  // Deletion priority:
  // 1. If any unsolved solutions exist -> delete the most recent unsolved (highest index)
  // 2. If all solutions are solved -> delete the longest one (most moves)
  function findWorstSolutionIndex(bestSubmittedIndex: number | null): number {
    if (solutions.value.length < MAX_SOLUTIONS) {
      return -1 // No need to delete anything
    }

    // Find indices that are candidates for deletion (not the best submitted)
    const candidates: number[] = []
    for (let i = 0; i < solutions.value.length; i++) {
      if (i !== bestSubmittedIndex) {
        candidates.push(i)
      }
    }

    if (candidates.length === 0) {
      return -1 // Can't delete anything (only best submitted exists)
    }

    // Check if any candidate is unsolved
    const unsolvedCandidates = candidates.filter(i => !solutions.value[i]!.isSolved)

    if (unsolvedCandidates.length > 0) {
      // Delete the most recent unsolved solution (highest index)
      return Math.max(...unsolvedCandidates)
    }

    // All candidates are solved - delete the longest one (most moves)
    let worstIndex = candidates[0]!
    let worstMoveCount = solutions.value[worstIndex]!.moves.length
    for (let i = 1; i < candidates.length; i++) {
      const candidateIndex = candidates[i]!
      const moveCount = solutions.value[candidateIndex]!.moves.length
      if (moveCount > worstMoveCount) {
        worstMoveCount = moveCount
        worstIndex = candidateIndex
      }
    }

    return worstIndex
  }

  // Actions
  function selectRobot(robotId: number) {
    if (selectedRobotId.value === robotId) {
      selectedRobotId.value = null
    } else {
      selectedRobotId.value = robotId
    }
  }

  function moveRobot(direction: Direction) {
    if (selectedRobotId.value === null) return
    if (isSolved.value) return

    const robot = findRobotById(selectedRobotId.value)
    if (!robot) return

    const solutionMoves = activeSolution.value.moves
    const lastMoveProxy = solutionMoves[solutionMoves.length - 1]
    const lastMove = lastMoveProxy ? toRaw(lastMoveProxy) : undefined

    // Input forgiveness: detect opposite direction move on same robot
    if (lastMove && lastMove.robotId === robot.id && direction === OPPOSITE_DIRECTION[lastMove.direction]) {
      // Calculate where the robot would go from the original position (before last move)
      // We need to temporarily move the robot back to simulate from original position
      const originalX = lastMove.fromX
      const originalY = lastMove.fromY
      const currentX = robot.x
      const currentY = robot.y

      // Temporarily move robot back to calculate destination from original position
      robot.x = originalX
      robot.y = originalY
      const newDestination = calculateDestination(robot, direction, robots.value, vWalls.value, hWalls.value)
      // Restore robot position
      robot.x = currentX
      robot.y = currentY

      // Cancel pending timeout for the last move's committed dot
      const pendingTimeout = pendingMoveTimeouts.get(lastMove)
      if (pendingTimeout !== undefined) {
        clearTimeout(pendingTimeout)
        pendingMoveTimeouts.delete(lastMove)
      }

      if (newDestination.x === originalX && newDestination.y === originalY) {
        // Robot would end up back at original position - undo both moves (treat as no move)
        solutionMoves.pop()
        robot.x = originalX
        robot.y = originalY

        // Remove from committedMoves if present
        const committedIndex = committedMoves.value.indexOf(lastMove)
        if (committedIndex !== -1) {
          committedMoves.value.splice(committedIndex, 1)
        }
        return
      } else {
        // Robot ends up somewhere different - replace last move with direct move from original to new destination
        solutionMoves.pop()

        // Remove from committedMoves if present
        const committedIndex = committedMoves.value.indexOf(lastMove)
        if (committedIndex !== -1) {
          committedMoves.value.splice(committedIndex, 1)
        }

        // Create new move from original position to new destination
        const newMove: Move = {
          robotId: robot.id,
          direction,
          fromX: originalX,
          fromY: originalY,
          toX: newDestination.x,
          toY: newDestination.y,
        }
        solutionMoves.push(newMove)
        robot.x = newDestination.x
        robot.y = newDestination.y

        // Delay adding to committedMoves to match animation
        const timeoutId = window.setTimeout(() => {
          pendingMoveTimeouts.delete(newMove)
          committedMoves.value.push(newMove)
        }, ANIMATION_TIMING.MOVE_DELAY)
        pendingMoveTimeouts.set(newMove, timeoutId)

        // Check if puzzle is now solved
        const targetRobot = findRobotById(target.value.robotId)
        if (targetRobot && targetRobot.x === target.value.x && targetRobot.y === target.value.y) {
          activeSolution.value.isSolved = true
        }
        return
      }
    }

    // Normal move processing
    const destination = calculateDestination(robot, direction, robots.value, vWalls.value, hWalls.value)

    // Only count as a move if the robot actually moved
    if (destination.x !== robot.x || destination.y !== robot.y) {
      const move: Move = {
        robotId: robot.id,
        direction,
        fromX: robot.x,
        fromY: robot.y,
        toX: destination.x,
        toY: destination.y,
      }
      activeSolution.value.moves.push(move)
      robot.x = destination.x
      robot.y = destination.y

      // Delay adding to committedMoves to match animation
      const timeoutId = window.setTimeout(() => {
        pendingMoveTimeouts.delete(move)
        committedMoves.value.push(move)
      }, ANIMATION_TIMING.MOVE_DELAY)
      pendingMoveTimeouts.set(move, timeoutId)

      // Check if puzzle is now solved and mark the solution
      const targetRobot = findRobotById(target.value.robotId)
      if (targetRobot && targetRobot.x === target.value.x && targetRobot.y === target.value.y) {
        activeSolution.value.isSolved = true
      }
    }
  }

  function undoMove() {
    const solutionMoves = activeSolution.value.moves
    if (solutionMoves.length === 0) return

    const lastMove = toRaw(solutionMoves.pop()!)

    // Cancel pending timeout for this move so it doesn't re-add to committedMoves
    const pendingTimeout = pendingMoveTimeouts.get(lastMove)
    if (pendingTimeout !== undefined) {
      clearTimeout(pendingTimeout)
      pendingMoveTimeouts.delete(lastMove)
    }

    // Also remove from committedMoves if already committed
    const committedIndex = committedMoves.value.indexOf(lastMove)
    if (committedIndex !== -1) {
      committedMoves.value.splice(committedIndex, 1)
    }

    const robot = findRobotById(lastMove.robotId)
    if (!robot) return

    robot.x = lastMove.fromX
    robot.y = lastMove.fromY
    selectedRobotId.value = lastMove.robotId

    // Mark solution as unsolved since we undid a move
    activeSolution.value.isSolved = false
  }

  // Cancel all pending replay timeouts and animation frames
  function cancelPendingTimeouts() {
    for (const id of pendingTimeouts.value) {
      clearTimeout(id)
    }
    pendingTimeouts.value = []
    if (pendingAnimationFrame.value !== null) {
      cancelAnimationFrame(pendingAnimationFrame.value)
      pendingAnimationFrame.value = null
    }
  }

  // Shared function to replay moves with animation, starting after a delay
  function replayMoves(movesToReplay: Move[], startDelay: number): number {
    movesToReplay.forEach((move, i) => {
      const moveStartTime = startDelay + i * ANIMATION_TIMING.MOVE_DELAY
      const moveTimeoutId = window.setTimeout(() => {
        animatingMoveIndex.value = i
        const robot = findRobotById(move.robotId)
        if (robot) {
          robot.x = move.toX
          robot.y = move.toY
        }
      }, moveStartTime)
      pendingTimeouts.value.push(moveTimeoutId)
      // Add dot after animation completes
      const dotTimeoutId = window.setTimeout(() => {
        committedMoves.value.push(move)
      }, moveStartTime + ANIMATION_TIMING.MOVE_ANIMATION)
      pendingTimeouts.value.push(dotTimeoutId)
    })
    return startDelay + movesToReplay.length * ANIMATION_TIMING.MOVE_DELAY
  }

  function switchSolution(index: number) {
    if (index < 0 || index >= solutions.value.length) return
    if (index === activeSolutionIndex.value) return

    // Cancel any in-progress replay
    cancelPendingTimeouts()

    const targetMoves = solutions.value[index]!.moves

    selectedRobotId.value = null

    // Reset to initial positions immediately
    resetBoard()
    committedMoves.value = []
    animatingMoveIndex.value = null

    // Wait a frame for reset to settle, then switch and replay
    pendingAnimationFrame.value = requestAnimationFrame(() => {
      pendingAnimationFrame.value = null
      activeSolutionIndex.value = index

      // Wait before replaying new solution
      const resetDelay = SOLUTION_SWITCH_DELAY_MS

      // Replay target moves after delay
      const totalTime = replayMoves(targetMoves, resetDelay)

      // Clear highlight after replay completes
      const clearTimeoutId = window.setTimeout(() => {
        animatingMoveIndex.value = null
      }, totalTime)
      pendingTimeouts.value.push(clearTimeoutId)
    })
  }

  function replaySolution() {
    const targetMoves = solutions.value[activeSolutionIndex.value]?.moves
    if (!targetMoves) return

    // Cancel any in-progress replay
    cancelPendingTimeouts()

    selectedRobotId.value = null

    // Reset to initial positions immediately
    resetBoard()
    committedMoves.value = []
    animatingMoveIndex.value = null

    // Wait before replaying solution
    const resetDelay = SOLUTION_SWITCH_DELAY_MS

    // Replay target moves after delay
    const totalTime = replayMoves(targetMoves, resetDelay)

    // Clear highlight after replay completes
    const clearTimeoutId = window.setTimeout(() => {
      animatingMoveIndex.value = null
    }, totalTime)
    pendingTimeouts.value.push(clearTimeoutId)
  }

  type StartNewSolutionResult = {
    newIndex: number
    deletedIndex: number | null
  }

  // Start a new solution slot. If at max capacity, auto-deletes the worst solution first.
  // bestSubmittedIndex: the index of the best submitted solution to protect from deletion.
  // Returns an object with the new active solution index and the deleted index (if any).
  // Returns newIndex: -1 if a new solution couldn't be created.
  function startNewSolution(bestSubmittedIndex: number | null = null): StartNewSolutionResult {
    // Cancel any in-progress replay
    cancelPendingTimeouts()

    let deletedIndex: number | null = null

    // If at max capacity, delete the worst solution first
    if (solutions.value.length >= MAX_SOLUTIONS) {
      const worstIndex = findWorstSolutionIndex(bestSubmittedIndex)
      if (worstIndex === -1) {
        // Can't delete anything, can't create new solution
        return { newIndex: -1, deletedIndex: null }
      }
      // Delete the worst solution (internally, not through deleteSolution to avoid animation)
      solutions.value.splice(worstIndex, 1)
      deletedIndex = worstIndex
      // Adjust active index if needed
      if (activeSolutionIndex.value > worstIndex) {
        activeSolutionIndex.value--
      } else if (activeSolutionIndex.value === worstIndex) {
        // We just deleted the active solution, pick the first one
        activeSolutionIndex.value = 0
      }
    }

    // Create new empty solution
    solutions.value.push({ moves: [], isSolved: false })
    const newIndex = solutions.value.length - 1

    selectedRobotId.value = null

    // Reset to initial positions immediately
    resetBoard()
    committedMoves.value = []
    animatingMoveIndex.value = null
    activeSolutionIndex.value = newIndex

    return { newIndex, deletedIndex }
  }

  function deleteSolution(index: number) {
    // Can't delete if only one solution remains
    if (solutions.value.length <= 1) return
    if (index < 0 || index >= solutions.value.length) return

    // Cancel any in-progress replay
    cancelPendingTimeouts()

    // If deleting the active solution, switch to another first
    if (index === activeSolutionIndex.value) {
      // Switch to previous solution, or first if deleting index 0
      const newActiveIndex = index > 0 ? index - 1 : 1
      const targetMoves = solutions.value[newActiveIndex]!.moves

      selectedRobotId.value = null

      // Reset to initial positions immediately
      resetBoard()
      committedMoves.value = []
      animatingMoveIndex.value = null

      // Remove solution and update index
      solutions.value.splice(index, 1)
      activeSolutionIndex.value = index > 0 ? index - 1 : 0

      // Wait before replaying new solution
      const resetDelay = SOLUTION_SWITCH_DELAY_MS

      // Replay new active solution
      const totalTime = replayMoves(targetMoves, resetDelay)

      const clearTimeoutId = window.setTimeout(() => {
        animatingMoveIndex.value = null
      }, totalTime)
      pendingTimeouts.value.push(clearTimeoutId)
    } else {
      // Not deleting active solution, just remove it
      solutions.value.splice(index, 1)
      // Adjust active index if it was after the deleted one
      if (activeSolutionIndex.value > index) {
        activeSolutionIndex.value--
      }
    }
  }

  function applyGame(game: Game, roomId?: string, gameNumber?: number) {
    // Cancel any pending replay timeouts before starting new game
    cancelPendingTimeouts()

    // Parse robots
    const newRobots: Robot[] = game.bots.map(bot => ({
      id: bot.id,
      x: bot.pos?.x ?? 0,
      y: bot.pos?.y ?? 0,
    }))
    initialRobots.value = newRobots.map(r => ({ ...r }))
    robots.value = newRobots

    // Parse walls
    vWalls.value = game.board?.vWalls.map(w => ({ x: w.x, y: w.y })) ?? []
    hWalls.value = game.board?.hWalls.map(w => ({ x: w.x, y: w.y })) ?? []

    // Parse target
    target.value = {
      robotId: game.target?.id ?? 0,
      x: game.target?.pos?.x ?? 0,
      y: game.target?.pos?.y ?? 0,
    }

    // Update room/game tracking
    currentRoomId.value = roomId ?? null
    currentGameNumber.value = gameNumber ?? 0

    // Try to restore persisted solutions if game number matches
    let restored = false
    if (roomId && gameNumber) {
      const persisted = loadPersistedSolutions(roomId)
      if (persisted && persisted.gameNumber === gameNumber) {
        // Restore solutions from localStorage
        solutions.value = persisted.solutions
        activeSolutionIndex.value = persisted.activeSolutionIndex
        restored = true

        // Replay the active solution's moves to restore robot positions
        const activeMoves = solutions.value[activeSolutionIndex.value]?.moves ?? []
        for (const move of activeMoves) {
          const robot = findRobotById(move.robotId)
          if (robot) {
            robot.x = move.toX
            robot.y = move.toY
          }
        }
        // Also restore committed moves for the active solution
        committedMoves.value = [...activeMoves]
      } else if (persisted) {
        // Different game number, clear stale persisted solutions
        clearPersistedSolutions(roomId)
      }
    }

    // Reset to fresh state if not restored
    if (!restored) {
      solutions.value = [{ moves: [], isSolved: false }]
      activeSolutionIndex.value = 0
      committedMoves.value = []
    }

    selectedRobotId.value = target.value.robotId
  }

  // Reset board to initial robot positions (for replay)
  function resetBoard() {
    robots.value = initialRobots.value.map(r => ({ ...r }))
    selectedRobotId.value = null
  }

  // Compute direction from position change
  function computeDirection(fromX: number, fromY: number, toX: number, toY: number): Direction {
    if (toX > fromX) return 'right'
    if (toX < fromX) return 'left'
    if (toY > fromY) return 'down'
    return 'up'
  }

  // Apply a single replay move (robot id + destination) with dot trail
  function applyReplayMove(robotId: number, x: number, y: number) {
    const robot = findRobotById(robotId)
    if (robot) {
      const fromX = robot.x
      const fromY = robot.y
      robot.x = x
      robot.y = y

      // Add to committedMoves after animation delay so dot appears after robot arrives
      const direction = computeDirection(fromX, fromY, x, y)
      const move: Move = { robotId, direction, fromX, fromY, toX: x, toY: y }
      const timeoutId = window.setTimeout(() => {
        committedMoves.value.push(move)
      }, ANIMATION_TIMING.REPLAY_ANIMATION)
      pendingTimeouts.value.push(timeoutId)
    }
  }

  // Clear committed moves (for replay reset)
  function clearCommittedMoves() {
    cancelPendingTimeouts()
    committedMoves.value = []
  }

  // Reset all solutions to a single empty solution
  function resetSolutions() {
    solutions.value = [{ moves: [], isSolved: false }]
    activeSolutionIndex.value = 0
    committedMoves.value = []
    animatingMoveIndex.value = null
  }

  // Reset current solution back to starting point
  function resetCurrentSolution() {
    const solution = solutions.value[activeSolutionIndex.value]
    if (!solution) return

    // Clear moves and reset solved state
    solution.moves = []
    solution.isSolved = false

    // Reset board state
    selectedRobotId.value = null
    resetBoard()
    committedMoves.value = []
    animatingMoveIndex.value = null
  }

  return {
    // State
    robots,
    initialRobots,
    vWalls,
    hWalls,
    target,
    selectedRobotId,
    moves,
    committedMoves,
    isLoading,
    error,
    solutions,
    activeSolutionIndex,
    animatingMoveIndex,
    // Computed
    moveCount,
    isSolved,
    hasAnySolvedSolution,
    canStartNewSolution,
    // Actions
    selectRobot,
    moveRobot,
    undoMove,
    applyGame,
    switchSolution,
    replaySolution,
    startNewSolution,
    deleteSolution,
    resetBoard,
    applyReplayMove,
    clearCommittedMoves,
    resetCurrentSolution,
    resetSolutions,
  }
})
