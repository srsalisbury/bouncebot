import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type { Game } from '../gen/bouncebot_pb'
import { MAX_SOLUTIONS, type Direction } from '../constants'
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

  const canStartNewSolution = computed(() => {
    return solutions.value.length < MAX_SOLUTIONS
  })

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
      setTimeout(() => {
        committedMoves.value.push(move)
      }, ANIMATION_TIMING.MOVE_DELAY)

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

    const lastMove = solutionMoves.pop()!

    // Also remove from committedMoves if present
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

  // Shared function to replay moves with animation, starting after a delay
  function replayMoves(movesToReplay: Move[], startDelay: number): number {
    movesToReplay.forEach((move, i) => {
      setTimeout(() => {
        animatingMoveIndex.value = i
        const robot = findRobotById(move.robotId)
        if (robot) {
          robot.x = move.toX
          robot.y = move.toY
        }
        committedMoves.value.push(move)
      }, startDelay + i * ANIMATION_TIMING.MOVE_DELAY)
    })
    return startDelay + movesToReplay.length * ANIMATION_TIMING.MOVE_DELAY
  }

  function switchSolution(index: number) {
    if (index < 0 || index >= solutions.value.length) return
    if (index === activeSolutionIndex.value) return

    const targetMoves = solutions.value[index]!.moves

    selectedRobotId.value = null

    // Reset to initial positions immediately
    resetBoard()
    committedMoves.value = []
    animatingMoveIndex.value = null
    activeSolutionIndex.value = index

    // Wait before replaying new solution
    const resetDelay = 500

    // Replay target moves after delay
    const totalTime = replayMoves(targetMoves, resetDelay)

    // Clear highlight after replay completes
    setTimeout(() => {
      animatingMoveIndex.value = null
    }, totalTime)
  }

  function startNewSolution() {
    if (!canStartNewSolution.value) return

    // Create new empty solution
    solutions.value.push({ moves: [], isSolved: false })
    const newIndex = solutions.value.length - 1

    selectedRobotId.value = null

    // Reset to initial positions immediately
    resetBoard()
    committedMoves.value = []
    animatingMoveIndex.value = null
    activeSolutionIndex.value = newIndex
  }

  function deleteSolution(index: number) {
    // Can't delete if only one solution remains
    if (solutions.value.length <= 1) return
    if (index < 0 || index >= solutions.value.length) return

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
      const resetDelay = 500

      // Replay new active solution
      const totalTime = replayMoves(targetMoves, resetDelay)

      setTimeout(() => {
        animatingMoveIndex.value = null
      }, totalTime)
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

    selectedRobotId.value = null
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
      setTimeout(() => {
        committedMoves.value.push(move)
      }, ANIMATION_TIMING.MOVE_DELAY)
    }
  }

  // Clear committed moves (for replay reset)
  function clearCommittedMoves() {
    committedMoves.value = []
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
    canStartNewSolution,
    // Actions
    selectRobot,
    moveRobot,
    undoMove,
    applyGame,
    switchSolution,
    startNewSolution,
    deleteSolution,
    resetBoard,
    applyReplayMove,
    clearCommittedMoves,
    resetCurrentSolution,
  }
})
