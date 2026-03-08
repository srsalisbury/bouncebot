<script setup lang="ts">
import { ref, computed } from 'vue'
import { DIRECTION_ARROWS, getRobotColor, MOBILE_ASPECT_RATIO, MOBILE_WIDTH_BREAKPOINT } from '../constants'
import { useSwipe } from '../composables/useSwipe'
import { useSolutionDisplay } from '../composables/useSolutionDisplay'
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

const {
  isCurrentPlayerInTopThree,
  solverStartIndex,
  currentPlayerStartIndex,
  hasTopThreeSolutions,
  showCurrentPlayerSolution,
  getSolveTime
} = useSolutionDisplay(props)

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

// Total count of solutions for swipe navigation
const totalSolutionCount = computed(() => {
  let count = solverStartIndex.value + props.solverSolutions.length
  if (props.currentPlayerSolution && !isCurrentPlayerInTopThree.value) {
    count += 1
  }
  return count
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

// Pill descriptors for collapsed header
interface PillInfo {
  index: number
  moves: number
  color: string
  isSolver: boolean
  isWinner: boolean
  group: 'player' | 'solver' | 'current'
}

const pills = computed<PillInfo[]>(() => {
  const result: PillInfo[] = []

  if (props.singlePlayer) {
    for (let i = 0; i < props.playerSolutions.length; i++) {
      result.push({
        index: i,
        moves: props.playerSolutions[i].moves.length,
        color: props.getPlayerColor(props.playerSolutions[i].playerId),
        isSolver: false,
        isWinner: false,
        group: 'player',
      })
    }
    for (let i = 0; i < props.solverSolutions.length; i++) {
      result.push({
        index: solverStartIndex.value + i,
        moves: props.solverSolutions[i].moves.length,
        color: '',
        isSolver: true,
        isWinner: false,
        group: 'solver',
      })
    }
  } else {
    for (let i = 0; i < props.topThreeSolutions.length; i++) {
      result.push({
        index: i,
        moves: props.topThreeSolutions[i].moves.length,
        color: props.getPlayerColor(props.topThreeSolutions[i].playerId),
        isSolver: false,
        isWinner: i === 0 && props.topThreeSolutions.length > 1,
        group: 'player',
      })
    }
    for (let i = 0; i < props.solverSolutions.length; i++) {
      result.push({
        index: solverStartIndex.value + i,
        moves: props.solverSolutions[i].moves.length,
        color: '',
        isSolver: true,
        isWinner: false,
        group: 'solver',
      })
    }
    if (showCurrentPlayerSolution.value) {
      result.push({
        index: currentPlayerStartIndex.value,
        moves: props.currentPlayerSolution!.moves.length,
        color: props.getPlayerColor(props.currentPlayerSolution!.playerId),
        isSolver: false,
        isWinner: false,
        group: 'current',
      })
    }
  }

  return result
})

// Handle pill click: tap active pill to expand, tap another to switch
function handlePillClick(index: number) {
  if (index === props.activeIndex) {
    toggleExpanded()
  } else {
    emit('switchSolution', index)
  }
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
      <div v-if="!isExpanded" class="header-content">
        <div class="solution-indicators">
          <template v-for="(pill, i) in pills" :key="pill.index">
            <span
              v-if="i > 0 && pill.group !== pills[i - 1].group"
              class="pill-divider"
            />
            <button
              class="solution-pill"
              :class="{
                active: pill.index === activeIndex,
                winner: pill.isWinner,
              }"
              :style="pill.index !== activeIndex ? (pill.isSolver ? { backgroundColor: '#555' } : { backgroundColor: pill.color }) : {}"
              @click.stop="handlePillClick(pill.index)"
            >
              <template v-if="pill.isSolver">
                <img src="/favicon_light.svg" alt="" class="pill-solver-icon pill-solver-icon-light" />
                <img src="/favicon_dark.svg" alt="" class="pill-solver-icon pill-solver-icon-dark" />
              </template>
              {{ pill.moves }}
            </button>
          </template>
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
          <!-- Top solutions (best players) -->
          <template v-if="topThreeSolutions.length > 0">
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
                <span class="solution-time">{{ getSolveTime(solution.playerId) }}</span>
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
                <span class="solution-time">{{ getSolveTime(currentPlayerSolution!.playerId) }}</span>
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
  background: var(--color-bg-dark);
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.3);
  z-index: 100;
  transition: max-height 0.3s ease;
  max-height: calc(60px + env(safe-area-inset-bottom, 0px));
  overflow: hidden;
  padding-bottom: env(safe-area-inset-bottom, 0px);
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

.player-solutions-drawer.expanded .drawer-header {
  padding: 0.4rem 1rem;
  min-height: 0;
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

.solution-indicators {
  display: flex;
  flex-direction: row;
  gap: 0.5rem;
  align-items: center;
  justify-content: center;
}

.solution-pill {
  -webkit-appearance: none;
  appearance: none;
  min-width: 54px;
  height: 32px;
  border-radius: 16px;
  border: 2px solid transparent;
  background: var(--color-bg-surface);
  color: #888;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 0 10px;
  box-sizing: border-box;
}

.solution-pill.active {
  background: var(--color-accent);
  color: #fff;
  border-color: transparent;
}

.solution-pill:not(.active) {
  color: #fff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.solution-pill.winner:not(.active) {
  border-color: var(--color-winner);
}

.pill-divider {
  width: 1px;
  height: 20px;
  background: #555;
  flex-shrink: 0;
}

.pill-solver-icon {
  width: 14px;
  height: 14px;
}

.pill-solver-icon-light {
  display: none;
}

.pill-solver-icon-dark {
  display: block;
}

.solution-pill.active .pill-solver-icon-light {
  display: block;
}

.solution-pill.active .pill-solver-icon-dark {
  display: none;
}

@media (prefers-color-scheme: light) {
  .pill-solver-icon-light {
    display: block;
  }

  .pill-solver-icon-dark {
    display: none;
  }

  .solution-pill.active .pill-solver-icon-light {
    display: block;
  }

  .solution-pill.active .pill-solver-icon-dark {
    display: none;
  }
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
  min-width: 78px;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.4rem;
  border-radius: 6px;
  background: var(--color-bg-panel);
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
  background: var(--color-accent);
  color: #333;
  border: 2px solid var(--color-bg-dark);
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
  background: var(--color-accent-hover);
}

.play-icon {
  margin-left: 2px;
}

.solution-column:hover {
  background: var(--color-bg-surface);
}

.solution-column.active {
  background: var(--color-bg-panel);
  box-shadow: 0 0 0 2px var(--color-accent);
}

.solution-column.winner {
  background: var(--color-winner-bg);
  border: 2px solid var(--color-winner-border);
}

.solution-column.winner.active {
  box-shadow: 0 0 0 2px var(--color-accent);
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
  background: var(--color-accent);
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
