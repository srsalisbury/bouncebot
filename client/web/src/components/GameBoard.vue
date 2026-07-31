<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useGameStore } from '../stores/gameStore'
import { BOARD_SIZE, WALL_COLOR, DIRECTION_ARROWS, getRobotColor, UNDO_HOLD_DURATION_MS, DOUBLE_TAP_THRESHOLD_MS } from '../constants'
import HowToPlayModal from './HowToPlayModal.vue'
import FirstGameOverlay from './FirstGameOverlay.vue'
import SolutionsDrawer from './SolutionsDrawer.vue'
import PlayerSolutionsDrawer from './PlayerSolutionsDrawer.vue'
import GameBoardPlayerSolutions from './GameBoardPlayerSolutions.vue'
import { useGameInput } from '../composables/useGameInput'
import { useReplay } from '../composables/useReplay'
import { useSwipe } from '../composables/useSwipe'
import type { PlayerSolution } from '../gen/bouncebot_pb'
import type { Timestamp } from '@bufbuild/protobuf/wkt'

const props = defineProps<{
  onBeforeModifyBest?: (solutionIndex: number, action: () => void) => void
  gameEnded?: boolean
  playerSolutions?: PlayerSolution[]
  currentPlayerSolution?: PlayerSolution | null
  topThreeSolutions?: PlayerSolution[]
  solverSolutions?: PlayerSolution[]
  getPlayerName?: (playerId: string) => string
  getPlayerColor?: (playerId: string) => string
  isSolverSolution?: (playerId: string) => boolean
  gameStartedAt?: Timestamp
  roomId?: string
  gameNumber?: number
  inputBlocked?: boolean
  // For pending players watching the in-progress round they'll join next.
  // Implies inputBlocked (a spectator can never move anything, full stop -
  // this isn't the same as inputBlocked alone, which just *temporarily*
  // blocks input, e.g. while a modal is open, and leaves the action buttons
  // visible for when it's released). Also hides the action buttons and
  // in-progress solutions panel, which show the *viewer's own* candidate
  // solutions and don't apply to someone who isn't playing this round.
  spectatorMode?: boolean
  singlePlayer?: boolean
  getBestSubmittedIndex?: () => number | null
  onSolutionDeleted?: (index: number) => void
  playerColor?: string
}>()

const emit = defineEmits<{
  share: []
}>()

const store = useGameStore()
const showHowToPlay = ref(false)
const boardRef = ref<HTMLElement | null>(null)

// First-game onboarding
const onboardingSeen = ref(!!localStorage.getItem('bouncebot_onboarding_seen'))
const isFirstGame = computed(() => !onboardingSeen.value)

function dismissOnboarding() {
  localStorage.setItem('bouncebot_onboarding_seen', '1')
  onboardingSeen.value = true
}

// Percentage-based sizing for responsive board
const CELL_PERCENT = 100 / BOARD_SIZE  // 6.25%

