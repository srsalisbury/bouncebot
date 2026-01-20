<script setup lang="ts">
import { ref, computed } from 'vue'
import { DIRECTION_ARROWS, getRobotColor, MOBILE_ASPECT_RATIO, MOBILE_WIDTH_BREAKPOINT } from '../constants'
import { useSwipe } from '../composables/useSwipe'
import type { PlayerSolution } from '../gen/bouncebot_pb'
import type { Timestamp } from '@bufbuild/protobuf/wkt'
import type { Direction } from '../constants'

interface MoveWithDirection {
  robotId: number
  direction: Direction
}

const props = defineProps<{
  playerSolutions: PlayerSolution[]
  currentPlayerSolution?: PlayerSolution | null
  topThreeSolutions: PlayerSolution[]
  solverSolutions: PlayerSolution[]
  activeIndex: number
  replayMoveIndex: number
  getPlayerName: (playerId: string) => string
  getPlayerColor: (playerId: string) => string
  isSolverSolution: (playerId: string) => boolean
  getPlayerSolutionMoves: (solution: PlayerSolution) => MoveWithDirection[]
  singlePlayer?: boolean
  gameStartedAt?: Timestamp
}>()

const emit = defineEmits<{
  switchSolution: [index: number]
  replaySolution: []
}>()

const isExpanded = ref(false)
const drawerRef = ref<HTMLElement | null>(null)

// Check if we're on mobile/vertical layout
function isMobile(): boolean {
  const aspectRatio = window.innerWidth / window.innerHeight
  return aspectRatio < MOBILE_ASPECT_RATIO || window.innerWidth <= MOBILE_WIDTH_BREAKPOINT
}

// Switch solution and auto-collapse on mobile
function handleSolutionClick(index: number) {
  emit('switchSolution', index)
  if (isMobile()) {
    isExpanded.value = false
  }
}

// Replay solution and auto-collapse on mobile
function handleReplayClick() {
  emit('replaySolution')
  if (isMobile()) {
    isExpanded.value = false
  }
}

// Check if current player is in top 3
const isCurrentPlayerInTopThree = computed(() => {
  if (!props.currentPlayerSolution) return false
  return props.topThreeSolutions.some(s => s.playerId === props.currentPlayerSolution?.playerId)
})

// Calculate the starting index for solver solutions in the combined array
// New order: [top 3, solver, current player (if not in top 3)]
const solverStartIndex = computed(() => {
  if (props.singlePlayer) {
    return props.playerSolutions.length
  }
  // Multiplayer: top 3 (if more than 1)
  return props.topThreeSolutions.length > 1 ? props.topThreeSolutions.length : 0
})

// Calculate the starting index for current player's solution (after solver)
const currentPlayerStartIndex = computed(() => {
  return solverStartIndex.value + props.solverSolutions.length
})

// Total count of solutions for swipe navigation
const totalSolutionCount = computed(() => {
  let count = solverStartIndex.value + props.solverSolutions.length
  if (props.currentPlayerSolution && !isCurrentPlayerInTopThree.value) {
    count += 1
  }
  return count
})

// Check if any player solutions are displayed before solver (for divider logic)
const hasTopThreeSolutions = computed(() => {
  if (props.singlePlayer) {
    return props.playerSolutions.length > 0
  }
  return props.topThreeSolutions.length > 1
})

// Check if current player solution should be shown (not in top 3)
const showCurrentPlayerSolution = computed(() => {
  return !!props.currentPlayerSolution && !isCurrentPlayerInTopThree.value
})

// Swipe to expand/collapse and switch solutions
useSwipe({
  target: drawerRef,
  onSwipe: (direction) => {
    if (direction === 'down' && isExpanded.value) {
      isExpanded.value = false
    } else if (direction === 'up' && !isExpanded.value) {
      isExpanded.value = true
    } else if (!isExpanded.value) {
      // Swipe left/right on collapsed drawer to switch solutions
      if (direction === 'left' && props.activeIndex < totalSolutionCount.value - 1) {
        emit('switchSolution', props.activeIndex + 1)
      } else if (direction === 'right' && props.activeIndex > 0) {
        emit('switchSolution', props.activeIndex - 1)
      }
    }
  },
  minDistance: 30,
})

