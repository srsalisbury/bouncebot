<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const BOARD_SIZE = 16
const CELL_SIZE = 32
const WALL_COLOR = '#8b4513'

const boardPixelSize = BOARD_SIZE * CELL_SIZE

const ROBOT_COLORS = {
  RED: '#e53935',
  BLUE: '#1e88e5',
  GREEN: '#43a047',
  YELLOW: '#fdd835',
}

// Hardcoded robot positions for now (reactive for movement)
const robots = ref([
  { id: 0, x: 2, y: 3, color: ROBOT_COLORS.RED },
  { id: 1, x: 14, y: 1, color: ROBOT_COLORS.BLUE },
  { id: 2, x: 5, y: 12, color: ROBOT_COLORS.GREEN },
  { id: 3, x: 10, y: 8, color: ROBOT_COLORS.YELLOW },
])

// Hardcoded walls for now
// vWalls: vertical wall to the RIGHT of the cell at (x, y)
// hWalls: horizontal wall BELOW the cell at (x, y)
const vWalls = [
  { x: 3, y: 2 },
  { x: 7, y: 5 },
  { x: 10, y: 8 },
  { x: 5, y: 12 },
]

const hWalls = [
  { x: 2, y: 3 },
  { x: 5, y: 7 },
  { x: 8, y: 10 },
  { x: 12, y: 5 },
]

// Target: which robot needs to reach which position
const target = {
  robotId: 0,  // Red robot
  x: 7,
  y: 7,
}

const WALL_THICKNESS = 4

// Selected robot state
const selectedRobotId = ref<number | null>(null)

function selectRobot(robotId: number) {
  if (selectedRobotId.value === robotId) {
    selectedRobotId.value = null
  } else {
    selectedRobotId.value = robotId
  }
}

type Direction = 'up' | 'down' | 'left' | 'right'

