<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { useGameStore, BOARD_SIZE, getRobotColor, type Direction } from '../stores/gameStore'

const store = useGameStore()

// Auto-validate when puzzle is solved
watch(() => store.isSolved, (solved) => {
  if (solved && !store.validationResult) {
    store.checkSolution()
  }
})

const CELL_SIZE = 32
const WALL_COLOR = '#8b4513'
const WALL_THICKNESS = 4

const boardPixelSize = BOARD_SIZE * CELL_SIZE

const DIRECTION_ARROWS: Record<Direction, string> = {
  up: '↑',
  down: '↓',
  left: '←',
  right: '→',
}

function handleKeydown(event: KeyboardEvent) {
  // Undo with z, u, or Escape
  if (event.key === 'z' || event.key === 'u' || event.key === 'Escape') {
    store.undoMove()
    return
  }

  // Reset with R (shift+r)
  if (event.key === 'R') {
    store.resetPuzzle()
    return
  }

  // Number keys for robot selection
  const num = parseInt(event.key)
  if (num >= 1 && num <= store.robots.length) {
    store.selectRobot(num - 1)
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
  if (direction && store.selectedRobotId !== null) {
    event.preventDefault()
    store.moveRobot(direction)
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
  store.loadGame()
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

function getRobotStyle(robot: { id: number; x: number; y: number }) {
  const padding = CELL_SIZE * 0.1
  return {
    left: `${robot.x * CELL_SIZE + padding}px`,
    top: `${robot.y * CELL_SIZE + padding}px`,
    width: `${CELL_SIZE - padding * 2}px`,
    height: `${CELL_SIZE - padding * 2}px`,
    backgroundColor: getRobotColor(robot.id),
  }
}

function getTargetContainerStyle() {
  return {
    left: `${store.target.x * CELL_SIZE}px`,
    top: `${store.target.y * CELL_SIZE}px`,
    width: `${CELL_SIZE}px`,
    height: `${CELL_SIZE}px`,
  }
}

function getTargetBackgroundStyle() {
  const color = getRobotColor(store.target.robotId)
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

function getHistoryDotStyle(x: number, y: number, robotId: number, isStart: boolean) {
  const size = isStart ? CELL_SIZE * 0.35 : CELL_SIZE * 0.25
  const offset = (CELL_SIZE - size) / 2
  return {
    left: `${x * CELL_SIZE + offset}px`,
    top: `${y * CELL_SIZE + offset}px`,
    width: `${size}px`,
    height: `${size}px`,
    backgroundColor: getRobotColor(robotId),
  }
}
</script>

<template>
  <div class="game-container">
    <!-- Loading state -->
    <div v-if="store.isLoading" class="loading">Loading game...</div>

    <!-- Error state -->
    <div v-else-if="store.error" class="error">
      {{ store.error }}
      <button @click="store.loadGame()">Retry</button>
    </div>

    <!-- Game content wrapper -->
    <div v-else class="game-content">
      <!-- Game board -->
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
      <span class="target-number">{{ store.target.robotId + 1 }}</span>
    </div>

    <!-- Robot starting positions (large dots) -->
    <div
      v-for="robot in store.initialRobots"
      :key="`start-${robot.id}`"
      class="history-dot start-dot"
      :style="getHistoryDotStyle(robot.x, robot.y, robot.id, true)"
    />

    <!-- Robot move history (small dots at destinations) -->
    <div
      v-for="(move, i) in store.committedMoves"
      :key="`move-${i}`"
      class="history-dot"
      :style="getHistoryDotStyle(move.toX, move.toY, move.robotId, false)"
    />

    <!-- Robots -->
    <div
      v-for="robot in store.robots"
      :key="`robot-${robot.id}`"
      class="robot"
      :class="{ selected: store.selectedRobotId === robot.id }"
      :style="getRobotStyle(robot)"
      @click="store.selectRobot(robot.id)"
    >
      {{ robot.id + 1 }}
    </div>

    <!-- Vertical walls -->
    <div
      v-for="(wall, i) in store.vWalls"
      :key="`vwall-${i}`"
      class="wall"
      :style="getVWallStyle(wall)"
    />

    <!-- Horizontal walls -->
    <div
      v-for="(wall, i) in store.hWalls"
      :key="`hwall-${i}`"
      class="wall"
      :style="getHWallStyle(wall)"
    />
    </div>

      <!-- Move history panel -->
      <div class="move-panel">
        <button class="new-game-btn" @click="store.loadGame()">New Game</button>
        <div class="move-count">
          Moves: {{ store.moveCount }}
          <span v-if="store.isSolved" class="solved-label">Solved</span>
        </div>
        <!-- Validation result -->
        <div v-if="store.isValidating" class="validation-status validating">
          Validating...
        </div>
        <div
          v-else-if="store.validationResult"
          class="validation-status"
          :class="{ valid: store.validationResult.isValid, invalid: !store.validationResult.isValid }"
        >
          {{ store.validationResult.message }}
        </div>
        <div class="move-list">
          <div v-for="(move, i) in store.moves" :key="i" class="move-item">
            <span class="move-robot" :style="{ backgroundColor: getRobotColor(move.robotId) }">
              {{ move.robotId + 1 }}
            </span>
            <span class="move-arrow">{{ DIRECTION_ARROWS[move.direction] }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Keyboard hints under board -->
    <div v-if="!store.isLoading && !store.error" class="keyboard-hints">
      <kbd>1-4</kbd> select · <kbd>↑↓←→</kbd> move · <kbd>z</kbd> undo · <kbd>R</kbd> reset
    </div>
  </div>
</template>

<style scoped>
.game-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.game-content {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  gap: 2rem;
}

.move-panel {
  min-width: 140px;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.new-game-btn {
  padding: 0.6rem 1.2rem;
  cursor: pointer;
  font-size: 0.95rem;
  font-weight: 500;
  background: #42b883;
  color: white;
  border: none;
  border-radius: 6px;
  transition: background 0.15s, transform 0.1s;
}

.new-game-btn:hover {
  background: #3aa876;
  transform: translateY(-1px);
}

.new-game-btn:active {
  transform: translateY(0);
}

.move-count {
  font-size: 1.1rem;
  font-weight: 600;
}

.solved-label {
  color: #43a047;
  margin-left: 0.5rem;
}

.validation-status {
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
}

.validation-status.validating {
  color: #888;
}

.validation-status.valid {
  color: #43a047;
}

.validation-status.invalid {
  color: #e53935;
}

.move-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 512px;
  overflow-y: auto;
}

.move-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.move-robot {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 12px;
  color: white;
}

.move-arrow {
  font-size: 18px;
}

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

.history-dot {
  position: absolute;
  border-radius: 50%;
  opacity: 0.8;
  z-index: 1;
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
  transition: left 0.15s ease-out, top 0.15s ease-out, transform 0.1s, box-shadow 0.1s;
  z-index: 2;
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

.loading {
  font-size: 1.1rem;
  color: #888;
  padding: 2rem;
}

.error {
  color: #e53935;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: flex-start;
  padding: 2rem;
}

.error button {
  padding: 0.6rem 1.2rem;
  cursor: pointer;
  font-size: 0.95rem;
  background: #e53935;
  color: white;
  border: none;
  border-radius: 6px;
}

.error button:hover {
  background: #c62828;
}

.keyboard-hints {
  font-size: 0.8rem;
  color: #888;
  margin-top: 0.5rem;
}

.keyboard-hints kbd {
  background: #333;
  color: #fff;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: inherit;
  font-size: 0.75rem;
}
</style>
