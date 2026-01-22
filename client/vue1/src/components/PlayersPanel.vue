<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import type { Player, PlayerSolution, PlayerScore } from '../gen/bouncebot_pb'
import type { Timestamp } from '@bufbuild/protobuf/wkt'
import { useRoomStore } from '../stores/roomStore'
import { getPlayerColor, MAX_GAME_TIMER_SECONDS } from '../constants'
import { getFormattedTimes, calculateDurationSeconds } from '../services/timeUtils'

const props = defineProps<{
  players: Player[]
  solutions?: PlayerSolution[]
  scores?: PlayerScore[]
  gameStartedAt?: Timestamp
  finishedSolving?: string[]
  compact?: boolean
  hideWaitingMessage?: boolean
  showHost?: boolean
}>()

// Dropdown state for vertical layout
const isDropdownOpen = ref(false)

function toggleDropdown() {
  isDropdownOpen.value = !isDropdownOpen.value
}

// Timer state
const elapsedTime = ref<string>('0:00')
const timerInterval = ref<number | null>(null)

function formatElapsedTime(seconds: number): string {
  // Cap at 30 minutes
  const cappedSeconds = Math.min(seconds, MAX_GAME_TIMER_SECONDS)
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

const roomStore = useRoomStore()

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

// Get leader player object
const leaderPlayer = computed(() => {
  if (!leaderPlayerId.value) return null
  return props.players.find(p => p.id === leaderPlayerId.value) ?? null
})

// Get leader's solution
const leaderSolution = computed(() => {
  if (!leaderPlayerId.value || !props.solutions) return null
  return props.solutions.find(s => s.playerId === leaderPlayerId.value) ?? null
})

function getPlayerSolution(player: Player): PlayerSolution | undefined {
  return props.solutions?.find((s) => s.playerId === player.id)
}

function getPlayerWins(player: Player): number {
  const score = props.scores?.find((s) => s.playerId === player.id)
  return score?.wins ?? 0
}

function getPlayerColorFor(player: Player): string {
  return getPlayerColor(player.colorIndex)
}

function isCurrentPlayer(player: Player): boolean {
  return roomStore.currentPlayerId != null && player.id === roomStore.currentPlayerId
}

function isLeader(player: Player): boolean {
  return player.id === leaderPlayerId.value
}

function isFinishedSolving(player: Player): boolean {
  return props.finishedSolving?.includes(player.id) ?? false
}

function isHost(player: Player): boolean {
  return props.showHost && props.players.length > 0 && player.id === props.players[0]?.id
}

const formattedTimesMap = computed(() => {
  if (!props.solutions || !props.gameStartedAt) return new Map<string, string>()

  const timeData = props.solutions
    .filter(s => s.solvedAt)
    .map(s => ({
      id: s.playerId,
      seconds: calculateDurationSeconds(props.gameStartedAt!, s.solvedAt!)
    }))

  return getFormattedTimes(timeData)
})

function getSolveTime(solution: PlayerSolution): string | null {
  return formattedTimesMap.value.get(solution.playerId) ?? null
}
</script>

<template>
  <div class="players-panel" :class="{ compact, 'dropdown-open': isDropdownOpen }">
    <div v-if="players.length === 1 && !compact && !hideWaitingMessage" class="waiting-message">
      Waiting for players...
    </div>

    <!-- Dropdown summary for vertical layout (compact mode) -->
    <div v-if="compact" class="dropdown-summary" :class="{ 'has-leader': leaderPlayer }" @click="toggleDropdown">
      <template v-if="leaderPlayer && leaderSolution">
        <span class="player-dot" :style="{ backgroundColor: getPlayerColorFor(leaderPlayer) }" />
        <span class="player-name">{{ leaderPlayer.name }}</span>
        <span class="solution-badge leader-badge">
          {{ leaderSolution.moves.length }}
        </span>
      </template>
      <template v-else>
        <span class="no-solutions">{{ players.length }} {{ players.length === 1 ? 'player' : 'players' }}</span>
      </template>
      <span class="dropdown-arrow">{{ isDropdownOpen ? '▲' : '▼' }}</span>
    </div>

    <!-- Full player list (always shown on desktop, dropdown on vertical layout) -->
    <TransitionGroup name="player-list" tag="div" class="players-list">
      <div
        v-for="player in sortedPlayers"
        :key="player.id"
        class="player-item"
        :class="{ current: isCurrentPlayer(player), solved: getPlayerSolution(player), leader: isLeader(player) }"
      >
        <span class="player-dot" :style="{ backgroundColor: getPlayerColorFor(player) }" />
        <span class="player-name">{{ player.name }}</span>
        <svg v-if="isHost(player)" class="host-crown" title="Host" viewBox="0 0 24 24" fill="currentColor">
          <path d="M5 16L3 5l5.5 5L12 4l3.5 6L21 5l-2 11H5zm14 3c0 .6-.4 1-1 1H6c-.6 0-1-.4-1-1v-1h14v1z"/>
        </svg>
        <span v-if="isCurrentPlayer(player)" class="you-label">(you)</span>
        <span v-if="!compact && getPlayerWins(player) > 0" class="wins-badge">
          {{ getPlayerWins(player) }} {{ getPlayerWins(player) === 1 ? 'win' : 'wins' }}
        </span>
        <span v-if="getPlayerSolution(player)" class="solution-badge">
          {{ getPlayerSolution(player)?.moves.length }}<span class="solution-details">
            {{ getPlayerSolution(player)?.moves.length === 1 ? ' move' : ' moves' }}<span v-if="getSolveTime(getPlayerSolution(player)!)" class="solve-time">
              {{ getSolveTime(getPlayerSolution(player)!) }}
            </span>
          </span>
        </span>
        <span v-if="isFinishedSolving(player)" class="done-check" title="Finished solving">✓</span>
      </div>
    </TransitionGroup>
    <!-- Timer display in compact mode (during game) - on right end -->
    <div v-if="compact && gameStartedAt" class="timer-display">
      <img src="/timer.svg" alt="" class="timer-icon" />
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
  border: 1px solid #333;
  border-radius: 6px;
  color: #ddd;
  font-size: 0.9rem;
}

