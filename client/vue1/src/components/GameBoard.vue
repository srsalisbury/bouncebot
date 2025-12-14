<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useGameStore } from '../stores/gameStore'
import { BOARD_SIZE, CELL_SIZE, WALL_COLOR, WALL_THICKNESS, DIRECTION_ARROWS, getRobotColor, type Direction } from '../constants'
import HowToPlayModal from './HowToPlayModal.vue'
import type { PlayerSolution, BotPos } from '../gen/bouncebot_pb'
import type { Timestamp } from '@bufbuild/protobuf/wkt'

const props = defineProps<{
  onBeforeRetract?: (action: () => void) => void
  gameEnded?: boolean
  playerSolutions?: PlayerSolution[]
  getPlayerName?: (playerId: string) => string
  gameStartedAt?: Timestamp
  gameNumber?: number
  inputBlocked?: boolean
}>()

const store = useGameStore()
const showHowToPlay = ref(false)

// Replay state for game-ended mode
const activePlayerSolutionIndex = ref(0)
const displayedSolutionIndex = ref(-1) // Which solution is currently shown on board (-1 = none)
const isReplaying = ref(false)
const isUnwinding = ref(false)
const replayMoveIndex = ref(0)
const unwindMoveIndex = ref(0)
const replayTimeout = ref<number | null>(null)
const replayRobotPositions = ref<Map<number, { x: number; y: number }>>(new Map())

// Compute direction from position change
function computeDirection(fromX: number, fromY: number, toX: number, toY: number): Direction | null {
  if (toX > fromX) return 'right'
  if (toX < fromX) return 'left'
  if (toY > fromY) return 'down'
  if (toY < fromY) return 'up'
  return null
}

// Get moves with computed directions for a player solution
function getPlayerSolutionMoves(solutionIndex: number) {
  if (!props.playerSolutions?.[solutionIndex]) return []
  const solution = props.playerSolutions[solutionIndex]

  // Build a map of robot positions starting from initial positions
  const positions = new Map<number, { x: number; y: number }>()
  for (const robot of store.initialRobots) {
    positions.set(robot.id, { x: robot.x, y: robot.y })
  }

  return solution.moves.map(move => {
    const robotId = move.id
    const toX = move.pos?.x ?? 0
    const toY = move.pos?.y ?? 0
    const from = positions.get(robotId) ?? { x: 0, y: 0 }
    const direction = computeDirection(from.x, from.y, toX, toY)

    // Update position for next move
    positions.set(robotId, { x: toX, y: toY })

    return {
      robotId,
      direction: direction ?? 'right' as Direction,
      toX,
      toY,
    }
  })
}

// Format solve time relative to game start (e.g., "1:23")
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

// Switch to a different player's solution with replay
function switchToPlayerSolution(index: number) {
  if (!props.playerSolutions || index < 0 || index >= props.playerSolutions.length) return
  if (index === activePlayerSolutionIndex.value && !isReplaying.value && !isUnwinding.value) return

  // Stop any current replay
  stopReplay()

  activePlayerSolutionIndex.value = index

  // Unwind current board state, then replay new solution
  unwindThenReplay()
}

function unwindThenReplay() {
  // If no solution is currently displayed, just start replay
  if (displayedSolutionIndex.value < 0 || !props.playerSolutions?.[displayedSolutionIndex.value]) {
    store.resetBoard()
    store.clearCommittedMoves()
    startReplayWithDelay()
    return
  }

  const displayedSolution = props.playerSolutions[displayedSolutionIndex.value]
  if (!displayedSolution || displayedSolution.moves.length === 0) {
    store.resetBoard()
    store.clearCommittedMoves()
    startReplayWithDelay()
    return
  }

  // Start unwinding from the last move
  isUnwinding.value = true
  unwindMoveIndex.value = displayedSolution.moves.length - 1

  // Build position map to know where each robot was before each move
  // We need to track positions through all moves to know "from" positions
  const positionHistory: Map<number, { x: number; y: number }>[] = []

  // Initial positions
  const initialPositions = new Map<number, { x: number; y: number }>()
  for (const robot of store.initialRobots) {
    initialPositions.set(robot.id, { x: robot.x, y: robot.y })
  }
  positionHistory.push(new Map(initialPositions))

  // Track positions after each move
  let currentPositions = new Map(initialPositions)
  for (const move of displayedSolution.moves) {
    currentPositions = new Map(currentPositions)
    currentPositions.set(move.id, { x: move.pos?.x ?? 0, y: move.pos?.y ?? 0 })
    positionHistory.push(new Map(currentPositions))
  }

  // Store the history for unwind steps
  replayRobotPositions.value = positionHistory[unwindMoveIndex.value] ?? new Map()

  unwindStep(displayedSolution.moves, positionHistory)
}

function unwindStep(moves: BotPos[], positionHistory: Map<number, { x: number; y: number }>[]) {
  if (unwindMoveIndex.value < 0) {
    // Done unwinding, start replay with delay
    isUnwinding.value = false
    displayedSolutionIndex.value = -1
    startReplayWithDelay()
    return
  }

  const move = moves[unwindMoveIndex.value]
  if (!move) {
    isUnwinding.value = false
    displayedSolutionIndex.value = -1
    startReplayWithDelay()
    return
  }

  // Get the position this robot was at BEFORE this move
  const beforePositions = positionHistory[unwindMoveIndex.value]
  const beforePos = beforePositions?.get(move.id)

  if (beforePos) {
    store.unwindReplayMove(move.id, beforePos.x, beforePos.y)
  }

  unwindMoveIndex.value--

  // Schedule next unwind step
  replayTimeout.value = window.setTimeout(() => {
    unwindStep(moves, positionHistory)
  }, 150) // Same speed as in-game unwind
}