// Check if there's a wall blocking movement from a position in a direction
function hasWall(x: number, y: number, direction: Direction): boolean {
  // Check board edges
  if (direction === 'up' && y === 0) return true
  if (direction === 'down' && y === BOARD_SIZE - 1) return true
  if (direction === 'left' && x === 0) return true
  if (direction === 'right' && x === BOARD_SIZE - 1) return true

  // Check internal walls
  // vWalls: wall is to the RIGHT of the cell at position
  // hWalls: wall is BELOW the cell at position
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

// Check if a position is occupied by another robot
function isOccupied(x: number, y: number, excludeRobotId: number): boolean {
  return robots.value.some(r => r.id !== excludeRobotId && r.x === x && r.y === y)
}

// Calculate where a robot would stop if moved in a direction (sliding)
function calculateDestination(robot: { id: number; x: number; y: number }, direction: Direction): { x: number; y: number } {
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

  const robot = robots.value.find(r => r.id === selectedRobotId.value)
  if (!robot) return

  const destination = calculateDestination(robot, direction)
  robot.x = destination.x
  robot.y = destination.y
}

function handleKeydown(event: KeyboardEvent) {
  // Number keys for robot selection
  const num = parseInt(event.key)
  if (num >= 1 && num <= robots.value.length) {
    selectRobot(num - 1)
    return
  }

  // Arrow keys for movement
  const keyMap: Record<string, Direction> = {
    ArrowUp: 'up',
    ArrowDown: 'down',
    ArrowLeft: 'left',
    ArrowRight: 'right',
    w: 'up',
    s: 'down',
    a: 'left',
    d: 'right',
  }

  const direction = keyMap[event.key]
  if (direction && selectedRobotId.value !== null) {
    event.preventDefault()
    moveRobot(direction)
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

function getVWallStyle(wall: { x: number; y: number }) {
  return {
    left: `${(wall.x + 1) * CELL_SIZE - WALL_THICKNESS / 2}px`,
    top: `${wall.y * CELL_SIZE}px`,
    height: `${CELL_SIZE}px`,
    width: `${WALL_THICKNESS}px`,
    backgroundColor: WALL_COLOR,
  }
}

function getHWallStyle(wall: { x: number; y: number }) {
  return {
    left: `${wall.x * CELL_SIZE}px`,
    top: `${(wall.y + 1) * CELL_SIZE - WALL_THICKNESS / 2}px`,
    width: `${CELL_SIZE}px`,
    height: `${WALL_THICKNESS}px`,
    backgroundColor: WALL_COLOR,
  }
}

function getRobotStyle(robot: { x: number; y: number; color: string }) {
  const padding = CELL_SIZE * 0.1
  return {
    left: `${robot.x * CELL_SIZE + padding}px`,
    top: `${robot.y * CELL_SIZE + padding}px`,
    width: `${CELL_SIZE - padding * 2}px`,
    height: `${CELL_SIZE - padding * 2}px`,
    backgroundColor: robot.color,
  }
}

function getTargetContainerStyle() {
  return {
    left: `${target.x * CELL_SIZE}px`,
    top: `${target.y * CELL_SIZE}px`,
    width: `${CELL_SIZE}px`,
    height: `${CELL_SIZE}px`,
  }
}

function getTargetBackgroundStyle() {
  const targetRobot = robots.value.find(r => r.id === target.robotId)
  const color = targetRobot?.color ?? '#ffffff'
  const robotPadding = CELL_SIZE * 0.1
  const holeSize = CELL_SIZE - robotPadding * 2
  return {
    width: '100%',
    height: '100%',
    backgroundColor: color,
    maskImage: `radial-gradient(circle at center, transparent ${holeSize / 2}px, black ${holeSize / 2}px)`,
    WebkitMaskImage: `radial-gradient(circle at center, transparent ${holeSize / 2}px, black ${holeSize / 2}px)`,
  }
}
</script>

<template>
  <div
    class="board"
    :style="{
      width: `${boardPixelSize}px`,
      height: `${boardPixelSize}px`,
      gridTemplateColumns: `repeat(${BOARD_SIZE}, ${CELL_SIZE}px)`,
      gridTemplateRows: `repeat(${BOARD_SIZE}, ${CELL_SIZE}px)`,
      borderColor: WALL_COLOR,
    }"
  >
    <div
      v-for="i in BOARD_SIZE * BOARD_SIZE"
      :key="i"
      class="cell"
    />

    <!-- Target marker -->
    <div class="target-container" :style="getTargetContainerStyle()">
      <div class="target-background" :style="getTargetBackgroundStyle()" />
      <span class="target-number">{{ target.robotId + 1 }}</span>
    </div>

    <!-- Robots -->
    <div
      v-for="robot in robots"
      :key="`robot-${robot.id}`"
      class="robot"
      :class="{ selected: selectedRobotId === robot.id }"
      :style="getRobotStyle(robot)"
      @click="selectRobot(robot.id)"
    >
      {{ robot.id + 1 }}
    </div>

    <!-- Vertical walls -->
    <div
      v-for="(wall, i) in vWalls"
      :key="`vwall-${i}`"
      class="wall"
      :style="getVWallStyle(wall)"
    />

    <!-- Horizontal walls -->
    <div
      v-for="(wall, i) in hWalls"
      :key="`hwall-${i}`"
      class="wall"
      :style="getHWallStyle(wall)"
    />
  </div>
</template>

<style scoped>
.board {
  display: grid;
  background: #dddddd;
  border: 4px solid;
  position: relative;
}

.cell {
  border: 0.5px solid #aaaaaa;
  box-sizing: border-box;
}

.robot {
  position: absolute;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  color: white;
  font-size: 14px;
  user-select: none;
  cursor: pointer;
  transition: transform 0.1s, box-shadow 0.1s;
}

.robot:hover {
  transform: scale(1.05);
}

.robot.selected {
  box-shadow: 0 0 0 3px white, 0 0 0 4px black, 0 0 10px 3px rgba(255, 255, 255, 0.5);
  transform: scale(1.1);
}

.wall {
  position: absolute;
  z-index: 5;
}

.target-container {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
}

.target-background {
  position: absolute;
  border-radius: 4px;
}

.target-number {
  position: relative;
  font-weight: bold;
  font-size: 14px;
  color: black;
}
</style>
