<script setup lang="ts">
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

// Hardcoded robot positions for now
const robots = [
  { id: 0, x: 2, y: 3, color: ROBOT_COLORS.RED },
  { id: 1, x: 14, y: 1, color: ROBOT_COLORS.BLUE },
  { id: 2, x: 5, y: 12, color: ROBOT_COLORS.GREEN },
  { id: 3, x: 10, y: 8, color: ROBOT_COLORS.YELLOW },
]

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

const WALL_THICKNESS = 4

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

    <!-- Robots -->
    <div
      v-for="robot in robots"
      :key="`robot-${robot.id}`"
      class="robot"
      :style="getRobotStyle(robot)"
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
}

.wall {
  position: absolute;
  z-index: 5;
}
</style>
