<script setup lang="ts">
import type { Player } from '../gen/bouncebot_pb'
import { useSessionStore } from '../stores/sessionStore'

defineProps<{
  players: Player[]
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

function getPlayerColor(index: number): string {
  return PLAYER_COLORS[index % PLAYER_COLORS.length] ?? '#888888'
}

function isCurrentPlayer(player: Player): boolean {
  return sessionStore.currentPlayerName != null && player.name === sessionStore.currentPlayerName
}
</script>

<template>
  <div class="players-panel" :class="{ compact }">
    <div v-if="players.length === 1 && !compact" class="waiting-message">
      Waiting for players...
    </div>
    <div class="players-list">
      <div
        v-for="(player, index) in players"
        :key="player.id"
        class="player-item"
        :class="{ current: isCurrentPlayer(player) }"
      >
        <span class="player-dot" :style="{ backgroundColor: getPlayerColor(index) }" />
        <span class="player-name">{{ player.name }}</span>
        <span v-if="isCurrentPlayer(player)" class="you-label">(you)</span>
      </div>
    </div>
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
</style>
