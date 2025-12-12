<script setup lang="ts">
import { computed } from 'vue'
import type { Player, PlayerSolution } from '../gen/bouncebot_pb'
import type { Timestamp } from '@bufbuild/protobuf/wkt'
import { useSessionStore } from '../stores/sessionStore'

const props = defineProps<{
  players: Player[]
  solutions?: PlayerSolution[]
  gameStartedAt?: Timestamp
  compact?: boolean
}>()

const sessionStore = useSessionStore()

// Player colors matching robot colors from constants
const PLAYER_COLORS = [
  '#e53935', // red
  '#1e88e5', // blue
  '#43a047', // green
  '#fdd835', // yellow
  '#8e24aa', // purple
  '#fb8c00', // orange
  '#00acc1', // cyan
  '#d81b60', // pink
]

// Create a map of original player index to color (so colors stay stable when sorting)
const playerColorMap = computed(() => {
  const map = new Map<string, string>()
  props.players.forEach((player, index) => {
    map.set(player.id, PLAYER_COLORS[index % PLAYER_COLORS.length] ?? '#888888')
  })
  return map
})

// Sort players: solved players first (by move count, then by time), then unsolved
const sortedPlayers = computed(() => {
  if (!props.solutions || props.solutions.length === 0) {
    return props.players
  }

  return [...props.players].sort((a, b) => {
    const solA = props.solutions?.find((s) => s.playerId === a.id)
    const solB = props.solutions?.find((s) => s.playerId === b.id)

    // Both solved: sort by move count (ascending), then by time (earlier first)
    if (solA && solB) {
      if (solA.moveCount !== solB.moveCount) {
        return solA.moveCount - solB.moveCount
      }
      // Same move count: sort by solved time (earlier first)
      const timeA = solA.solvedAt?.seconds ?? 0
      const timeB = solB.solvedAt?.seconds ?? 0
      return Number(timeA) - Number(timeB)
    }
    // Only A solved: A comes first
    if (solA) return -1
    // Only B solved: B comes first
    if (solB) return 1
    // Neither solved: keep original order
    return 0
  })
})

// Find the leader: best move count AND earliest to achieve it
const leaderPlayerId = computed(() => {
  if (!props.solutions || props.solutions.length === 0) return null

  const bestMoveCount = Math.min(...props.solutions.map((s) => s.moveCount))
  const bestSolutions = props.solutions.filter((s) => s.moveCount === bestMoveCount)

  // Find earliest among best solutions
  let earliest = bestSolutions[0]
  for (const sol of bestSolutions) {
    const solTime = sol.solvedAt?.seconds ?? 0
    const earliestTime = earliest?.solvedAt?.seconds ?? 0
    if (Number(solTime) < Number(earliestTime)) {
      earliest = sol
    }
  }

  return earliest?.playerId ?? null
})

function getPlayerSolution(player: Player): PlayerSolution | undefined {
  return props.solutions?.find((s) => s.playerId === player.id)
}

function getPlayerColor(player: Player): string {
  return playerColorMap.value.get(player.id) ?? '#888888'
}

function isCurrentPlayer(player: Player): boolean {
  return sessionStore.currentPlayerId != null && player.id === sessionStore.currentPlayerId
}

function isLeader(player: Player): boolean {
  return player.id === leaderPlayerId.value
}

function getSolveTime(solution: PlayerSolution): string | null {
  if (!props.gameStartedAt || !solution.solvedAt) return null

  const startSeconds = Number(props.gameStartedAt.seconds ?? 0)
  const startNanos = Number(props.gameStartedAt.nanos ?? 0)
  const solvedSeconds = Number(solution.solvedAt.seconds ?? 0)
  const solvedNanos = Number(solution.solvedAt.nanos ?? 0)

  const elapsedSeconds = (solvedSeconds - startSeconds) + (solvedNanos - startNanos) / 1e9

  if (elapsedSeconds < 60) {
    return `${elapsedSeconds.toFixed(1)}s`
  } else {
    const minutes = Math.floor(elapsedSeconds / 60)
    const seconds = elapsedSeconds % 60
    return `${minutes}:${seconds.toFixed(0).padStart(2, '0')}`
  }
}
</script>

<template>
  <div class="players-panel" :class="{ compact }">
    <div v-if="players.length === 1 && !compact" class="waiting-message">
      Waiting for players...
    </div>
    <TransitionGroup name="player-list" tag="div" class="players-list">
      <div
        v-for="player in sortedPlayers"
        :key="player.id"
        class="player-item"
        :class="{ current: isCurrentPlayer(player), solved: getPlayerSolution(player), leader: isLeader(player) }"
      >
        <span class="player-dot" :style="{ backgroundColor: getPlayerColor(player) }" />
        <span class="player-name">{{ player.name }}</span>
        <span v-if="isCurrentPlayer(player)" class="you-label">(you)</span>
        <span v-if="getPlayerSolution(player)" class="solution-badge">
          {{ getPlayerSolution(player)?.moveCount }} moves
          <span v-if="getSolveTime(getPlayerSolution(player)!)" class="solve-time">
            {{ getSolveTime(getPlayerSolution(player)!) }}
          </span>
        </span>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.players-panel {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.waiting-message {
  color: #888;
  font-size: 0.85rem;
  font-style: italic;
  margin-bottom: 0.25rem;
}

.players-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  position: relative;
}

.player-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.6rem;
  background: #242424;
  border-radius: 6px;
  color: #ddd;
  font-size: 0.9rem;
}

.player-item.current {
  background: #2a3a2a;
  border: 1px solid #42b883;
}

.player-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.player-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.you-label {
  color: #42b883;
  font-size: 0.8rem;
}

.solution-badge {
  margin-left: auto;
  padding: 0.15rem 0.4rem;
  background: #42b883;
  color: #fff;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.solve-time {
  opacity: 0.85;
  margin-left: 0.3rem;
}

.player-item.solved {
  background: #1a2e1a;
}

.player-item.leader {
  background: #2e2a1a;
  border: 1px solid #ffd700;
}

.player-item.leader .solution-badge {
  background: #ffd700;
  color: #000;
}

/* Compact mode for game view */
.compact .players-list {
  flex-direction: row;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.compact .player-item {
  padding: 0.3rem 0.5rem;
  font-size: 0.85rem;
}

.compact .player-dot {
  width: 10px;
  height: 10px;
}

/* List transition animations */
.player-list-move {
  transition: transform 0.4s ease;
}

.player-list-enter-active {
  transition: all 0.3s ease;
}

.player-list-leave-active {
  transition: all 0.3s ease;
  position: absolute;
}

.player-list-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.player-list-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