.timer-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

.timer-value {
  min-width: 2.8em;
  text-align: right;
  font-variant-numeric: tabular-nums;
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
  border: 1px solid #43a047;
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

.host-crown {
  width: 14px;
  height: 14px;
  color: #ffd700;
  flex-shrink: 0;
}

.you-label {
  color: #43a047;
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
  background: #43a047;
  color: #fff;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.solve-time {
  opacity: 0.85;
  margin-left: 0.3rem;
}

/* Hide details (moves word + time) by default in compact mode */
.compact .solution-details {
  display: none;
}

.done-check {
  color: #43a047;
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

/* Dropdown summary - hidden by default (shown only in vertical layout) */
.dropdown-summary {
  display: none;
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

/* Vertical layout responsive styles */
@media (max-aspect-ratio: 6/5), (max-width: 1050px) {
  .compact {
    flex-wrap: nowrap;
    overflow: visible;
    position: relative;
  }

  /* Show dropdown summary */
  .compact .dropdown-summary {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.3rem 0.5rem;
    background: #242424;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.85rem;
    color: #ddd;
  }

  .compact .dropdown-summary:hover {
    background: #2a2a2a;
  }

  .compact .dropdown-summary.has-leader {
    background: #2e2a1a;
    border: 1px solid #ffd700;
  }

  .compact .dropdown-summary .player-dot {
    width: 10px;
    height: 10px;
  }

  .compact .dropdown-summary .player-name {
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .compact .dropdown-summary .leader-badge {
    background: #ffd700;
    color: #000;
  }

  .compact .dropdown-summary .no-solutions {
    color: #888;
  }

  .compact .dropdown-summary .dropdown-arrow {
    margin-left: 0.25rem;
    font-size: 0.7rem;
    color: #888;
  }

  /* Hide player list by default */
  .compact .players-list {
    display: none;
    position: absolute;
    top: 100%;
    left: 0;
    z-index: 200;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.5rem;
    background: #1a1a1a;
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    margin-top: 0.25rem;
    min-width: 280px;
    width: max-content;
  }

  /* Show player list when dropdown is open */
  .compact.dropdown-open .players-list {
    display: flex;
  }

  /* Show solution details in expanded dropdown */
  .compact.dropdown-open .solution-details {
    display: inline;
  }

  .compact .player-item {
    flex-shrink: 0;
  }

  .compact .players-list .player-name {
    max-width: none;
  }

  .compact .timer-display {
    flex-shrink: 0;
  }
}
</style>
