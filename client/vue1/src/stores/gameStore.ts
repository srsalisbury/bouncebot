import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { bounceBotClient } from '../services/connectClient'
import type { Game } from '../gen/bouncebot_pb'

export type Direction = 'up' | 'down' | 'left' | 'right'

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

export const BOARD_SIZE = 16

// Color palette for robots - colors are assigned by robot ID (index)
export const ROBOT_COLORS = [
  '#e53935', // red
  '#1e88e5', // blue
  '#43a047', // green
  '#fdd835', // yellow
  '#8e24aa', // purple
  '#ff6f00', // orange
  '#00acc1', // cyan
  '#f06292', // pink
  '#5d4037', // brown
  '#546e7a', // blue-gray
]

export function getRobotColor(robotId: number): string {
  const index = robotId % ROBOT_COLORS.length
  return ROBOT_COLORS[index]!
}

export const useGameStore = defineStore('game', () => {
  // Loading state
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Initial state (stored for reset and validation)
  const initialRobots = ref<Robot[]>([])
  let initialGame: Game | null = null

  // Validation state
  const isValidating = ref(false)
  const validationResult = ref<{ isValid: boolean; message: string } | null>(null)

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

  // Move history
  const moves = ref<Move[]>([])
  // Committed moves (for history dots - delayed to match animation)
  const committedMoves = ref<Move[]>([])

  // Computed
  const moveCount = computed(() => moves.value.length)

  const isSolved = computed(() => {
    const targetRobot = robots.value.find(r => r.id === target.value.robotId)
    if (!targetRobot) return false
    return targetRobot.x === target.value.x && targetRobot.y === target.value.y
  })

  // Actions
  function selectRobot(robotId: number) {
    if (selectedRobotId.value === robotId) {
      selectedRobotId.value = null
    } else {
      selectedRobotId.value = robotId
    }
  }

  function hasWall(x: number, y: number, direction: Direction): boolean {
    // Check board edges
    if (direction === 'up' && y === 0) return true
    if (direction === 'down' && y === BOARD_SIZE - 1) return true
    if (direction === 'left' && x === 0) return true
    if (direction === 'right' && x === BOARD_SIZE - 1) return true

    // Check internal walls
    if (direction === 'right') {
      return vWalls.value.some(w => w.x === x && w.y === y)
    }
    if (direction === 'left') {
      return vWalls.value.some(w => w.x === x - 1 && w.y === y)
    }
    if (direction === 'down') {
      return hWalls.value.some(w => w.x === x && w.y === y)
    }
    if (direction === 'up') {
      return hWalls.value.some(w => w.x === x && w.y === y - 1)
    }

    return false
  }

  function isOccupied(x: number, y: number, excludeRobotId: number): boolean {
    return robots.value.some(r => r.id !== excludeRobotId && r.x === x && r.y === y)
  }

  function calculateDestination(robot: Robot, direction: Direction): { x: number; y: number } {
    let x = robot.x
    let y = robot.y

    const delta = {
      up: { dx: 0, dy: -1 },
      down: { dx: 0, dy: 1 },
      left: { dx: -1, dy: 0 },
      right: { dx: 1, dy: 0 },
    }[direction]

    // Slide until hitting a wall or another robot
    while (true) {
      if (hasWall(x, y, direction)) break

      const nextX = x + delta.dx
      const nextY = y + delta.dy

      if (isOccupied(nextX, nextY, robot.id)) break

      x = nextX
      y = nextY
    }

    return { x, y }
  }

  function moveRobot(direction: Direction) {
    if (selectedRobotId.value === null) return
    if (isSolved.value) return

    const robot = robots.value.find(r => r.id === selectedRobotId.value)
    if (!robot) return

    const destination = calculateDestination(robot, direction)

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
      moves.value.push(move)
      robot.x = destination.x
      robot.y = destination.y

      // Delay adding to committedMoves to match animation (150ms)
      setTimeout(() => {
        committedMoves.value.push(move)
      }, 150)
    }
  }

  function undoMove() {
    const lastMove = moves.value.pop()
    if (!lastMove) return

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
    validationResult.value = null
  }

  function resetPuzzle() {
    robots.value = initialRobots.value.map(r => ({ ...r }))
    moves.value = []
    committedMoves.value = []
    selectedRobotId.value = null
  }

  function applyGame(game: Game) {
    // Store initial game for validation
    initialGame = game

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
    moves.value = []
    committedMoves.value = []
    selectedRobotId.value = null
    validationResult.value = null
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

  async function checkSolution() {
    if (!initialGame) return
    if (moves.value.length === 0) return

    isValidating.value = true

    try {
      // Convert moves to BotPos format (robot id + destination position)
      const movesForServer = moves.value.map(m => ({
        id: m.robotId,
        pos: { x: m.toX, y: m.toY },
      }))

      const response = await bounceBotClient.checkSolution({
        game: initialGame,
        moves: movesForServer,
      })

      if (response.isValid) {
        validationResult.value = {
          isValid: true,
          message: `Solution verified! ${response.numMoves} moves.`,
        }
      } else {
        validationResult.value = {
          isValid: false,
          message: response.firstBadMove?.errorDescription ?? 'Invalid solution',
        }
      }
    } catch (e) {
      validationResult.value = {
        isValid: false,
        message: e instanceof Error ? e.message : 'Validation failed',
      }
    } finally {
      isValidating.value = false
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
    isValidating,
    validationResult,
    // Computed
    moveCount,
    isSolved,
    // Actions
    selectRobot,
    moveRobot,
    undoMove,
    resetPuzzle,
    loadGame,
    checkSolution,
  }
})
