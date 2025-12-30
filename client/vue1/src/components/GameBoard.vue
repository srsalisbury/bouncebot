<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useGameStore } from '../stores/gameStore'
import { BOARD_SIZE, CELL_SIZE, WALL_COLOR, WALL_THICKNESS, DIRECTION_ARROWS, getRobotColor } from '../constants'
import HowToPlayModal from './HowToPlayModal.vue'
import { useGameInput } from '../composables/useGameInput'
import { useReplay } from '../composables/useReplay'
import type { PlayerSolution } from '../gen/bouncebot_pb'
import type { Timestamp } from '@bufbuild/protobuf/wkt'

const props = defineProps<{
  onBeforeRetract?: (action: () => void) => void
  gameEnded?: boolean
  playerSolutions?: PlayerSolution[]
  getPlayerName?: (playerId: string) => string
  getPlayerColor?: (playerId: string) => string
  gameStartedAt?: Timestamp
  gameNumber?: number
  inputBlocked?: boolean
}>()

const store = useGameStore()
const showHowToPlay = ref(false)

const boardPixelSize = BOARD_SIZE * CELL_SIZE

// Replay composable
const {
  activePlayerSolutionIndex,
  replayMoveIndex,
  switchToPlayerSolution,
  startInitialReplay,
  getPlayerSolutionMoves,
} = useReplay(
  computed(() => store.initialRobots),
  {
    resetBoard: () => store.resetBoard(),
    clearCommittedMoves: () => store.clearCommittedMoves(),
    applyReplayMove: (robotId, x, y) => store.applyReplayMove(robotId, x, y),
    unwindReplayMove: (robotId, x, y) => store.unwindReplayMove(robotId, x, y),
  }
)

// Wrap actions that could retract a solution
function doUndo() {
  const currentSolution = store.solutions[store.activeSolutionIndex]
  if (currentSolution?.isSolved && props.onBeforeRetract) {
    props.onBeforeRetract(() => store.undoMove())
  } else {
    store.undoMove()
  }
}

function doDelete() {
  const solution = store.solutions[store.activeSolutionIndex]
  if (solution?.isSolved && props.onBeforeRetract) {
    props.onBeforeRetract(() => store.deleteSolution(store.activeSolutionIndex))
  } else {
    store.deleteSolution(store.activeSolutionIndex)
  }
}

// Input handling composable
useGameInput(
  {
    onMove: (direction) => store.moveRobot(direction),
    onUndo: doUndo,
    onDelete: doDelete,
    onNewSolution: () => store.startNewSolution(),
    onSelectRobot: (index) => store.selectRobot(index),
    onSwitchSolution: (delta) => store.switchSolution(store.activeSolutionIndex + delta),
    onSwitchPlayerSolution: (delta) => {
      if (props.playerSolutions) {
        switchToPlayerSolution(activePlayerSolutionIndex.value + delta, props.playerSolutions)
      }
    },
    onToggleHelp: () => { showHowToPlay.value = !showHowToPlay.value },
    onCloseHelp: () => { showHowToPlay.value = false },
  },
  {
    inputBlocked: computed(() => props.inputBlocked ?? false),
    gameEnded: computed(() => props.gameEnded ?? false),
    helpOpen: showHowToPlay,
    canStartNewSolution: computed(() => store.canStartNewSolution),
    selectedRobotId: computed(() => store.selectedRobotId),
    robotCount: computed(() => store.robots.length),
  }
)

// Format solve time relative to game start
function formatSolveTime(solvedAt?: Timestamp): string {
  if (!solvedAt || !props.gameStartedAt) return ''
  const solvedMs = Number(solvedAt.seconds) * 1000 + Math.floor(solvedAt.nanos / 1_000_000)
  const startMs = Number(props.gameStartedAt.seconds) * 1000 + Math.floor(props.gameStartedAt.nanos / 1_000_000)
  const diffSeconds = Math.floor((solvedMs - startMs) / 1000)
  if (diffSeconds < 0) return ''
  const minutes = Math.floor(diffSeconds / 60)
  const seconds = diffSeconds % 60
  return `${minutes}:${seconds.toString().padStart(2, '0')}`
}

// When game ends and we have solutions, start showing the first one
watch(() => props.gameEnded, (ended) => {
  if (ended && props.playerSolutions?.length) {
    startInitialReplay(props.playerSolutions)
  }
})

// Style helpers
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

function handleSwitchPlayerSolution(index: number) {
  if (props.playerSolutions) {
    switchToPlayerSolution(index, props.playerSolutions)
  }
}
</script>

