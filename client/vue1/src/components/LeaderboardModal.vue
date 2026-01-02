<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import type { Player, PlayerScore } from '../gen/bouncebot_pb'
import { getPlayerColor } from '../constants'

const props = defineProps<{
  show: boolean
  players: Player[]
  scores: PlayerScore[]
  gamesPlayed: number
}>()

const emit = defineEmits<{
  close: []
}>()

// Detect mobile/vertical layout
const isMobile = ref(false)

function checkMobile() {
  const aspectRatio = window.innerWidth / window.innerHeight
  isMobile.value = aspectRatio < 6/5 || window.innerWidth <= 1050
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

// Map player IDs to their color index (based on join order)
const playerColorMap = computed(() => {
  const map = new Map<string, number>()
  props.players.forEach((player, index) => {
    map.set(player.id, index)
  })
  return map
})

function getPlayerColorById(playerId: string): string {
  const index = playerColorMap.value.get(playerId) ?? 0
  return getPlayerColor(index)
}

function getPlayerWins(playerId: string): number {
  const score = props.scores.find((s) => s.playerId === playerId)
  return score?.wins ?? 0
}

// Sort players by wins (descending), then by name
const rankedPlayers = computed(() => {
  return [...props.players].sort((a, b) => {
    const winsA = getPlayerWins(a.id)
    const winsB = getPlayerWins(b.id)
    if (winsA !== winsB) return winsB - winsA
    return a.name.localeCompare(b.name)
  })
})

function handleBackdropClick(event: MouseEvent) {
  if (event.target === event.currentTarget) {
    emit('close')
  }
}
</script>

<template>
  <!-- Desktop: Full modal -->
  <Teleport to="body" :disabled="isMobile">
    <div v-if="show && !isMobile" class="modal-backdrop" @click="handleBackdropClick">
      <div class="modal">
        <button class="close-btn" @click="emit('close')">×</button>
        <h2>Leaderboard</h2>
        <p class="games-played">{{ gamesPlayed }} {{ gamesPlayed === 1 ? 'game' : 'games' }} played</p>

        <div class="leaderboard">
          <div
            v-for="(player, index) in rankedPlayers"
            :key="player.id"
            class="player-row"
            :class="{ winner: index === 0 && getPlayerWins(player.id) > 0 }"
          >
            <span class="rank">{{ index + 1 }}</span>
            <span class="player-dot" :style="{ backgroundColor: getPlayerColorById(player.id) }" />
            <span class="player-name">{{ player.name }}</span>
            <span class="wins">{{ getPlayerWins(player.id) }} {{ getPlayerWins(player.id) === 1 ? 'win' : 'wins' }}</span>
          </div>
        </div>

        <p class="hint">Press <kbd>L</kbd> or click outside to close</p>
      </div>
    </div>
  </Teleport>

  <!-- Mobile: Dropdown -->
  <div v-if="isMobile" class="leaderboard-dropdown" :class="{ open: show }">
    <div class="dropdown-summary" @click="emit('close')">
      <span class="dropdown-label">Leaderboard</span>
      <span class="dropdown-arrow">{{ show ? '▲' : '▼' }}</span>
    </div>
    <div v-if="show" class="dropdown-content">
      <p class="games-played">{{ gamesPlayed }} {{ gamesPlayed === 1 ? 'game' : 'games' }} played</p>
      <div class="leaderboard">
        <div
          v-for="(player, index) in rankedPlayers"
          :key="player.id"
          class="player-row"
          :class="{ winner: index === 0 && getPlayerWins(player.id) > 0 }"
        >
          <span class="rank">{{ index + 1 }}</span>
          <span class="player-dot" :style="{ backgroundColor: getPlayerColorById(player.id) }" />
          <span class="player-name">{{ player.name }}</span>
          <span class="wins">{{ getPlayerWins(player.id) }} {{ getPlayerWins(player.id) === 1 ? 'win' : 'wins' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 1.5rem 2rem;
  min-width: 300px;
  max-width: 400px;
  max-height: 80vh;
  overflow-y: auto;
  position: relative;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
}

.close-btn {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  background: none;
  border: none;
  color: #888;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

h2 {
  margin: 0 0 0.25rem 0;
  color: #42b883;
  font-size: 1.5rem;
}

.games-played {
  margin: 0 0 1rem 0;
  color: #888;
  font-size: 0.85rem;
}

.leaderboard {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.player-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.75rem;
  background: #242424;
  border-radius: 6px;
}

.player-row.winner {
  background: #2e2a1a;
  border: 1px solid #ffd700;
}

.rank {
  font-size: 1rem;
  font-weight: 600;
  color: #888;
  min-width: 1.5rem;
  text-align: center;
}

.player-row.winner .rank {
  color: #ffd700;
}

.player-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.player-name {
  flex: 1;
  color: #eee;
  font-size: 1rem;
}

.wins {
  color: #42b883;
  font-size: 0.9rem;
  font-weight: 500;
}

.hint {
  margin: 1rem 0 0 0;
  color: #666;
  font-size: 0.8rem;
  text-align: center;
}

kbd {
  background: #333;
  color: #fff;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: inherit;
  font-size: 0.75rem;
}

/* Mobile dropdown styles */
.leaderboard-dropdown {
  position: relative;
}

.dropdown-summary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.8rem;
  background: #333;
  border: 1px solid #555;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  color: #ddd;
}

.dropdown-summary:hover {
  background: #444;
  border-color: #666;
}

.leaderboard-dropdown.open .dropdown-summary {
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
  border-bottom-color: transparent;
}

.dropdown-arrow {
  font-size: 0.7rem;
  color: #888;
}

.dropdown-content {
  position: absolute;
  top: 100%;
  left: 0;
  z-index: 200;
  padding: 0.75rem;
  background: #1a1a1a;
  border: 1px solid #555;
  border-top: none;
  border-radius: 0 0 6px 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  min-width: 280px;
  width: max-content;
}

.dropdown-content .games-played {
  margin: 0 0 0.5rem 0;
  font-size: 0.8rem;
}

.dropdown-content .player-row {
  padding: 0.4rem 0.5rem;
  font-size: 0.85rem;
}

.dropdown-content .player-name {
  text-align: left;
}

.dropdown-content .player-dot {
  width: 10px;
  height: 10px;
}

.dropdown-content .rank {
  font-size: 0.85rem;
  min-width: 1.25rem;
}

.dropdown-content .wins {
  font-size: 0.8rem;
}
</style>