function toggleExpanded() {
  isExpanded.value = !isExpanded.value
}

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

// Current active solution for collapsed header
// New order: [top 3, solver, current player (if not in top 3)]
function getActiveSolution(): PlayerSolution | undefined {
  if (props.singlePlayer) {
    const solverOffset = props.activeIndex - props.playerSolutions.length
    if (solverOffset >= 0 && solverOffset < props.solverSolutions.length) {
      return props.solverSolutions[solverOffset]
    }
    return props.playerSolutions[props.activeIndex]
  }

  // Multiplayer mode - new order: [top 3, solver, current player]

  // Top 3 solutions
  if (props.topThreeSolutions.length > 1 && props.activeIndex < props.topThreeSolutions.length) {
    return props.topThreeSolutions[props.activeIndex]
  }

  // Solver solutions
  const solverOffset = props.activeIndex - solverStartIndex.value
  if (solverOffset >= 0 && solverOffset < props.solverSolutions.length) {
    return props.solverSolutions[solverOffset]
  }

  // Current player's solution (at the end, if not in top 3)
  if (showCurrentPlayerSolution.value && props.activeIndex === currentPlayerStartIndex.value) {
    return props.currentPlayerSolution!
  }

  return undefined
}

// Check if current active solution is a solver
function isActiveSolver() {
  if (props.singlePlayer) {
    return props.activeIndex >= props.playerSolutions.length && props.solverSolutions.length > 0
  }
  const solverOffset = props.activeIndex - solverStartIndex.value
  return solverOffset >= 0 && solverOffset < props.solverSolutions.length
}

</script>

