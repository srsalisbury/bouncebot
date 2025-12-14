import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { bounceBotClient } from '../services/connectClient'
import type { Game } from '../gen/bouncebot_pb'
import { BOARD_SIZE, MAX_SOLUTIONS, type Direction } from '../constants'
import { calculateDestination } from '../gamePhysics'

// Re-export for backward compatibility
export { BOARD_SIZE, MAX_SOLUTIONS, type Direction } from '../constants'
export { getRobotColor, ROBOT_COLORS } from '../constants'

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

export const useGameStore = defineStore('game', () => {
  // Loading state
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Initial state (stored for reset)
  const initialRobots = ref<Robot[]>([])


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

  // Computed
  const activeSolution = computed(() => solutions.value[activeSolutionIndex.value]!)
  const moves = computed(() => activeSolution.value.moves)
  const moveCount = computed(() => moves.value.length)

  const isSolved = computed(() => {
    const targetRobot = robots.value.find(r => r.id === target.value.robotId)
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

    const robot = robots.value.find(r => r.id === selectedRobotId.value)
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

      // Delay adding to committedMoves to match animation (150ms)
      setTimeout(() => {
        committedMoves.value.push(move)
      }, 150)

      // Check if puzzle is now solved and mark the solution
      const targetRobot = robots.value.find(r => r.id === target.value.robotId)
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

    const robot = robots.value.find(r => r.id === lastMove.robotId)
    if (!robot) return

    robot.x = lastMove.fromX
    robot.y = lastMove.fromY
    selectedRobotId.value = lastMove.robotId

    // Mark solution as unsolved since we undid a move
    activeSolution.value.isSolved = false
  }

  function resetPuzzle() {
    robots.value = initialRobots.value.map(r => ({ ...r }))
    activeSolution.value.moves.length = 0
    activeSolution.value.isSolved = false
    committedMoves.value = []
    selectedRobotId.value = null
  }

  // Shared function to unwind moves with animation, returns total time in ms
  function unwindMoves(movesToUnwind: Move[]): number {
    movesToUnwind.slice().reverse().forEach((move, i) => {
      const moveIndex = movesToUnwind.length - 1 - i
      setTimeout(() => {
        animatingMoveIndex.value = moveIndex
        // Remove dot first (before robot moves)
        const idx = committedMoves.value.indexOf(move)
        if (idx !== -1) {
          committedMoves.value.splice(idx, 1)
        }
        // Then move robot
        const robot = robots.value.find(r => r.id === move.robotId)
        if (robot) {
          robot.x = move.fromX
          robot.y = move.fromY
        }
      }, i * 150)
    })
    return movesToUnwind.length * 150
  }

  // Shared function to replay moves with animation, starting after a delay
  function replayMoves(movesToReplay: Move[], startDelay: number): number {
    movesToReplay.forEach((move, i) => {
      setTimeout(() => {
        animatingMoveIndex.value = i
        const robot = robots.value.find(r => r.id === move.robotId)
        if (robot) {
          robot.x = move.toX
          robot.y = move.toY
        }
        committedMoves.value.push(move)
      }, startDelay + i * 150)
    })
    return startDelay + movesToReplay.length * 150
  }

  function switchSolution(index: number) {
    if (index < 0 || index >= solutions.value.length) return
    if (index === activeSolutionIndex.value) return

    const currentMoves = [...activeSolution.value.moves]
    const targetMoves = solutions.value[index]!.moves

    selectedRobotId.value = null

    // Unwind current solution
    const unwindTime = unwindMoves(currentMoves)

    // Switch to target solution after unwind completes
    setTimeout(() => {
      animatingMoveIndex.value = null
      activeSolutionIndex.value = index
    }, unwindTime)

    // Replay target moves after switching
    const totalTime = replayMoves(targetMoves, unwindTime)

    // Clear highlight after replay completes
    setTimeout(() => {
      animatingMoveIndex.value = null
    }, totalTime)
  }

  function startNewSolution() {
    if (!canStartNewSolution.value) return

    const currentMoves = [...activeSolution.value.moves]

    // Create new empty solution
    solutions.value.push({ moves: [], isSolved: false })
    const newIndex = solutions.value.length - 1

    selectedRobotId.value = null

    // Unwind current solution
    const unwindTime = unwindMoves(currentMoves)

    // Switch to new solution after unwind completes
    setTimeout(() => {
      animatingMoveIndex.value = null
      activeSolutionIndex.value = newIndex
    }, unwindTime)
  }

  function deleteSolution(index: number) {
    // Can't delete if only one solution remains
    if (solutions.value.length <= 1) return
    if (index < 0 || index >= solutions.value.length) return

    // If deleting the active solution, switch to another first
    if (index === activeSolutionIndex.value) {
      // Switch to previous solution, or first if deleting index 0
      const newActiveIndex = index > 0 ? index - 1 : 1
      const currentMoves = [...activeSolution.value.moves]
      const targetMoves = solutions.value[newActiveIndex]!.moves

      selectedRobotId.value = null

      // Unwind current solution
      const unwindTime = unwindMoves(currentMoves)

      // After unwind, remove solution and replay new active
      setTimeout(() => {
        solutions.value.splice(index, 1)
        // Adjust active index after removal
        activeSolutionIndex.value = index > 0 ? index - 1 : 0
        animatingMoveIndex.value = null
      }, unwindTime)

      // Replay new active solution
      const totalTime = replayMoves(targetMoves, unwindTime)

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

  function applyGame(game: Game) {
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

    // Reset game state
    solutions.value = [{ moves: [], isSolved: false }]
    activeSolutionIndex.value = 0
    committedMoves.value = []
    selectedRobotId.value = null
  }

  async function loadGame() {
    isLoading.value = true
    error.value = null

    try {
      const game = await bounceBotClient.makeGame({ size: BOARD_SIZE })

      // Validate game response
      if (!game.board || !game.bots || game.bots.length === 0 || !game.target) {
        throw new Error('Invalid game data received from server')
      }

      applyGame(game)
    } catch (e) {
      // Format error message for user
      const message = e instanceof Error ? e.message : 'Failed to load game'
      if (message.includes('fetch') || message.includes('network') || message.includes('Failed to fetch')) {
        error.value = 'Unable to connect to server. Please check your connection and try again.'
      } else {
        error.value = message
      }
    } finally {
      isLoading.value = false
    }
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
    const robot = robots.value.find(r => r.id === robotId)
    if (robot) {
      const fromX = robot.x
      const fromY = robot.y
      robot.x = x
      robot.y = y

      // Add to committedMoves after animation delay (150ms) so dot appears after robot arrives
      const direction = computeDirection(fromX, fromY, x, y)
      const move: Move = { robotId, direction, fromX, fromY, toX: x, toY: y }
      setTimeout(() => {
        committedMoves.value.push(move)
      }, 150)
    }
  }

  // Clear committed moves (for replay reset)
  function clearCommittedMoves() {
    committedMoves.value = []
  }

  // Unwind a replay move (remove dot first, then move robot)
  function unwindReplayMove(robotId: number, x: number, y: number) {
    const robot = robots.value.find(r => r.id === robotId)
    if (robot) {
      // Remove the dot immediately (before robot moves)
      let lastIndex = -1
      for (let i = committedMoves.value.length - 1; i >= 0; i--) {
        const move = committedMoves.value[i]
        if (move && move.robotId === robotId) {
          lastIndex = i
          break
        }
      }
      if (lastIndex !== -1) {
        committedMoves.value.splice(lastIndex, 1)
      }

      // Then move the robot
      robot.x = x
      robot.y = y
    }
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
    resetPuzzle,
    loadGame,
    applyGame,
    switchSolution,
    startNewSolution,
    deleteSolution,
    resetBoard,
    applyReplayMove,
    clearCommittedMoves,
    unwindReplayMove,
  }
})
