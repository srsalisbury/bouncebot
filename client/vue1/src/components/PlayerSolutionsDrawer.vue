<script setup lang="ts">
import { ref } from 'vue'
import { DIRECTION_ARROWS, getRobotColor } from '../constants'
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
  activeIndex: number
  replayMoveIndex: number
  getPlayerName: (playerId: string) => string
  getPlayerColor: (playerId: string) => string
  getPlayerSolutionMoves: (solution: PlayerSolution) => MoveWithDirection[]
  gameStartedAt?: Timestamp
}>()

const emit = defineEmits<{
  switchSolution: [index: number]
}>()

const isExpanded = ref(false)
const drawerRef = ref<HTMLElement | null>(null)

// Swipe down to collapse, up to expand
useSwipe({
  target: drawerRef,
  onSwipe: (direction) => {
    if (direction === 'down' && isExpanded.value) {
      isExpanded.value = false
    } else if (direction === 'up' && !isExpanded.value) {
      isExpanded.value = true
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
function getActiveSolution() {
  return props.playerSolutions[props.activeIndex]
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
          <span
            class="player-dot"
            :style="{ backgroundColor: getPlayerColor(getActiveSolution()?.playerId ?? '') }"
          />
          <span class="player-name">{{ getPlayerName(getActiveSolution()?.playerId ?? '') }}</span>
          <span class="move-count">
            {{ getActiveSolution()?.moves.length ?? 0 }}
            {{ getActiveSolution()?.moves.length === 1 ? 'move' : 'moves' }}
          </span>
          <span v-if="activeIndex === 0" class="winner-badge">Winner</span>
        </div>
      </div>
    </div>

    <!-- Expanded content -->
    <div v-if="isExpanded" class="drawer-content">
      <div class="solutions-columns">
        <div
          v-for="(solution, index) in playerSolutions"
          :key="solution.playerId"
          class="solution-column"
          :class="{ active: index === activeIndex, winner: index === 0 }"
          @click="emit('switchSolution', index)"
        >
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
              :class="{ animating: index === activeIndex && i < replayMoveIndex }"
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
  padding: 4px;
  padding-bottom: 0.5rem;
}

.solution-column {
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

.solution-column:hover {
  background: #333;
}

.solution-column.active {
  background: #2a2a2a;
  box-shadow: 0 0 0 2px #42b883;
}

.solution-column.winner {
  background: #3d3820;
  border: 2px solid #b8960b;
}

.solution-column.winner.active {
  box-shadow: 0 0 0 2px #42b883;
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
