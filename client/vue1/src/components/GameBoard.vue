<script setup lang="ts">
const BOARD_SIZE = 16
const CELL_SIZE = 32

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
  </div>
</template>

<style scoped>
.board {
  display: grid;
  background: #dddddd;
  border: 4px solid #546e7a;
  position: relative;
}

.cell {
  border: 1px solid #37474f;
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
</style>
