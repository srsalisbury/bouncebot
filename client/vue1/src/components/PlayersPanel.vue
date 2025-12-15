<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import type { Player, PlayerSolution, PlayerScore } from '../gen/bouncebot_pb'
import type { Timestamp } from '@bufbuild/protobuf/wkt'
import { useSessionStore } from '../stores/sessionStore'
import { getPlayerColor } from '../constants'

const props = defineProps<{
  players: Player[]
  solutions?: PlayerSolution[]
  scores?: PlayerScore[]
  gameStartedAt?: Timestamp
  finishedSolving?: string[]
  compact?: boolean
}>()

// Timer state
const elapsedTime = ref<string>('0:00')
const timerInterval = ref<number | null>(null)

const MAX_TIMER_SECONDS = 30 * 60 // 30 minutes

function formatElapsedTime(seconds: number): string {
  // Cap at 30 minutes
  const cappedSeconds = Math.min(seconds, MAX_TIMER_SECONDS)
  const mins = Math.floor(cappedSeconds / 60)
  const secs = Math.floor(cappedSeconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function updateTimer() {
  if (!props.gameStartedAt) {
    elapsedTime.value = '0:00'
    return
  }

  const startSeconds = Number(props.gameStartedAt.seconds ?? 0)
  const startNanos = Number(props.gameStartedAt.nanos ?? 0)
  const startTime = startSeconds + startNanos / 1e9

  const now = Date.now() / 1000
  const elapsed = Math.max(0, now - startTime)
  elapsedTime.value = formatElapsedTime(elapsed)
}

function startTimer() {
  stopTimer()
  updateTimer()
  timerInterval.value = window.setInterval(updateTimer, 1000)
}

function stopTimer() {
  if (timerInterval.value !== null) {
    clearInterval(timerInterval.value)
    timerInterval.value = null
  }
}

// Start timer when gameStartedAt changes
watch(() => props.gameStartedAt, (newVal) => {
  if (newVal) {
    startTimer()
  } else {
    stopTimer()
    elapsedTime.value = '0:00'
  }
}, { immediate: true })

onMounted(() => {
  if (props.gameStartedAt) {
    startTimer()
  }
})

onUnmounted(() => {
  stopTimer()
})

const sessionStore = useSessionStore()

// Map player IDs to their color index (based on join order, so colors stay stable when sorting)
const playerColorMap = computed(() => {
  const map = new Map<string, number>()
  props.players.forEach((player, index) => {
    map.set(player.id, index)
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
      if (solA.moves.length !== solB.moves.length) {
        return solA.moves.length - solB.moves.length
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

  const bestMoveCount = Math.min(...props.solutions.map((s) => s.moves.length))
  const bestSolutions = props.solutions.filter((s) => s.moves.length === bestMoveCount)

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

function getPlayerWins(player: Player): number {
  const score = props.scores?.find((s) => s.playerId === player.id)
  return score?.wins ?? 0
}

function getPlayerColorFor(player: Player): string {
  const index = playerColorMap.value.get(player.id) ?? 0
  return getPlayerColor(index)
}

function isCurrentPlayer(player: Player): boolean {
  return sessionStore.currentPlayerId != null && player.id === sessionStore.currentPlayerId
}

function isLeader(player: Player): boolean {
  return player.id === leaderPlayerId.value
}

function isFinishedSolving(player: Player): boolean {
  return props.finishedSolving?.includes(player.id) ?? false
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
        <span class="player-dot" :style="{ backgroundColor: getPlayerColorFor(player) }" />
        <span class="player-name">{{ player.name }}</span>
        <span v-if="isCurrentPlayer(player)" class="you-label">(you)</span>
        <span v-if="!compact && getPlayerWins(player) > 0" class="wins-badge">
          {{ getPlayerWins(player) }} {{ getPlayerWins(player) === 1 ? 'win' : 'wins' }}
        </span>
        <span v-if="getPlayerSolution(player)" class="solution-badge">
          {{ getPlayerSolution(player)?.moves.length }} moves
          <span v-if="getSolveTime(getPlayerSolution(player)!)" class="solve-time">
            {{ getSolveTime(getPlayerSolution(player)!) }}
          </span>
        </span>
        <span v-if="isFinishedSolving(player)" class="done-check" title="Finished solving">✓</span>
      </div>
    </TransitionGroup>
    <!-- Timer display in compact mode (during game) - on right end -->
    <div v-if="compact && gameStartedAt" class="timer-display">
      <span class="timer-icon">⏱</span>
      <span class="timer-value">{{ elapsedTime }}</span>
    </div>
  </div>
</template>

<style scoped>
.players-panel {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.timer-display {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.6rem;
  background: #242424;
  border-radius: 6px;
  color: #ddd;
  font-size: 0.9rem;
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

.wins-badge {
  padding: 0.1rem 0.35rem;
  background: #4a4a6a;
  color: #ddd;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 500;
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

.done-check {
  color: #42b883;
  font-weight: bold;
  font-size: 1rem;
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
.compact {
  flex-direction: row;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.compact .timer-display {
  padding: 0.3rem 0.5rem;
  margin-left: auto;
  font-size: 0.85rem;
}

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