<template>
  <div
    ref="drawerRef"
    class="player-solutions-drawer"
    :class="{ expanded: isExpanded }"
  >
    <!-- Collapsed header bar -->
    <div class="drawer-header" @click="toggleExpanded">
      <div class="drawer-handle" />
      <div class="header-content">
        <div class="winner-info">
          <template v-if="isActiveSolver()">
            <img src="/favicon_light.svg" alt="" class="solver-icon solver-icon-light" />
            <img src="/favicon_dark.svg" alt="" class="solver-icon solver-icon-dark" />
          </template>
          <template v-else>
            <span
              class="player-dot"
              :style="{ backgroundColor: getPlayerColor(getActiveSolution()?.playerId ?? '') }"
            />
          </template>
          <span class="player-name">{{ getPlayerName(getActiveSolution()?.playerId ?? '') }}</span>
          <span class="move-count">
            {{ getActiveSolution()?.moves.length ?? 0 }}
            {{ getActiveSolution()?.moves.length === 1 ? 'move' : 'moves' }}
          </span>
          <span v-if="!singlePlayer && topThreeSolutions.length > 1 && activeIndex === 0" class="winner-badge">Winner</span>
        </div>
      </div>
    </div>

    <!-- Expanded content -->
    <div v-if="isExpanded" class="drawer-content">
      <div class="solutions-columns">
        <!-- Single player mode: use playerSolutions directly -->
        <template v-if="singlePlayer">
          <div
            v-for="(solution, index) in playerSolutions"
            :key="solution.playerId"
            class="solution-column"
            :class="{ active: index === activeIndex }"
            @click="handleSolutionClick(index)"
          >
            <!-- Replay button on active solution -->
            <button
              v-if="index === activeIndex && solution.moves.length > 0"
              class="replay-btn"
              @click.stop="handleReplayClick()"
            >
              <span class="play-icon">▶</span>
            </button>
            <div class="player-solution-header">
              <div class="player-name-row">
                <span
                  class="player-dot"
                  :style="{ backgroundColor: getPlayerColor(solution.playerId) }"
                />
                <span class="player-name">{{ getPlayerName(solution.playerId) }}</span>
              </div>
              <span class="solution-moves">{{ solution.moves.length }}</span>
            </div>
            <div class="move-list">
              <div
                v-for="(move, i) in getPlayerSolutionMoves(solution)"
                :key="i"
                class="move-item"
                :class="{ animating: index === activeIndex && i < replayMoveIndex }"
              >
                <span class="move-robot" :style="{ backgroundColor: getRobotColor(move.robotId) }">
                  {{ move.robotId + 1 }}
                </span>
                <span class="move-arrow">{{ DIRECTION_ARROWS[move.direction] }}</span>
              </div>
            </div>
          </div>
        </template>

        <!-- Multiplayer mode: grouped layout -->
        <!-- Order: Top 3 | Bot | Current player (if not in top 3) -->
        <template v-else>
          <!-- Top 3 solutions (skip if only 1 player) -->
          <template v-if="topThreeSolutions.length > 1">
            <div
              v-for="(solution, index) in topThreeSolutions"
              :key="'top-' + solution.playerId"
              class="solution-column"
              :class="{ active: activeIndex === index, winner: index === 0 }"
              @click="handleSolutionClick(index)"
            >
              <!-- Replay button on active solution -->
              <button
                v-if="activeIndex === index && solution.moves.length > 0"
                class="replay-btn"
                @click.stop="handleReplayClick()"
              >
                <span class="play-icon">▶</span>
              </button>
              <div class="player-solution-header">
                <div class="player-name-row">
                  <span
                    class="player-dot"
                    :style="{ backgroundColor: getPlayerColor(solution.playerId) }"
                  />
                  <span class="player-name">{{ getPlayerName(solution.playerId) }}</span>
                </div>
                <span class="solution-moves">{{ solution.moves.length }}</span>
                <span class="solution-time">{{ formatSolveTime(solution.solvedAt) }}</span>
              </div>
              <div class="move-list">
                <div
                  v-for="(move, i) in getPlayerSolutionMoves(solution)"
                  :key="i"
                  class="move-item"
                  :class="{ animating: activeIndex === index && i < replayMoveIndex }"
                >
                  <span class="move-robot" :style="{ backgroundColor: getRobotColor(move.robotId) }">
                    {{ move.robotId + 1 }}
                  </span>
                  <span class="move-arrow">{{ DIRECTION_ARROWS[move.direction] }}</span>
                </div>
              </div>
            </div>
          </template>

          <!-- Divider and solver solutions -->
          <template v-if="solverSolutions.length">
            <div v-if="hasTopThreeSolutions" class="solutions-divider"></div>
            <div
              v-for="(solverSolution, solverIndex) in solverSolutions"
              :key="solverSolution.playerId"
              class="solution-column solver"
              :class="{ active: activeIndex === solverStartIndex + solverIndex }"
              @click="handleSolutionClick(solverStartIndex + solverIndex)"
            >
              <!-- Replay button on active solution -->
              <button
                v-if="activeIndex === solverStartIndex + solverIndex && solverSolution.moves.length > 0"
                class="replay-btn"
                @click.stop="handleReplayClick()"
              >
                <span class="play-icon">▶</span>
              </button>
              <div class="player-solution-header">
                <div class="player-name-row">
                  <img src="/favicon_light.svg" alt="" class="solver-icon solver-icon-light" />
                  <img src="/favicon_dark.svg" alt="" class="solver-icon solver-icon-dark" />
                  <span class="player-name">{{ getPlayerName(solverSolution.playerId) }}</span>
                </div>
                <span class="solution-moves">{{ solverSolution.moves.length }}</span>
                <span class="solution-time">&nbsp;</span>
              </div>
              <div class="move-list">
                <div
                  v-for="(move, i) in getPlayerSolutionMoves(solverSolution)"
                  :key="i"
                  class="move-item"
                  :class="{ animating: activeIndex === solverStartIndex + solverIndex && i < replayMoveIndex }"
                >
                  <span class="move-robot" :style="{ backgroundColor: getRobotColor(move.robotId) }">
                    {{ move.robotId + 1 }}
                  </span>
                  <span class="move-arrow">{{ DIRECTION_ARROWS[move.direction] }}</span>
                </div>
              </div>
            </div>
          </template>

          <!-- Current player's solution (only if not in top 3) -->
          <template v-if="showCurrentPlayerSolution">
            <div v-if="hasTopThreeSolutions || solverSolutions.length" class="solutions-divider"></div>
            <div
              class="solution-column current-player"
              :class="{ active: activeIndex === currentPlayerStartIndex }"
              @click="handleSolutionClick(currentPlayerStartIndex)"
            >
              <!-- Replay button on active solution -->
              <button
                v-if="activeIndex === currentPlayerStartIndex && currentPlayerSolution!.moves.length > 0"
                class="replay-btn"
                @click.stop="handleReplayClick()"
              >
                <span class="play-icon">▶</span>
              </button>
              <div class="player-solution-header">
                <div class="player-name-row">
                  <span
                    class="player-dot"
                    :style="{ backgroundColor: getPlayerColor(currentPlayerSolution!.playerId) }"
                  />
                  <span class="player-name">{{ getPlayerName(currentPlayerSolution!.playerId) }}</span>
                </div>
                <span class="solution-moves">{{ currentPlayerSolution!.moves.length }}</span>
                <span class="solution-time">{{ formatSolveTime(currentPlayerSolution!.solvedAt) }}</span>
              </div>
              <div class="move-list">
                <div
                  v-for="(move, i) in getPlayerSolutionMoves(currentPlayerSolution!)"
                  :key="i"
                  class="move-item"
                  :class="{ animating: activeIndex === currentPlayerStartIndex && i < replayMoveIndex }"
                >
                  <span class="move-robot" :style="{ backgroundColor: getRobotColor(move.robotId) }">
                    {{ move.robotId + 1 }}
                  </span>
                  <span class="move-arrow">{{ DIRECTION_ARROWS[move.direction] }}</span>
                </div>
              </div>
            </div>
          </template>
        </template>

        <!-- Solver solutions for single player mode -->
        <template v-if="singlePlayer && solverSolutions.length">
          <div v-if="playerSolutions.length > 0" class="solutions-divider"></div>
          <div
            v-for="(solverSolution, solverIndex) in solverSolutions"
            :key="solverSolution.playerId"
            class="solution-column solver"
            :class="{ active: activeIndex === solverStartIndex + solverIndex }"
            @click="handleSolutionClick(solverStartIndex + solverIndex)"
          >
            <!-- Replay button on active solution -->
            <button
              v-if="activeIndex === solverStartIndex + solverIndex && solverSolution.moves.length > 0"
              class="replay-btn"
              @click.stop="handleReplayClick()"
            >
              <span class="play-icon">▶</span>
            </button>
            <div class="player-solution-header">
              <div class="player-name-row">
                <img src="/favicon_light.svg" alt="" class="solver-icon solver-icon-light" />
                <img src="/favicon_dark.svg" alt="" class="solver-icon solver-icon-dark" />
                <span class="player-name">{{ getPlayerName(solverSolution.playerId) }}</span>
              </div>
              <span class="solution-moves">{{ solverSolution.moves.length }}</span>
              <span class="solution-time">&nbsp;</span>
            </div>
            <div class="move-list">
              <div
                v-for="(move, i) in getPlayerSolutionMoves(solverSolution)"
                :key="i"
                class="move-item"
                :class="{ animating: activeIndex === solverStartIndex + solverIndex && i < replayMoveIndex }"
              >
                <span class="move-robot" :style="{ backgroundColor: getRobotColor(move.robotId) }">
                  {{ move.robotId + 1 }}
                </span>
                <span class="move-arrow">{{ DIRECTION_ARROWS[move.direction] }}</span>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.player-solutions-drawer {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: #1a1a1a;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.3);
  z-index: 100;
  transition: max-height 0.3s ease;
  max-height: 60px;
  overflow: hidden;
}

