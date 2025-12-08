import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

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
  // Hardcoded robot positions for now
  const robots = ref<Robot[]>([
    { id: 0, x: 2, y: 3 },
    { id: 1, x: 14, y: 1 },
    { id: 2, x: 5, y: 12 },
    { id: 3, x: 10, y: 8 },
  ])

  // Hardcoded walls for now
  // vWalls: vertical wall to the RIGHT of the cell at (x, y)
  // hWalls: horizontal wall BELOW the cell at (x, y)
  const vWalls = ref<Wall[]>([
    { x: 3, y: 2 },
    { x: 7, y: 5 },
    { x: 7, y: 7 },
    { x: 10, y: 8 },
    { x: 5, y: 12 },
  ])

  const hWalls = ref<Wall[]>([
    { x: 0, y: 7 },
    { x: 2, y: 3 },
    { x: 5, y: 7 },
    { x: 8, y: 10 },
    { x: 12, y: 5 },
  ])

  // Target: which robot needs to reach which position
  const target = ref<Target>({
    robotId: 0, // Red robot
    x: 7,
    y: 7,
  })

  // Selected robot state
  const selectedRobotId = ref<number | null>(null)

  // Move history
  const moves = ref<Move[]>([])

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
      moves.value.push({ robotId: robot.id, direction, fromX: robot.x, fromY: robot.y })
      robot.x = destination.x
      robot.y = destination.y
    }
  }

  function undoMove() {
    const lastMove = moves.value.pop()
    if (!lastMove) return

    const robot = robots.value.find(r => r.id === lastMove.robotId)
    if (!robot) return

    robot.x = lastMove.fromX
    robot.y = lastMove.fromY
    selectedRobotId.value = lastMove.robotId
  }

  return {
    // State
    robots,
    vWalls,
    hWalls,
    target,
    selectedRobotId,
    moves,
    // Computed
    moveCount,
    isSolved,
    // Actions
    selectRobot,
    moveRobot,
    undoMove,
  }
})
