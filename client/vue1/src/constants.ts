// Board constants
export const BOARD_SIZE = 16
export const CELL_SIZE = 32
export const WALL_THICKNESS = 4

// Colors - uses CSS custom property for dark mode support
export const WALL_COLOR = 'var(--wall-color, #2a2a2a)'

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

// Color palette for players - assigned by join order (index)
export const PLAYER_COLORS = [
  '#e53935', // red
  '#1e88e5', // blue
  '#43a047', // green
  '#fdd835', // yellow
  '#8e24aa', // purple
  '#fb8c00', // orange
  '#00acc1', // cyan
  '#d81b60', // pink
]

export function getPlayerColor(index: number): string {
  return PLAYER_COLORS[index % PLAYER_COLORS.length] ?? '#888888'
}

// Game limits
export const MAX_SOLUTIONS = 4

// Direction arrows for UI
export type Direction = 'up' | 'down' | 'left' | 'right'

export const DIRECTION_ARROWS: Record<Direction, string> = {
  up: '↑',
  down: '↓',
  left: '←',
  right: '→',
}