.player-solutions-drawer.expanded {
  max-height: 50vh;
  overflow-y: auto;
}

.drawer-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  cursor: pointer;
  user-select: none;
  min-height: 44px;
}

.drawer-handle {
  position: absolute;
  top: 8px;
  left: 50%;
  transform: translateX(-50%);
  width: 40px;
  height: 4px;
  background: #444;
  border-radius: 2px;
}

.header-content {
  flex: 1;
  display: flex;
  justify-content: center;
}

.winner-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.winner-info .player-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.winner-info .solver-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.winner-info .solver-icon-light {
  display: block;
}

.winner-info .solver-icon-dark {
  display: none;
}

@media (prefers-color-scheme: dark) {
  .winner-info .solver-icon-light {
    display: none;
  }

  .winner-info .solver-icon-dark {
    display: block;
  }
}

.winner-info .player-name {
  font-size: 1.1rem;
  font-weight: 600;
  color: #fff;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.winner-info .move-count {
  font-size: 0.9rem;
  color: #888;
}

.winner-badge {
  font-size: 0.7rem;
  font-weight: 600;
  color: #000;
  background: #ffd700;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}

.drawer-content {
  padding: 0 1rem 1rem;
}

.solutions-columns {
  display: flex;
  flex-direction: row;
  gap: 0.5rem;
  justify-content: center;
  overflow-x: auto;
  padding: 12px;
  padding-bottom: 0.5rem;
}

.solutions-divider {
  width: 1px;
  background: #555;
  align-self: stretch;
  margin: 0 0.25rem;
  flex-shrink: 0;
}

.solution-column {
  position: relative;
  min-width: 70px;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.4rem;
  border-radius: 6px;
  background: #2a2a2a;
  cursor: pointer;
  flex-shrink: 0;
}

.replay-btn {
  /* Reset iOS button styling */
  -webkit-appearance: none;
  appearance: none;
  position: absolute;
  top: -8px;
  left: -8px;
  width: 22px;
  min-width: 22px;
  max-width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #43a047;
  color: #333;
  border: 2px solid #1a1a1a;
  font-size: 10px;
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

.replay-btn:hover {
  background: #3aa876;
}

.play-icon {
  margin-left: 2px;
}

.solution-column:hover {
  background: #333;
}

.solution-column.active {
  background: #2a2a2a;
  box-shadow: 0 0 0 2px #43a047;
}

.solution-column.winner {
  background: #3d3820;
  border: 2px solid #b8960b;
}

.solution-column.winner.active {
  box-shadow: 0 0 0 2px #43a047;
}

.player-solution-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.2rem;
  padding-bottom: 0.4rem;
  margin-bottom: 0.25rem;
  border-bottom: 1px solid #444;
}

.player-name-row {
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.player-name-row .player-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.player-name-row .solver-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

.player-name-row .solver-icon-light {
  display: block;
}

.player-name-row .solver-icon-dark {
  display: none;
}

@media (prefers-color-scheme: dark) {
  .player-name-row .solver-icon-light {
    display: none;
  }

  .player-name-row .solver-icon-dark {
    display: block;
  }
}

.player-name-row .player-name {
  font-size: 0.8rem;
  font-weight: 600;
  color: #ddd;
  max-width: 60px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.solution-moves {
  font-size: 1.2rem;
  font-weight: 600;
  color: #fff;
}

.solution-time {
  font-size: 0.75rem;
  color: #888;
}

.move-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  max-height: 200px;
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
  background: #43a047;
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
  border: 0.5px solid black;
  text-shadow: -0.3px -0.3px 0 black, 0.3px -0.3px 0 black, -0.3px 0.3px 0 black, 0.3px 0.3px 0 black;
}

.move-arrow {
  font-size: 18px;
  color: #ddd;
  width: 18px;
  text-align: center;
}
</style>