// Replay composable
const {
  activePlayerSolutionIndex,
  replayMoveIndex,
  isReplaying,
  switchToPlayerSolution,
  replayCurrentSolution,
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

// Check if current player is in top 3
const isCurrentPlayerInTopThree = computed(() => {
  if (!props.currentPlayerSolution) return false
  const topThree = props.topThreeSolutions ?? []
  return topThree.some(s => s.playerId === props.currentPlayerSolution?.playerId)
})

// Combined array of player solutions + solver solutions for replay
// In multiplayer: [top 3, solver solutions, current player (if not in top 3)]
// In single player: [playerSolutions, solverSolutions]
const allSolutions = computed(() => {
  const solverSols = props.solverSolutions ?? []

  if (props.singlePlayer) {
    // Single player mode: use playerSolutions directly
    const solutions = props.playerSolutions ?? []
    return [...solutions, ...solverSols]
  }

  // Multiplayer mode: combine top solutions, solver solutions, and current player (if not in top 3)
  const combined: PlayerSolution[] = []

  // Add top solutions first
  const topThree = props.topThreeSolutions ?? []
  combined.push(...topThree)

  // Add solver solutions
  combined.push(...solverSols)

  // Add current player's solution last (only if not already in top solutions)
  if (props.currentPlayerSolution && !isCurrentPlayerInTopThree.value) {
    combined.push(props.currentPlayerSolution)
  }

  return combined
})

// Wrap actions that could modify the best submitted solution
function doUndo() {
  const currentSolution = store.solutions[store.activeSolutionIndex]
  if (currentSolution?.isSolved && props.onBeforeModifyBest) {
    props.onBeforeModifyBest(store.activeSolutionIndex, () => store.undoMove())
  } else {
    store.undoMove()
  }
}

function doDelete() {
  const solution = store.solutions[store.activeSolutionIndex]
  if (solution?.isSolved && props.onBeforeModifyBest) {
    props.onBeforeModifyBest(store.activeSolutionIndex, () => store.deleteSolution(store.activeSolutionIndex))
  } else {
    store.deleteSolution(store.activeSolutionIndex)
  }
}

function doReset() {
  const currentSolution = store.solutions[store.activeSolutionIndex]
  if (currentSolution?.isSolved && props.onBeforeModifyBest) {
    props.onBeforeModifyBest(store.activeSolutionIndex, () => store.resetCurrentSolution())
  } else {
    store.resetCurrentSolution()
  }
}

// Start a new solution, handling auto-delete if at max capacity
function doNewSolution() {
  const bestIndex = props.getBestSubmittedIndex?.() ?? null
  const result = store.startNewSolution(bestIndex)
  if (result.deletedIndex !== null) {
    props.onSolutionDeleted?.(result.deletedIndex)
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
  }, UNDO_HOLD_DURATION_MS)
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

// Clicking a robot directly bypasses useGameInput entirely (that only
// handles keyboard/swipe), so it needs its own guard against the same
// conditions - otherwise a spectator (or anyone else input should be
// blocked for, e.g. a modal open) could still select a robot by clicking it.
function handleRobotClick(robotId: number) {
  if (props.inputBlocked || props.spectatorMode) return
  store.selectRobot(robotId)
}

// Input handling composable
useGameInput(
  {
    onMove: (direction) => store.moveRobot(direction),
    onUndo: doUndo,
    onDelete: doDelete,
    onNewSolution: doNewSolution,
    onSelectRobot: (index) => store.selectRobot(index),
    onSwitchSolution: (delta) => store.switchSolution(store.activeSolutionIndex + delta),
    onSwitchPlayerSolution: (delta) => {
      if (allSolutions.value.length) {
        switchToPlayerSolution(activePlayerSolutionIndex.value + delta, allSolutions.value)
      }
    },
    onReplaySolution: () => store.replaySolution(),
    onReplayPlayerSolution: () => {
      if (allSolutions.value.length) {
        replayCurrentSolution(allSolutions.value)
      }
    },
    onToggleHelp: () => { showHowToPlay.value = !showHowToPlay.value },
    onCloseHelp: () => { showHowToPlay.value = false },
  },
  {
    inputBlocked: computed(() => props.inputBlocked ?? false),
    gameEnded: computed(() => props.gameEnded ?? false),
    spectatorMode: computed(() => props.spectatorMode ?? false),
    helpOpen: showHowToPlay,
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
    if (props.inputBlocked || props.gameEnded || props.spectatorMode) return
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
    if (props.inputBlocked || props.gameEnded || props.spectatorMode) return
    // If swipe started on a robot, select it (if not already selected)
    if (swipeStartRobotId !== null && store.selectedRobotId !== swipeStartRobotId) {
      store.selectRobot(swipeStartRobotId)
    }
    // Move the selected robot
    if (store.selectedRobotId !== null) {
      store.moveRobot(direction)
    }
  },
  enabled: computed(() => !props.inputBlocked && !props.gameEnded && !props.spectatorMode),
})

// When game ends, start showing solutions; when new round starts, stop replay
watch(() => props.gameEnded, (ended) => {
  if (ended && allSolutions.value.length) {
    startInitialReplay(allSolutions.value)
  } else if (!ended) {
    stopReplay()
  }
})