<template>
  <div class="game-container">
    <!-- Loading state -->
    <div v-if="store.isLoading" class="loading">Loading game...</div>

    <!-- Error state -->
    <div v-else-if="store.error" class="error">
      <div class="error-icon">⚠</div>
      <div class="error-message">{{ store.error }}</div>
      <button @click="store.loadGame()">Try Again</button>
    </div>

    <!-- Game content wrapper -->
    <div v-else class="game-content">
      <!-- Board layout (grid: title on top, board and solutions below) -->
      <div class="board-layout">
        <h1 class="title">BounceBot<span v-if="props.gameNumber" class="game-number"> - Game #{{ props.gameNumber }}</span></h1>
        <!-- Board area (board + hints) -->
        <div class="board-area">
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

          <!-- Keyboard hints under board -->
          <div class="keyboard-hints">
            <template v-if="props.gameEnded">
              <kbd>Shift+←→</kbd> switch solutions
            </template>
            <template v-else>
              <kbd>1-4</kbd> select · <kbd>↑↓←→</kbd> move · <kbd>z</kbd> undo · <kbd>?</kbd> help
            </template>
          </div>
        </div>

        <!-- Player solutions panel (when game ended) -->
        <div v-if="props.gameEnded && props.playerSolutions?.length" class="solutions-panel">
          <div class="solutions-columns">
            <div
              v-for="(solution, index) in props.playerSolutions"
              :key="solution.playerId"
              class="solution-column player-solution"
              :class="{ active: index === activePlayerSolutionIndex, winner: index === 0 }"
              @click="handleSwitchPlayerSolution(index)"
            >
              <div class="player-solution-header">
                <div class="player-name-row">
                  <span class="player-dot" :style="{ backgroundColor: props.getPlayerColor?.(solution.playerId) ?? '#888888' }"></span>
                  <span class="player-name">{{ props.getPlayerName?.(solution.playerId) ?? 'Unknown' }}</span>
                </div>
                <span class="solution-moves">{{ solution.moves.length }}</span>
                <span class="solution-time">{{ formatSolveTime(solution.solvedAt) }}</span>
              </div>
              <div class="move-list">
                <div
                  v-for="(move, i) in getPlayerSolutionMoves(solution)"
                  :key="i"
                  class="move-item"
                  :class="{ animating: index === activePlayerSolutionIndex && i < replayMoveIndex }"
                >
                  <span class="move-robot" :style="{ backgroundColor: getRobotColor(move.robotId) }">
                    {{ move.robotId + 1 }}
                  </span>
                  <span class="move-arrow">{{ DIRECTION_ARROWS[move.direction] }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Normal solutions panel (during game) -->
        <div v-else-if="!props.gameEnded" class="solutions-panel">
          <div class="solutions-columns">
            <div
              v-for="(solution, index) in store.solutions"
              :key="index"
              class="solution-column"
              :class="{ active: index === store.activeSolutionIndex }"
              @click="store.switchSolution(index)"
            >
              <div class="solution-header">
                <span class="solution-moves">{{ solution.moves.length }}</span>
                <span class="solved-check" :class="{ visible: solution.isSolved }">✓</span>
              </div>
              <div class="move-list">
                <div
                  v-for="(move, i) in solution.moves"
                  :key="i"
                  class="move-item"
                  :class="{ animating: index === store.activeSolutionIndex && store.animatingMoveIndex === i }"
                >
                  <span class="move-robot" :style="{ backgroundColor: getRobotColor(move.robotId) }">
                    {{ move.robotId + 1 }}
                  </span>
                  <span class="move-arrow">{{ DIRECTION_ARROWS[move.direction] }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- How to Play modal -->
    <HowToPlayModal :show="showHowToPlay" @close="showHowToPlay = false" />
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
  flex-direction: column;
  align-items: flex-start;
}

.board-layout {
  display: grid;
  grid-template-columns: auto auto;
  grid-template-rows: auto auto;
  gap: 0.5rem 2rem;
}

.title {
  grid-column: 1;
  grid-row: 1;
  color: #42b883;
  margin: 0;
  font-size: 1.8rem;
  text-align: center;
}

.game-number {
  font-size: 1.2rem;
  font-weight: normal;
  color: #333;
}

.board-area {
  grid-column: 1;
  grid-row: 2;
}

.solutions-panel {
  grid-column: 2;
  grid-row: 2;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.solutions-columns {
  display: flex;
  flex-direction: row;
  gap: 0.5rem;
  align-items: flex-start;
  width: 280px;
}

.solution-column {
  min-width: 54px;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.4rem;
  border-radius: 6px;
  background: #dddddd;
  cursor: pointer;
  transition: background 0.15s, box-shadow 0.15s;
}

.solution-column:hover {
  background: #cccccc;
}

.solution-column.active {
  background: #dddddd;
  box-shadow: 0 0 0 2px #42b883;
}

.solution-header {
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  font-weight: 600;
  font-size: 1.2rem;
  padding-bottom: 0.4rem;
  margin-bottom: 0.25rem;
  border-bottom: 1px solid #999;
}

.solution-moves {
  color: #333;
}

.solved-check {
  position: absolute;
  right: 0;
  color: #43a047;
  opacity: 0;
}

.solved-check.visible {
  opacity: 1;
}

.move-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  max-height: 512px;
  overflow-y: auto;
}

.move-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 2px 4px;
  border-radius: 4px;
}

.move-item.animating {
  background: #42b883;
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
  color: #333;
  width: 18px;
  text-align: center;
}

.move-pos {
  font-size: 11px;
  color: #666;
  font-family: monospace;
}

/* Player solution column styles */
.solution-column.player-solution {
  min-width: 54px;
}

.player-solution-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.2rem;
  padding-bottom: 0.4rem;
  margin-bottom: 0.25rem;
  border-bottom: 1px solid #999;
}

.player-solution-header .player-name-row {
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.player-solution-header .player-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.player-solution-header .player-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: #333;
  max-width: 70px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.player-solution-header .solution-moves {
  font-size: 1.2rem;
  font-weight: 600;
  color: #333;
}

.player-solution-header .solution-time {
  font-size: 0.8rem;
  color: #666;
}

.solution-column.winner {
  background: #fff8dc;
  border: 2px solid #ffd700;
}

.solution-column.winner.active {
  box-shadow: 0 0 0 2px #42b883;
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
  border: 1px solid black;
  text-shadow: -0.5px -0.5px 0 black, 0.5px -0.5px 0 black, -0.5px 0.5px 0 black, 0.5px 0.5px 0 black;
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
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
  padding: 2rem;
  text-align: center;
}

.error-icon {
  font-size: 2.5rem;
  color: #e53935;
}

.error-message {
  color: #e53935;
  max-width: 300px;
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