function startReplayWithDelay() {
  if (!props.playerSolutions?.length) return
  const solution = props.playerSolutions[activePlayerSolutionIndex.value]
  if (!solution || !solution.moves.length) return

  isReplaying.value = true
  replayMoveIndex.value = 0

  // Initialize robot positions for tracking
  replayRobotPositions.value = new Map()
  for (const robot of store.initialRobots) {
    replayRobotPositions.value.set(robot.id, { x: robot.x, y: robot.y })
  }

  // Delay before first move
  replayTimeout.value = window.setTimeout(() => {
    stepReplay()
  }, 600)
}

function stepReplay() {
  if (!props.playerSolutions?.length) {
    isReplaying.value = false
    displayedSolutionIndex.value = activePlayerSolutionIndex.value
    return
  }
  const solution = props.playerSolutions[activePlayerSolutionIndex.value]
  if (!solution || replayMoveIndex.value >= solution.moves.length) {
    isReplaying.value = false
    displayedSolutionIndex.value = activePlayerSolutionIndex.value
    return
  }

  const move = solution.moves[replayMoveIndex.value]
  if (!move || !move.pos) {
    isReplaying.value = false
    displayedSolutionIndex.value = activePlayerSolutionIndex.value
    return
  }

  store.applyReplayMove(move.id, move.pos.x, move.pos.y)
  replayMoveIndex.value++

  // Schedule next move with 600ms delay
  replayTimeout.value = window.setTimeout(() => {
    stepReplay()
  }, 600)
}

function stopReplay() {
  if (replayTimeout.value) {
    clearTimeout(replayTimeout.value)
    replayTimeout.value = null
  }
  isReplaying.value = false
  isUnwinding.value = false
  replayMoveIndex.value = 0
}

// When game ends and we have solutions, start showing the first one
watch(() => props.gameEnded, (ended) => {
  if (ended && props.playerSolutions?.length) {
    activePlayerSolutionIndex.value = 0
    store.resetBoard()
    store.clearCommittedMoves()
    startReplayWithDelay()
  }
})

// Wrap actions that could retract a solution
function doUndo() {
  // Only need confirmation if current solution is solved and this is the last move
  const currentSolution = store.solutions[store.activeSolutionIndex]
  if (currentSolution?.isSolved && props.onBeforeRetract) {
    props.onBeforeRetract(() => store.undoMove())
  } else {
    store.undoMove()
  }
}

function doDelete(index: number) {
  const solution = store.solutions[index]
  if (solution?.isSolved && props.onBeforeRetract) {
    props.onBeforeRetract(() => store.deleteSolution(index))
  } else {
    store.deleteSolution(index)
  }
}

const boardPixelSize = BOARD_SIZE * CELL_SIZE

// Key mappings for movement directions
const MOVEMENT_KEYS: Record<string, Direction> = {
  ArrowUp: 'up', ArrowDown: 'down', ArrowLeft: 'left', ArrowRight: 'right',
  w: 'up', s: 'down', a: 'left', d: 'right',
}

function handleKeydown(event: KeyboardEvent) {
  const { key, shiftKey } = event

  // Help toggle (works in all modes)
  if (key === '?') {
    showHowToPlay.value = !showHowToPlay.value
    return
  }

  // Block all other input when a modal is open
  if (props.inputBlocked) {
    return
  }

  // Game ended mode - only allow navigating player solutions
  if (props.gameEnded) {
    if (shiftKey) {
      if (key === 'ArrowLeft') {
        event.preventDefault()
        switchToPlayerSolution(activePlayerSolutionIndex.value - 1)
        return
      }
      if (key === 'ArrowRight') {
        event.preventDefault()
        switchToPlayerSolution(activePlayerSolutionIndex.value + 1)
        return
      }
    }
    return // Ignore other keys in game-ended mode
  }

  // Normal game mode below

  // Undo
  if (key === 'z' || key === 'u' || key === 'Escape') {
    doUndo()
    return
  }

  // Shift commands
  if (shiftKey) {
    if (key === 'R') { store.resetPuzzle(); return }
    if (key === 'D') { doDelete(store.activeSolutionIndex); return }
    if (key === 'ArrowLeft') { event.preventDefault(); store.switchSolution(store.activeSolutionIndex - 1); return }
    if (key === 'ArrowRight') { event.preventDefault(); store.switchSolution(store.activeSolutionIndex + 1); return }
  }

  // New solution
  if ((key === 'n' || key === '+') && store.canStartNewSolution) {
    store.startNewSolution()
    return
  }

  // Robot selection (1-4)
  const num = parseInt(key)
  if (num >= 1 && num <= store.robots.length) {
    store.selectRobot(num - 1)
    return
  }

  // Movement
  const direction = MOVEMENT_KEYS[key]
  if (direction && store.selectedRobotId !== null) {
    event.preventDefault()
    store.moveRobot(direction)
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  stopReplay()
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
              @click="switchToPlayerSolution(index)"
            >
              <div class="player-solution-header">
                <span class="player-name">{{ props.getPlayerName?.(solution.playerId) ?? 'Unknown' }}</span>
                <span class="solution-moves">{{ solution.moves.length }}</span>
                <span class="solution-time">{{ formatSolveTime(solution.solvedAt) }}</span>
              </div>
              <div class="move-list">
                <div
                  v-for="(move, i) in getPlayerSolutionMoves(index)"
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