// Prevent double-tap zoom on iOS/iPad
let lastTouchEnd = 0
function preventDoubleTapZoom(event: TouchEvent) {
  const now = Date.now()
  if (now - lastTouchEnd <= DOUBLE_TAP_THRESHOLD_MS) {
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
  if (allSolutions.value.length) {
    switchToPlayerSolution(index, allSolutions.value)
  }
}
</script>

<template>
  <div class="game-container">
    <!-- Game content wrapper -->
    <div class="game-content">
      <!-- Header slot for room controls -->
      <slot name="header" :toggle-help="() => { showHowToPlay = !showHowToPlay }"></slot>
      <!-- Board layout (grid: title on top, board and solutions below) -->
      <div class="board-layout">
        <h1 v-if="props.gameNumber != null" class="title">
          <span v-if="props.spectatorMode" class="watching-label">WATCHING</span>
          <span class="game-label">GAME</span>
          <span class="game-number">{{ props.gameNumber }}</span>
          <button class="share-btn" title="Share this board" aria-label="Share this board" @click="emit('share')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="white">
              <path d="M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z" />
            </svg>
          </button>
        </h1>
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

            <!-- Center logo -->
            <div class="board-logo-bg"></div>
            <img src="/favicon_light.svg" alt="" class="board-logo board-logo-light" />
            <img src="/favicon_dark.svg" alt="" class="board-logo board-logo-dark" />

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
              :class="{ selected: store.selectedRobotId === robot.id, replaying: isReplaying }"
              :style="getRobotStyle(robot)"
              @click="handleRobotClick(robot.id)"
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

            <!-- Room ID label (multiplayer only) -->
            <div v-if="!props.singlePlayer && props.roomId" class="room-id-label">
              Room ID: {{ props.roomId }}
            </div>
          </div>
        </div>

        <!-- Keyboard hints under board -->
        <div class="keyboard-hints">
          <div class="desktop-hints">
            <template v-if="props.gameEnded">
              <kbd>Shift+←→</kbd> switch solutions
            </template>
            <template v-else-if="props.spectatorMode">
              You're watching - you'll play the next game
            </template>
            <template v-else>
              <kbd>1-4</kbd> select · <kbd>↑↓←→</kbd> move · <kbd>z</kbd> undo · <kbd>?</kbd> help
            </template>
          </div>
          <div class="mobile-hints">
            <template v-if="props.gameEnded">
              swipe through solutions
            </template>
            <template v-else-if="props.spectatorMode">
              You're watching - you'll play the next game
            </template>
            <template v-else>
              tap to select · swipe to move
            </template>
          </div>
        </div>

        <!-- Player solutions panel (when game ended) -->
        <GameBoardPlayerSolutions
          v-if="props.gameEnded && (props.currentPlayerSolution || props.topThreeSolutions?.length || props.playerSolutions?.length || props.solverSolutions?.length)"
          :player-solutions="props.playerSolutions ?? []"
          :current-player-solution="props.currentPlayerSolution"
          :top-three-solutions="props.topThreeSolutions ?? []"
          :solver-solutions="props.solverSolutions ?? []"
          :active-index="activePlayerSolutionIndex"
          :replay-move-index="replayMoveIndex"
          :get-player-name="props.getPlayerName ?? (() => 'Unknown')"
          :get-player-color="props.getPlayerColor ?? (() => '#888888')"
          :get-player-solution-moves="getPlayerSolutionMoves"
          :single-player="props.singlePlayer"
          :game-started-at="props.gameStartedAt"
          @switch-solution="handleSwitchPlayerSolution"
        />

        <!-- Normal solutions panel (during game) - the viewer's own candidate
             solutions, not applicable to a spectator watching someone else's round -->
        <div v-else-if="!props.gameEnded && !props.spectatorMode" class="solutions-panel" :style="props.playerColor ? { '--player-color': props.playerColor } as any : undefined">
          <div class="solutions-columns">
            <div
              v-for="(solution, index) in store.solutions"
              :key="index"
              class="solution-column"
              :class="{ active: index === store.activeSolutionIndex }"
              @click="store.switchSolution(index)"
            >
              <!-- Replay button on active solution -->
              <button
                v-if="index === store.activeSolutionIndex && solution.moves.length > 0"
                class="col-replay-btn"
                @click.stop="store.replaySolution()"
              >
                <span class="col-play-icon">&#x25b6;</span>
              </button>
              <!-- Delete button on active solution -->
              <button
                v-if="index === store.activeSolutionIndex && store.solutions.length > 1"
                class="col-delete-btn"
                @click.stop="doDelete()"
              >
                &#x00d7;
              </button>
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
              @click="doNewSolution()"
            >
              New Solution
            </button>
          </div>
        </div>
      </div>

      <!-- Action buttons under board (mobile) -->
      <div v-if="!props.gameEnded && !props.spectatorMode" class="action-buttons mobile-actions">
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
          @click="doNewSolution()"
        >
          New Solution
        </button>
      </div>
    </div>

    <!-- Mobile solutions drawer (only during gameplay, hidden on desktop) -
         the viewer's own candidate solutions, not applicable to a spectator -->
    <SolutionsDrawer
      v-if="!props.gameEnded && !props.spectatorMode"
      class="mobile-drawer"
      :player-color="props.playerColor"
      :get-best-submitted-index="props.getBestSubmittedIndex"
      :on-solution-deleted="props.onSolutionDeleted"
    />

    <!-- Mobile player solutions drawer (only after game ends, hidden on desktop) -->
    <PlayerSolutionsDrawer
      v-if="props.gameEnded && (props.currentPlayerSolution || props.topThreeSolutions?.length || props.playerSolutions?.length || props.solverSolutions?.length)"
      class="mobile-drawer"
      :player-solutions="props.playerSolutions ?? []"
      :current-player-solution="props.currentPlayerSolution"
      :top-three-solutions="props.topThreeSolutions ?? []"
      :solver-solutions="props.solverSolutions ?? []"
      :active-index="activePlayerSolutionIndex"
      :replay-move-index="replayMoveIndex"
      :get-player-name="props.getPlayerName ?? (() => 'Unknown')"
      :get-player-color="props.getPlayerColor ?? (() => '#888888')"
      :is-solver-solution="props.isSolverSolution ?? (() => false)"
      :get-player-solution-moves="getPlayerSolutionMoves"
      :single-player="props.singlePlayer"
      :game-started-at="props.gameStartedAt"
      @switch-solution="(index) => handleSwitchPlayerSolution(index)"
      @replay-solution="replayCurrentSolution(allSolutions)"
    />

    <!-- How to Play modal -->
    <HowToPlayModal :show="showHowToPlay" @close="showHowToPlay = false" />

    <!-- First game onboarding overlay -->
    <FirstGameOverlay :show="isFirstGame" @dismiss="dismissOnboarding" />
  </div>
</template>

<style>
@font-face {
  font-family: 'Conthrax';
  src: url('/fonts/ConthraxRg-Bold.eot');
  src: url('/fonts/ConthraxRg-Bold.eot?#iefix') format('embedded-opentype'),
      url('/fonts/ConthraxRg-Bold.woff2') format('woff2'),
      url('/fonts/ConthraxRg-Bold.woff') format('woff'),
      url('/fonts/ConthraxRg-Bold.ttf') format('truetype'),
      url('/fonts/ConthraxRg-Bold.otf') format('opentype');
  font-weight: bold;
  font-style: normal;
  font-display: swap;
}
</style>

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
  margin: 0;
  margin-top: 1rem;
  font-size: 1.8rem;
  text-align: center;
  position: relative;
}

