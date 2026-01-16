<script setup lang="ts">
import { DIRECTION_ARROWS, getRobotColor } from '../constants'
import type { PlayerSolution } from '../gen/bouncebot_pb'
import type { Timestamp } from '@bufbuild/protobuf/wkt'
import type { Direction } from '../constants'

interface MoveWithDirection {
  robotId: number
  direction: Direction
}

const props = defineProps<{
  playerSolutions: PlayerSolution[]
  solverSolutions: PlayerSolution[]
  activeIndex: number
  replayMoveIndex: number
  getPlayerName: (playerId: string) => string
  getPlayerColor: (playerId: string) => string
  getPlayerSolutionMoves: (solution: PlayerSolution) => MoveWithDirection[]
  singlePlayer?: boolean
  gameStartedAt?: Timestamp
}>()

const emit = defineEmits<{
  switchSolution: [index: number]
}>()

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
</script>

<template>
  <div class="solutions-panel">
    <div class="solutions-columns">
      <!-- Player solutions -->
      <div
        v-for="(solution, index) in playerSolutions"
        :key="solution.playerId"
        class="solution-column player-solution"
        :class="{ active: index === activeIndex, winner: index === 0 }"
        @click="emit('switchSolution', index)"
      >
        <div class="player-solution-header">
          <div class="player-name-row">
            <span class="player-dot" :style="{ backgroundColor: getPlayerColor(solution.playerId) }"></span>
            <span class="player-name">{{ getPlayerName(solution.playerId) }}</span>
          </div>
          <span class="solution-moves">{{ solution.moves.length }}</span>
          <span v-if="!singlePlayer" class="solution-time">{{ formatSolveTime(solution.solvedAt) }}</span>
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

      <!-- Divider and solver solutions -->
      <template v-if="solverSolutions.length">
        <div class="solutions-divider"></div>
        <div
          v-for="(solverSolution, solverIndex) in solverSolutions"
          :key="solverSolution.playerId"
          class="solution-column player-solution solver"
          :class="{ active: activeIndex === playerSolutions.length + solverIndex }"
          @click="emit('switchSolution', playerSolutions.length + solverIndex)"
        >
          <div class="player-solution-header">
            <div class="player-name-row">
              <img src="/favicon_light.svg" alt="" class="solver-icon solver-icon-light" />
              <img src="/favicon_dark.svg" alt="" class="solver-icon solver-icon-dark" />
              <span class="player-name">{{ getPlayerName(solverSolution.playerId) }}</span>
            </div>
            <span class="solution-moves">{{ solverSolution.moves.length }}</span>
            <span v-if="!singlePlayer" class="solution-time">&nbsp;</span>
          </div>
          <div class="move-list">
            <div
              v-for="(move, i) in getPlayerSolutionMoves(solverSolution)"
              :key="i"
              class="move-item"
              :class="{ animating: activeIndex === playerSolutions.length + solverIndex && i < replayMoveIndex }"
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
</template>

<style scoped>
.solutions-panel {
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

.solution-column {
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
  box-shadow: 0 0 0 2px #43a047;
}

.solution-column.player-solution {
  min-width: 3.375rem;
}

.solution-column.winner {
  background: #fff8dc;
  border: 2px solid #ffd700;
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

.player-solution-header .solver-icon {
  width: 0.875rem;
  height: 0.875rem;
  flex-shrink: 0;
}

.player-solution-header .solver-icon-light {
  display: block;
}

.player-solution-header .solver-icon-dark {
  display: none;
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

.solutions-divider {
  width: 1px;
  background: #999;
  align-self: stretch;
  margin: 0 0.25rem;
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
  background: #43a047;
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

/* Dark mode */
@media (prefers-color-scheme: dark) {
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

  .player-solution-header {
    border-bottom-color: #555;
  }

  .player-solution-header .player-name,
  .player-solution-header .solution-moves {
    color: #ddd;
  }

  .player-solution-header .solution-time {
    color: #999;
  }

  .player-solution-header .solver-icon-light {
    display: none;
  }

  .player-solution-header .solver-icon-dark {
    display: block;
  }

  .solutions-divider {
    background: #555;
  }

  .move-arrow {
    color: #ddd;
  }
}

/* Vertical layout responsive styles */
@media (max-aspect-ratio: 6/5), (max-width: 1050px) {
  .solutions-panel {
    display: none;
  }
}
</style>
