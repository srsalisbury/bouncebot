<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { bounceBotClient } from '../services/connectClient'
import { useGameStore } from '../stores/gameStore'
import type { Session } from '../gen/bouncebot_pb'
import GameBoard from '../components/GameBoard.vue'

const props = defineProps<{
  sessionId: string
}>()

const router = useRouter()
const gameStore = useGameStore()

const session = ref<Session | null>(null)
const isLoading = ref(true)
const isStarting = ref(false)
const error = ref<string | null>(null)
const pollInterval = ref<number | null>(null)

const hasGame = computed(() => session.value?.currentGame != null)
const shareUrl = computed(() => window.location.href)

async function loadSession() {
  try {
    const sess = await bounceBotClient.getSession({ sessionId: props.sessionId })
    const hadGame = hasGame.value
    session.value = sess

    // Only apply game when it first appears (not on every poll)
    if (sess.currentGame && !hadGame) {
      gameStore.applyGame(sess.currentGame)
      // Stop polling once game starts
      if (pollInterval.value) {
        clearInterval(pollInterval.value)
        pollInterval.value = null
      }
    }

    error.value = null
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load session'
  } finally {
    isLoading.value = false
  }
}

async function startGame() {
  isStarting.value = true
  error.value = null

  try {
    const sess = await bounceBotClient.startGame({ sessionId: props.sessionId })
    session.value = sess

    if (sess.currentGame) {
      gameStore.applyGame(sess.currentGame)
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to start game'
  } finally {
    isStarting.value = false
  }
}

function copyShareUrl() {
  navigator.clipboard.writeText(shareUrl.value)
}

function goHome() {
  router.push('/')
}

onMounted(() => {
  loadSession()
  // Poll for session updates every 3 seconds
  pollInterval.value = window.setInterval(loadSession, 3000)
})

onUnmounted(() => {
  if (pollInterval.value) {
    clearInterval(pollInterval.value)
  }
})
</script>

<template>
  <div class="session-view">
    <!-- Loading state -->
    <div v-if="isLoading" class="loading">Loading session...</div>

    <!-- Error state -->
    <div v-else-if="error && !session" class="error-container">
      <div class="error-icon">⚠</div>
      <div class="error-message">{{ error }}</div>
      <button class="btn" @click="goHome">Back to Home</button>
    </div>

    <!-- Game in progress -->
    <div v-else-if="hasGame" class="game-container">
      <GameBoard />
    </div>

    <!-- Waiting room -->
    <div v-else-if="session" class="waiting-room">
      <h1 class="title">BounceBot</h1>
      <p class="subtitle">Waiting Room</p>

      <div class="card">
        <div class="session-info">
          <div class="info-row">
            <span class="label">Session ID:</span>
            <code class="session-id">{{ session.id }}</code>
          </div>
          <button class="btn-small" @click="copyShareUrl">Copy Link</button>
        </div>

        <div class="players-section">
          <h3>Players ({{ session.players.length }})</h3>
          <ul class="player-list">
            <li v-for="player in session.players" :key="player.id" class="player">
              {{ player.name }}
            </li>
          </ul>
        </div>

        <div v-if="error" class="error">{{ error }}</div>

        <button
          class="btn primary start-btn"
          :disabled="isStarting"
          @click="startGame"
        >
          {{ isStarting ? 'Starting...' : 'Start Game' }}
        </button>

        <p class="hint">Share the link above with friends to play together!</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.session-view {
  min-height: 100vh;
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  color: #888;
  font-size: 1.1rem;
}

.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  gap: 1rem;
  padding: 2rem;
}

.error-icon {
  font-size: 3rem;
  color: #e53935;
}

.error-message {
  color: #e53935;
  text-align: center;
}

.game-container {
  padding: 1rem;
}

.waiting-room {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
}

.title {
  color: #42b883;
  margin: 0;
  font-size: 2.5rem;
}

.subtitle {
  color: #888;
  margin: 0.5rem 0 2rem;
}

.card {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 2rem;
  width: 100%;
  max-width: 400px;
}

.session-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #333;
  margin-bottom: 1.5rem;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.label {
  color: #888;
  font-size: 0.9rem;
}

.session-id {
  background: #242424;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.85rem;
  color: #42b883;
}

.btn-small {
  padding: 0.4rem 0.75rem;
  background: #333;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 0.85rem;
  cursor: pointer;
}

.btn-small:hover {
  background: #444;
}

.players-section h3 {
  color: #eee;
  margin: 0 0 1rem;
  font-size: 1rem;
}

.player-list {
  list-style: none;
  padding: 0;
  margin: 0 0 1.5rem;
}

.player {
  padding: 0.5rem 0.75rem;
  background: #242424;
  border-radius: 6px;
  margin-bottom: 0.5rem;
  color: #ddd;
}

.error {
  margin-bottom: 1rem;
  padding: 0.75rem;
  background: rgba(229, 57, 53, 0.1);
  border: 1px solid #e53935;
  border-radius: 6px;
  color: #e53935;
  font-size: 0.9rem;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  background: #333;
  color: #fff;
}

.btn:hover {
  background: #444;
}

.btn.primary {
  background: #42b883;
  color: #fff;
}

.btn.primary:hover:not(:disabled) {
  background: #3aa876;
}

.btn.primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.start-btn {
  width: 100%;
  padding: 1rem;
  font-size: 1.1rem;
}

.hint {
  margin: 1rem 0 0;
  color: #666;
  font-size: 0.85rem;
  text-align: center;
}
</style>
