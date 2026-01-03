<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useGameStore } from '../stores/gameStore'
import { BOARD_SIZE, WALL_COLOR, DIRECTION_ARROWS, getRobotColor } from '../constants'
import HowToPlayModal from './HowToPlayModal.vue'
import SolutionsDrawer from './SolutionsDrawer.vue'
import PlayerSolutionsDrawer from './PlayerSolutionsDrawer.vue'
import { useGameInput } from '../composables/useGameInput'
import { useReplay } from '../composables/useReplay'
import { useSwipe } from '../composables/useSwipe'
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
const boardRef = ref<HTMLElement | null>(null)

// Percentage-based sizing for responsive board
const CELL_PERCENT = 100 / BOARD_SIZE  // 6.25%

// Replay composable
const {
  activePlayerSolutionIndex,
  replayMoveIndex,
  switchToPlayerSolution,
  startInitialReplay,
  stopReplay,
  getPlayerSolutionMoves,
} = useReplay(
  computed(() => store.initialRobots),
  {
    resetBoard: () => store.resetBoard(),
    clearCommittedMoves: () => store.clearCommittedMoves(),
    applyReplayMove: (robotId, x, y) => store.applyReplayMove(robotId, x, y),
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

function doReset() {
  const currentSolution = store.solutions[store.activeSolutionIndex]
  if (currentSolution?.isSolved && props.onBeforeRetract) {
    props.onBeforeRetract(() => store.resetCurrentSolution())
  } else {
    store.resetCurrentSolution()
  }
}

// Press-and-hold detection for reset functionality
let undoHoldTimer: ReturnType<typeof setTimeout> | null = null
let undoDidReset = false
let undoHoldActive = false

function onUndoPointerDown() {
  // Prevent double-triggering from both touch and pointer events
  if (undoHoldActive) return
  undoHoldActive = true
  undoDidReset = false
  undoHoldTimer = setTimeout(() => {
    undoDidReset = true
    doReset()
  }, 1000)
}

function onUndoPointerUp() {
  if (!undoHoldActive) return
  undoHoldActive = false
  if (undoHoldTimer) {
    clearTimeout(undoHoldTimer)
    undoHoldTimer = null
  }
  // If we didn't reset, do a normal undo
  if (!undoDidReset) {
    doUndo()
  }
}

function onUndoPointerCancel() {
  undoHoldActive = false
  if (undoHoldTimer) {
    clearTimeout(undoHoldTimer)
    undoHoldTimer = null
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

// Track robot under touch start for swipe-to-select
let swipeStartRobotId: number | null = null

// Swipe gesture handling for mobile
useSwipe({
  target: boardRef,
  onSwipeStart: ({ relativeX, relativeY }) => {
    swipeStartRobotId = null
    if (props.inputBlocked || props.gameEnded) return
    // Convert normalized position to cell coordinates
    const cellX = Math.floor(relativeX * BOARD_SIZE)
    const cellY = Math.floor(relativeY * BOARD_SIZE)
    // Record which robot (if any) the touch started on
    const robotAtCell = store.robots.find(r => r.x === cellX && r.y === cellY)
    if (robotAtCell) {
      swipeStartRobotId = robotAtCell.id
    }
  },
  onSwipe: (direction) => {
    if (props.inputBlocked || props.gameEnded) return
    // If swipe started on a robot, select it (if not already selected)
    if (swipeStartRobotId !== null && store.selectedRobotId !== swipeStartRobotId) {
      store.selectRobot(swipeStartRobotId)
    }
    // Move the selected robot
    if (store.selectedRobotId !== null) {
      store.moveRobot(direction)
    }
  },
  enabled: computed(() => !props.inputBlocked && !props.gameEnded),
})

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

// When game ends, start showing solutions; when new round starts, stop replay
watch(() => props.gameEnded, (ended) => {
  if (ended && props.playerSolutions?.length) {
    startInitialReplay(props.playerSolutions)
  } else if (!ended) {
    stopReplay()
  }
})

// Prevent double-tap zoom on iOS/iPad
let lastTouchEnd = 0
function preventDoubleTapZoom(event: TouchEvent) {
  const now = Date.now()
  if (now - lastTouchEnd <= 300) {
    event.preventDefault()
  }
  lastTouchEnd = now
}

onMounted(() => {
  if (boardRef.value) {
    boardRef.value.addEventListener('touchend', preventDoubleTapZoom, { passive: false })
  }
})

onUnmounted(() => {
  if (boardRef.value) {
    boardRef.value.removeEventListener('touchend', preventDoubleTapZoom)
  }
})

// Style helpers - all use percentages for responsive sizing
const WALL_THICKNESS_PERCENT = 0.78  // ~4px at 512px board size

function getVWallStyle(wall: { x: number; y: number }) {
  // Extend by half thickness on each end to fill corner gaps
  const extension = WALL_THICKNESS_PERCENT / 2
  return {
    left: `calc(${(wall.x + 1) * CELL_PERCENT}% - ${WALL_THICKNESS_PERCENT / 2}%)`,
    top: `${wall.y * CELL_PERCENT - extension}%`,
    height: `${CELL_PERCENT + WALL_THICKNESS_PERCENT}%`,
    width: `${WALL_THICKNESS_PERCENT}%`,
    backgroundColor: WALL_COLOR,
  }
}

function getHWallStyle(wall: { x: number; y: number }) {
  // Extend by half thickness on each end to fill corner gaps
  const extension = WALL_THICKNESS_PERCENT / 2
  return {
    left: `${wall.x * CELL_PERCENT - extension}%`,
    top: `calc(${(wall.y + 1) * CELL_PERCENT}% - ${WALL_THICKNESS_PERCENT / 2}%)`,
    width: `${CELL_PERCENT + WALL_THICKNESS_PERCENT}%`,
    height: `${WALL_THICKNESS_PERCENT}%`,
    backgroundColor: WALL_COLOR,
  }
}

function getRobotStyle(robot: { id: number; x: number; y: number }) {
  return {
    left: `${(robot.x + 0.5) * CELL_PERCENT}%`,
    top: `${(robot.y + 0.5) * CELL_PERCENT}%`,
    width: `${CELL_PERCENT * 0.8}%`,
    height: `${CELL_PERCENT * 0.8}%`,
    transform: 'translate(-50%, -50%)',
    backgroundColor: getRobotColor(robot.id),
  }
}

function getTargetContainerStyle() {
  return {
    left: `${(store.target.x + 0.5) * CELL_PERCENT}%`,
    top: `${(store.target.y + 0.5) * CELL_PERCENT}%`,
    width: `${CELL_PERCENT}%`,
    height: `${CELL_PERCENT}%`,
    transform: 'translate(-50%, -50%)',
  }
}

function getTargetBackgroundStyle() {
  const color = getRobotColor(store.target.robotId)
  // Use closest-side sizing so 80% = 80% of half-width = 40% of cell = robot radius
  return {
    width: '100%',
    height: '100%',
    backgroundColor: color,
    maskImage: `radial-gradient(circle closest-side at center, transparent 80%, black 80%)`,
    WebkitMaskImage: `radial-gradient(circle closest-side at center, transparent 80%, black 80%)`,
  }
}

function getHistoryDotStyle(x: number, y: number, robotId: number, isStart: boolean) {
  const sizePercent = isStart ? CELL_PERCENT * 0.35 : CELL_PERCENT * 0.25
  const offsetPercent = (CELL_PERCENT - sizePercent) / 2
  return {
    left: `${x * CELL_PERCENT + offsetPercent}%`,
    top: `${y * CELL_PERCENT + offsetPercent}%`,
    width: `${sizePercent}%`,
    height: `${sizePercent}%`,
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
    <!-- Game content wrapper -->
    <div class="game-content">
      <!-- Header slot for room controls -->
      <slot name="header"></slot>
      <!-- Board layout (grid: title on top, board and solutions below) -->
      <div class="board-layout">
        <h1 class="title">BounceBot<span v-if="props.gameNumber" class="game-number"> - Game #{{ props.gameNumber }}</span></h1>
        <!-- Board area (board + hints) -->
        <div class="board-area">
          <!-- Game board -->
          <div
            ref="boardRef"
            class="board"
            :style="{ borderColor: WALL_COLOR }"
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
          <!-- Action buttons under solutions (desktop) -->
          <div class="action-buttons desktop-actions">
            <button
              class="action-btn undo-btn"
              @pointerdown="onUndoPointerDown"
              @pointerup="onUndoPointerUp"
              @pointercancel="onUndoPointerCancel"
              @pointerleave="onUndoPointerCancel"
              @touchstart.prevent="onUndoPointerDown"
              @touchend.prevent="onUndoPointerUp"
              @touchcancel="onUndoPointerCancel"
              @contextmenu.prevent
            >Undo Move</button>
            <button
              class="action-btn new-solution-btn"
              :disabled="!store.canStartNewSolution"
              @click="store.startNewSolution()"
            >
              New Solution
            </button>
          </div>
        </div>
      </div>

      <!-- Action buttons under board (mobile) -->
      <div v-if="!props.gameEnded" class="action-buttons mobile-actions">
        <button
          class="action-btn undo-btn"
          @pointerdown="onUndoPointerDown"
          @pointerup="onUndoPointerUp"
          @pointercancel="onUndoPointerCancel"
          @pointerleave="onUndoPointerCancel"
          @touchstart.prevent="onUndoPointerDown"
          @touchend.prevent="onUndoPointerUp"
          @touchcancel="onUndoPointerCancel"
          @contextmenu.prevent
        >Undo Move</button>
        <button
          class="action-btn new-solution-btn"
          :disabled="!store.canStartNewSolution"
          @click="store.startNewSolution()"
        >
          New Solution
        </button>
      </div>
    </div>

    <!-- Mobile solutions drawer (only during gameplay, hidden on desktop) -->
    <SolutionsDrawer
      v-if="!props.gameEnded"
      class="mobile-drawer"
    />

    <!-- Mobile player solutions drawer (only after game ends, hidden on desktop) -->
    <PlayerSolutionsDrawer
      v-if="props.gameEnded && props.playerSolutions?.length"
      class="mobile-drawer"
      :player-solutions="props.playerSolutions"
      :active-index="activePlayerSolutionIndex"
      :replay-move-index="replayMoveIndex"
      :get-player-name="props.getPlayerName ?? (() => 'Unknown')"
      :get-player-color="props.getPlayerColor ?? (() => '#888888')"
      :get-player-solution-moves="getPlayerSolutionMoves"
      :game-started-at="props.gameStartedAt"
      @switch-solution="(index) => handleSwitchPlayerSolution(index)"
    />

    <!-- How to Play modal -->
    <HowToPlayModal :show="showHowToPlay" @close="showHowToPlay = false" />
  </div>
</template>

<style scoped>
.game-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.game-content {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 0.5rem;
  flex: 1;
  max-width: calc(100vw - 2rem);
}

.board-layout {
  display: grid;
  grid-template-columns: 1fr auto;
  grid-template-rows: auto auto auto;
  gap: 0.5rem 2rem;
  align-items: stretch;
  max-width: calc(100vw - 2rem);
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
}

/* Action buttons */
.action-buttons {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: auto;
}

.action-btn {
  padding: 0.5rem 1rem;
  background: #333;
  border: none;
  border-radius: 6px;
  color: #fff;
  font-size: 0.9rem;
  cursor: pointer;
  min-height: 44px;
  -webkit-user-select: none;
  user-select: none;
  -webkit-touch-callout: none;
}

.action-btn:hover:not(:disabled) {
  background: #444;
}

.action-btn.undo-btn {
  background: #c62828;
  /* Prevent iOS long-press behaviors that cancel pointer events */
  touch-action: manipulation;
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  user-select: none;
}

.action-btn.undo-btn:hover:not(:disabled) {
  background: #d32f2f;
}

.action-btn.new-solution-btn {
  background: #2e7d32;
}

.action-btn.new-solution-btn:hover:not(:disabled) {
  background: #388e3c;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Desktop: show under solutions, hide mobile actions */
.desktop-actions {
  margin-top: auto;
}

.mobile-actions {
  display: none;
}

/* Vertical layout when:
   - Screen aspect ratio ≤ 6/5 (1.2) - screen is too square/tall for side-by-side layout, OR
   - Screen is narrow (≤1050px) regardless of aspect ratio
*/
@media (max-aspect-ratio: 6/5), (max-width: 1050px) {
  .game-container {
    width: 100%;
  }

  .game-content {
    width: 100%;
    max-width: none;
  }

  .board-layout {
    grid-template-columns: 1fr;
    grid-template-rows: auto auto auto;
    gap: 0.5rem;
    width: 100%;
    max-width: none;
  }

  .title {
    grid-column: 1;
    font-size: 1.4rem;
  }

  .board-area {
    grid-column: 1;
    grid-row: 2;
    width: 100%;
  }

  .solutions-panel {
    grid-column: 1;
    grid-row: 3;
    width: 100%;
    display: none;
  }

  .solutions-columns {
    width: 100%;
    overflow-x: auto;
    justify-content: center;
  }

  .keyboard-hints {
    display: none;
  }

  .desktop-actions {
    display: none;
  }

  .mobile-actions {
    display: flex;
    justify-content: center;
    margin-top: 0.5rem;
    margin-bottom: 4.5rem; /* Space above drawer */
    width: 100%;
  }

}

/* Mobile drawer - hidden on desktop */
.mobile-drawer {
  display: none;
}

@media (max-aspect-ratio: 6/5), (max-width: 1050px) {
  .mobile-drawer {
    display: block;
  }
}

.solution-column {
  width: 5rem;
  flex-shrink: 0;
  box-sizing: border-box;
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
  gap: 0.25rem;
  max-height: 45rem;
  overflow-y: auto;
  overflow-x: hidden;
}

.move-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.125rem 0.25rem;
  border-radius: 4px;
}

.move-item.animating {
  background: #42b883;
}

.move-robot {
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 0.75rem;
  color: white;
  border: 0.5px solid black;
  text-shadow: -0.3px -0.3px 0 black, 0.3px -0.3px 0 black, -0.3px 0.3px 0 black, 0.3px 0.3px 0 black;
}

.move-arrow {
  font-size: 1.125rem;
  color: #333;
  width: 1.125rem;
  text-align: center;
}

.move-pos {
  font-size: 0.7rem;
  color: #666;
  font-family: monospace;
}

/* Player solution column styles */
.solution-column.player-solution {
  min-width: 3.375rem;
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
  width: 0.625rem;
  height: 0.625rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.player-solution-header .player-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: #333;
  max-width: 4.5rem;
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
  --wall-color: #2a2a2a;
  display: grid;
  grid-template-columns: repeat(16, 1fr);
  grid-template-rows: repeat(16, 1fr);
  background: #dddddd;
  position: relative;
  /* Account for: padding(2rem) + gap(2rem) + solutions(21rem) = 25rem horizontal */
  /* Account for: padding(2rem) + header(2.5rem) + title(2rem) + gaps(1rem) + hints(1.5rem) ≈ 12rem vertical */
  width: min(calc(100vw - 25rem), calc(100dvh - 12rem), calc(100vh - 12rem));
  aspect-ratio: 1;
  container-type: inline-size;
  /* Prevent unwanted touch behaviors */
  touch-action: none;
  user-select: none;
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  /* Border for iPad Safari and other browsers where ::before with cqw may not work */
  border: 4px solid var(--wall-color);
  box-sizing: border-box;
}

/* Enhanced border using cqw units - overlays the fallback border on supporting browsers */
.board::before {
  content: '';
  position: absolute;
  /* Position centered on edge: offset by half wall thickness */
  inset: -0.39cqw;
  border: 0.78cqw solid var(--wall-color);
  pointer-events: none;
  z-index: 10;
}

@media (max-aspect-ratio: 6/5), (max-width: 1050px) {
  .board {
    /* Fill width in mobile, but constrain by height. Account for header, title, buttons, drawer */
    width: min(calc(100% - 0.5rem), calc(100dvh - 17rem), calc(100vh - 17rem));
    margin: 0 auto;
  }
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

.history-dot.start-dot {
  border-radius: 2px;
  transform: rotate(45deg);
}

.robot {
  position: absolute;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  color: white;
  font-size: 3cqw;
  user-select: none;
  cursor: pointer;
  transition: left 0.15s ease-out, top 0.15s ease-out, transform 0.1s, box-shadow 0.1s;
  z-index: 2;
  border: 1px solid black;
  text-shadow: -0.5px -0.5px 0 black, 0.5px -0.5px 0 black, -0.5px 0.5px 0 black, 0.5px 0.5px 0 black;
}

.robot:hover {
  transform: translate(-50%, -50%) scale(1.05);
}

.robot.selected {
  /* Fallback for iPad Safari where cqw may not work */
  box-shadow: 0 0 0 3px white, 0 0 0 4px black, 0 0 8px 3px rgba(255, 255, 255, 0.5);
  /* Enhanced version using container query units */
  box-shadow: 0 0 0 0.5cqw white, 0 0 0 0.625cqw black, 0 0 1.5cqw 0.5cqw rgba(255, 255, 255, 0.5);
  transform: translate(-50%, -50%) scale(1.1);
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
  font-size: 3cqw;
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
  max-width: 18.75rem;
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
  grid-column: 1;
  grid-row: 3;
  font-size: 0.8rem;
  color: #888;
}

.keyboard-hints kbd {
  background: #333;
  color: #fff;
  padding: 0.125rem 0.375rem;
  border-radius: 3px;
  font-family: inherit;
  font-size: 0.75rem;
}

@media (prefers-color-scheme: dark) {
  .game-number {
    color: #aaa;
  }

  .board {
    --wall-color: #ccc;
    background: #2a2a2a;
  }

  .cell {
    border-color: #444;
  }

  .solution-column {
    background: #3a3a3a;
  }

  .solution-column:hover {
    background: #454545;
  }

  .solution-column.active {
    background: #3a3a3a;
  }

  .solution-column.winner {
    background: #3d3820;
    border-color: #b8960b;
  }

  .solution-header {
    border-bottom-color: #555;
  }

  .solution-moves,
  .move-arrow,
  .player-solution-header .player-name,
  .player-solution-header .solution-moves {
    color: #ddd;
  }

  .move-pos,
  .player-solution-header .solution-time {
    color: #999;
  }

  .player-solution-header {
    border-bottom-color: #555;
  }

  .target-number {
    color: white;
  }
}
</style>
