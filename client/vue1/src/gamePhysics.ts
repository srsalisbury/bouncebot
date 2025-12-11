import { BOARD_SIZE, type Direction } from './constants'
import type { Robot, Wall } from './stores/gameStore'

/**
 * Check if there's a wall blocking movement in a direction from a position.
 */
export function hasWall(
  x: number,
  y: number,
  direction: Direction,
  vWalls: Wall[],
  hWalls: Wall[]
): boolean {
  // Check board edges
  if (direction === 'up' && y === 0) return true
  if (direction === 'down' && y === BOARD_SIZE - 1) return true
  if (direction === 'left' && x === 0) return true
  if (direction === 'right' && x === BOARD_SIZE - 1) return true

  // Check internal walls
  if (direction === 'right') {
    return vWalls.some(w => w.x === x && w.y === y)
  }
  if (direction === 'left') {
    return vWalls.some(w => w.x === x - 1 && w.y === y)
  }
  if (direction === 'down') {
    return hWalls.some(w => w.x === x && w.y === y)
  }
  if (direction === 'up') {
    return hWalls.some(w => w.x === x && w.y === y - 1)
  }

  return false
}

/**
 * Check if a position is occupied by another robot.
 */
export function isOccupied(
  x: number,
  y: number,
  excludeRobotId: number,
  robots: Robot[]
): boolean {
  return robots.some(r => r.id !== excludeRobotId && r.x === x && r.y === y)
}

/**
 * Calculate where a robot will end up when sliding in a direction.
 */
export function calculateDestination(
  robot: Robot,
  direction: Direction,
  robots: Robot[],
  vWalls: Wall[],
  hWalls: Wall[]
): { x: number; y: number } {
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
    if (hasWall(x, y, direction, vWalls, hWalls)) break

    const nextX = x + delta.dx
    const nextY = y + delta.dy

    if (isOccupied(nextX, nextY, robot.id, robots)) break

    x = nextX
    y = nextY
  }

  return { x, y }
}