.share-btn {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: none;
  background: #1976d2;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}

.share-btn:hover {
  background: #1565c0;
}

.watching-label {
  display: block;
  font-size: 0.85rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  color: #888;
}

.game-label {
  font-family: 'Conthrax', sans-serif;
  font-size: 1.8rem;
  font-weight: bold;
  color: #000;
  margin-right: 0.4rem;
}

.game-number {
  font-family: 'Conthrax', sans-serif;
  font-size: 1.8rem;
  font-weight: bold;
  color: #1e88e5;
}

@media (prefers-color-scheme: dark) {
  .game-label {
    color: #fff;
  }
}

.board-logo-bg {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 12.5%;
  height: 12.5%;
  background: #dddddd;
  z-index: 0;
  pointer-events: none;
}

.board-logo {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 10%;
  height: 10%;
  opacity: 1;
  z-index: 1;
  pointer-events: none;
  filter: grayscale(100%) brightness(0.3);
}

.board-logo-light {
  display: block;
}

.board-logo-dark {
  display: none;
}

@media (prefers-color-scheme: dark) {
  .board-logo-bg {
    background: #2a2a2a;
  }
  .board-logo-light {
    display: none;
  }
  .board-logo-dark {
    display: block;
    filter: grayscale(100%) brightness(1.5);
    opacity: 0.75;
  }
}

.board-area {
  grid-column: 1;
  grid-row: 2;
  position: relative;
}

.room-id-label {
  position: absolute;
  bottom: -1.25rem;
  left: 0;
  font-size: 0.8rem;
  color: #888;
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
  padding: 10px;
  padding-bottom: 0;
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
  background: #43a047;
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

.mobile-hints {
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
    /* Match .board's width formula so the share button (right-aligned within
       .title) lines up with the board's actual right edge, not the screen's -
       the board is height-constrained and centered on mobile, so it's often
       narrower than the full row width. Must be `width`, not `max-width`: a
       grid item with auto margins and no explicit width shrinks to fit its
       content instead of stretching, which breaks the alignment. */
    width: min(calc(100% - 2rem), calc(100dvh - 19rem), calc(100vh - 19rem));
    margin-left: auto;
    margin-right: auto;
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

  .desktop-hints {
    display: none;
  }

  .mobile-hints {
    display: block;
  }

  .desktop-actions {
    display: none;
  }

  .mobile-actions {
    display: flex;
    justify-content: center;
    margin-top: 0.5rem;
    margin-bottom: 5rem; /* Space above drawer */
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
  position: relative;
  width: 5rem;
  flex-shrink: 0;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.4rem;
  border-radius: 6px;
  border: 2px solid transparent;
  background: #dddddd;
  cursor: pointer;
  transition: background 0.15s, box-shadow 0.15s;
}

.solution-column:hover {
  background: #cccccc;
}

.solution-column.active {
  background: #dddddd;
  box-shadow: 0 0 0 2px var(--player-color, #43a047);
}

.col-replay-btn,
.col-delete-btn {
  -webkit-appearance: none;
  appearance: none;
  position: absolute;
  width: 22px;
  min-width: 22px;
  max-width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 2px solid #dddddd;
  font-size: 16px;
  font-weight: bold;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  box-sizing: border-box;
  padding: 0;
}

.col-replay-btn {
  top: -8px;
  left: -8px;
  background: var(--player-color, #43a047);
  color: #333;
}

.col-replay-btn:hover {
  filter: brightness(0.85);
}

.col-play-icon {
  font-size: 10px;
}

.col-delete-btn {
  top: -8px;
  right: -8px;
  background: #e53935;
  color: white;
}

.col-delete-btn:hover {
  background: #c62828;
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
  color: var(--player-color, #43a047);
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
  background: var(--player-color, #43a047);
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
    /* Leave 1rem (16px) on each side to reduce accidental browser back/forward gestures */
    width: min(calc(100% - 2rem), calc(100dvh - 19rem), calc(100vh - 19rem));
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
  z-index: 11;
  border: 1px solid black;
  text-shadow: -0.5px -0.5px 0 black, 0.5px -0.5px 0 black, -0.5px 0.5px 0 black, 0.5px 0.5px 0 black;
}

.robot.replaying {
  transition: left 0.4s ease-out, top 0.4s ease-out, transform 0.1s, box-shadow 0.1s;
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

/* Enlarged tap target for mobile - 2x robot size */
@media (max-aspect-ratio: 6/5), (max-width: 1050px) {
  .robot::after {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    width: 200%;
    height: 200%;
    transform: translate(-50%, -50%);
    /* Debug: uncomment to see tap target */
    /* background: rgba(0, 150, 255, 0.3); */
  }
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

  .col-replay-btn,
  .col-delete-btn {
    border-color: #3a3a3a;
  }

  .solution-header {
    border-bottom-color: #555;
  }

  .solution-moves,
  .move-arrow {
    color: #ddd;
  }

  .move-pos {
    color: #999;
  }

  .target-number {
    color: white;
  }
}
</style>
