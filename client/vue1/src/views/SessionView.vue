<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { bounceBotClient } from '../services/connectClient'
import { useGameStore } from '../stores/gameStore'
import { useSessionStore } from '../stores/sessionStore'
import { websocketService, type WebSocketEvent, type PlayerSolvedPayload, type SolutionRetractedPayload } from '../services/websocket'
import type { Session } from '../gen/bouncebot_pb'
import GameBoard from '../components/GameBoard.vue'
import PlayersPanel from '../components/PlayersPanel.vue'

const props = defineProps<{
  sessionId: string
}>()

const router = useRouter()
const gameStore = useGameStore()
const sessionStore = useSessionStore()

const session = ref<Session | null>(null)
const isLoading = ref(true)
const isStarting = ref(false)
const isJoining = ref(false)
const error = ref<string | null>(null)
const pollInterval = ref<number | null>(null)
const joinName = ref('')
const notification = ref<string | null>(null)
const notificationTimeout = ref<number | null>(null)
const bestSubmittedMoveCount = ref<number | null>(null)
const showRetractConfirm = ref(false)
const pendingRetractAction = ref<(() => void) | null>(null)
const useFixedBoard = ref(false)

const hasGame = computed(() => session.value?.currentGame != null)
const shareUrl = computed(() => window.location.href)
const hasJoined = computed(() => sessionStore.currentPlayerId != null)

function getPlayerName(playerId: string): string {
  const player = session.value?.players.find(p => p.id === playerId)
  return player?.name ?? 'Unknown'
}

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

async function startGame(useFixedBoard = false) {
  isStarting.value = true
  error.value = null

  try {
    const sess = await bounceBotClient.startGame({ sessionId: props.sessionId, useFixedBoard })
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

async function joinSession() {
  if (!joinName.value.trim()) {
    error.value = 'Please enter your name'
    return
  }

  isJoining.value = true
  error.value = null

  try {
    const sess = await bounceBotClient.joinSession({
      sessionId: props.sessionId,
      playerName: joinName.value.trim(),
    })
    // Find ourselves in the players list (we're the last one added)
    const player = sess.players[sess.players.length - 1]
    if (player) {
      sessionStore.setCurrentPlayer(player.id, player.name)
    }
    // Reload session to get updated player list
    await loadSession()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to join session'
  } finally {
    isJoining.value = false
  }
}

function copyShareUrl() {
  navigator.clipboard.writeText(shareUrl.value)
}

function goHome() {
  router.push('/')
}

function showNotification(message: string) {
  notification.value = message
  if (notificationTimeout.value) {
    clearTimeout(notificationTimeout.value)
  }
  notificationTimeout.value = window.setTimeout(() => {
    notification.value = null
  }, 4000)
}

async function submitSolution(moveCount: number) {
  if (!sessionStore.currentPlayerId) return
  // Only submit if this is better than our previous best (or first submission)
  if (bestSubmittedMoveCount.value !== null && moveCount >= bestSubmittedMoveCount.value) return

  try {
    await bounceBotClient.submitSolution({
      sessionId: props.sessionId,
      playerId: sessionStore.currentPlayerId,
      moveCount,
    })
    bestSubmittedMoveCount.value = moveCount
    // Reload session to get updated solutions list
    await loadSession()
  } catch (e) {
    console.error('Failed to submit solution:', e)
  }
}

async function retractSolution() {
  if (!sessionStore.currentPlayerId) return

  try {
    await bounceBotClient.retractSolution({
      sessionId: props.sessionId,
      playerId: sessionStore.currentPlayerId,
    })
    // Reload session to get updated solutions list
    await loadSession()
  } catch (e) {
    console.error('Failed to retract solution:', e)
  }
}

// Called by GameBoard before undoing/deleting a solved solution
function onBeforeRetract(action: () => void) {
  if (bestSubmittedMoveCount.value !== null) {
    // User has a submitted solution - show confirmation
    pendingRetractAction.value = action
    showRetractConfirm.value = true
  } else {
    // No submitted solution - just proceed
    action()
  }
}

async function confirmRetract() {
  showRetractConfirm.value = false
  await retractSolution()
  if (pendingRetractAction.value) {
    pendingRetractAction.value()
    pendingRetractAction.value = null
  }
  // Clear after the action so the watch doesn't re-submit while puzzle is still solved
  bestSubmittedMoveCount.value = null
}

function cancelRetract() {
  showRetractConfirm.value = false
  pendingRetractAction.value = null
}


function handleWebSocketEvent(event: WebSocketEvent) {
  if (event.type === 'player_joined') {
    // Refresh session to get updated player list
    loadSession()
  } else if (event.type === 'game_started') {
    // Refresh session to get the game
    bestSubmittedMoveCount.value = null // Reset for new game
    loadSession()
  } else if (event.type === 'player_solved') {
    const payload = event.payload as PlayerSolvedPayload
    // Show notification for other players' solutions
    if (payload.playerId !== sessionStore.currentPlayerId) {
      const playerName = getPlayerName(payload.playerId)
      showNotification(`${playerName} solved in ${payload.moveCount} moves!`)
    }
    // Refresh session to get updated solutions list
    loadSession()
  } else if (event.type === 'solution_retracted') {
    const payload = event.payload as SolutionRetractedPayload
    // Show notification for other players' retractions
    if (payload.playerId !== sessionStore.currentPlayerId) {
      const playerName = getPlayerName(payload.playerId)
      showNotification(`${playerName} retracted their solution`)
    }
    // Refresh session to get updated solutions list
    loadSession()
  }
}

function connectWebSocket() {
  if (hasJoined.value) {
    websocketService.connect(props.sessionId, handleWebSocketEvent)
  }
}

// Connect to WebSocket when user joins
watch(hasJoined, (joined) => {
  if (joined) {
    connectWebSocket()
    // Stop polling once connected via WebSocket
    if (pollInterval.value) {
      clearInterval(pollInterval.value)
      pollInterval.value = null
    }
  }
})

// Submit solution when puzzle is solved (or improved)
watch(
  () => gameStore.isSolved,
  (solved) => {
    if (solved && hasGame.value) {
      submitSolution(gameStore.moveCount)
    }
  }
)

// Handle dialog keyboard events at window level (capture phase) to intercept before GameBoard
function globalKeydownHandler(event: KeyboardEvent) {
  if (!showRetractConfirm.value) return
  if (event.key === 'Enter') {
    event.preventDefault()
    event.stopPropagation()
    confirmRetract()
  } else if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    cancelRetract()
  }
}

// Add/remove global listener when dialog opens/closes
watch(showRetractConfirm, (show) => {
  if (show) {
    window.addEventListener('keydown', globalKeydownHandler, true) // capture phase
  } else {
    window.removeEventListener('keydown', globalKeydownHandler, true)
  }
})

onMounted(() => {
  loadSession()
  // If already joined, connect WebSocket immediately
  if (hasJoined.value) {
    connectWebSocket()
  } else {
    // Poll until joined (for users who haven't joined yet)
    pollInterval.value = window.setInterval(loadSession, 3000)
  }
})

onUnmounted(() => {
  if (pollInterval.value) {
    clearInterval(pollInterval.value)
  }
  if (notificationTimeout.value) {
    clearTimeout(notificationTimeout.value)
  }
  websocketService.disconnect()
})
</script>

<template>
  <div class="session-view">
    <!-- Notification toast -->
    <div v-if="notification" class="notification">
      {{ notification }}
    </div>

    <!-- Loading state -->
    <div v-if="isLoading" class="loading">Loading session...</div>

    <!-- Error state -->
    <div v-else-if="error && !session" class="error-container">
      <div class="error-icon">⚠</div>
      <div class="error-message">{{ error }}</div>
      <button class="btn" @click="goHome">Back to Home</button>
    </div>

    <!-- Join form (for users who navigated directly to session URL) -->
    <div v-else-if="session && !hasJoined" class="join-view">
      <h1 class="title">BounceBot</h1>
      <p class="subtitle">Join Session</p>

      <div class="card">
        <div class="players-section">
          <h3>Players in session ({{ session.players.length }})</h3>
          <PlayersPanel :players="session.players" />
        </div>

        <div class="form-group">
          <label for="joinName">Your Name</label>
          <input
            id="joinName"
            v-model="joinName"
            type="text"
            placeholder="Enter your name"
            maxlength="20"
            @keyup.enter="joinSession"
          />
        </div>

        <div v-if="error" class="error">{{ error }}</div>

        <button
          class="btn primary join-btn"
          :disabled="isJoining"
          @click="joinSession"
        >
          {{ isJoining ? 'Joining...' : 'Join Session' }}
        </button>
      </div>
    </div>

    <!-- Game in progress -->
    <div v-else-if="hasGame && session" class="game-wrapper">
      <div class="game-header">
        <PlayersPanel :players="session.players" :solutions="session.solutions" :game-started-at="session.gameStartedAt" compact />
      </div>
      <div class="game-container">
        <GameBoard :on-before-retract="onBeforeRetract" />
      </div>
    </div>

    <!-- Waiting room -->
    <div v-else-if="session && hasJoined" class="waiting-room">
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
          <PlayersPanel :players="session.players" />
        </div>

        <div v-if="error" class="error">{{ error }}</div>

        <div class="start-options">
          <button
            class="btn primary start-btn"
            :disabled="isStarting"
            @click="startGame(useFixedBoard)"
          >
            {{ isStarting ? 'Starting...' : 'Start Game' }}
          </button>

          <label class="fixed-board-option">
            <input type="checkbox" v-model="useFixedBoard" />
            Use fixed board (for testing)
          </label>
        </div>

        <p class="hint">Share the link above with friends to play together!</p>
      </div>
    </div>

    <!-- Retract confirmation dialog -->
    <div v-if="showRetractConfirm" class="dialog-overlay" @click.self="cancelRetract">
      <div class="dialog">
        <h3>Retract Solution?</h3>
        <p>
          You have a submitted solution. This action will retract it and you'll
          lose credit for finding this solution.
        </p>
        <div class="dialog-actions">
          <button class="btn" @click="cancelRetract">Cancel</button>
          <button class="btn danger" @click="confirmRetract">Retract</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.session-view {
  min-height: 100vh;
  position: relative;
}

.notification {
  position: fixed;
  top: 1rem;
  left: 50%;
  transform: translateX(-50%);
  background: #42b883;
  color: #fff;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  z-index: 1000;
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
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

.game-wrapper {
  display: flex;
  flex-direction: column;
  padding: 1rem;
}

.game-header {
  margin-bottom: 1rem;
  padding: 0.5rem 1rem;
  background: #1a1a1a;
  border-radius: 8px;
}

.game-container {
  /* contains GameBoard */
}

.waiting-room,
.join-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #aaa;
  font-size: 0.9rem;
}

.form-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #333;
  border-radius: 6px;
  background: #242424;
  color: #fff;
  font-size: 1rem;
  box-sizing: border-box;
}

.form-group input:focus {
  outline: none;
  border-color: #42b883;
}

.join-btn {
  width: 100%;
  padding: 1rem;
  font-size: 1.1rem;
  margin-top: 0.5rem;
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

.players-section :deep(.players-panel) {
  margin-bottom: 1.5rem;
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

.start-options {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.start-btn {
  width: 100%;
  padding: 1rem;
  font-size: 1.1rem;
}

.fixed-board-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #888;
  font-size: 0.9rem;
  cursor: pointer;
}

.fixed-board-option input {
  cursor: pointer;
}

.hint {
  margin: 1rem 0 0;
  color: #666;
  font-size: 0.85rem;
  text-align: center;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 1.5rem;
  max-width: 400px;
  margin: 1rem;
}

.dialog h3 {
  margin: 0 0 1rem;
  color: #eee;
}

.dialog p {
  margin: 0 0 1.5rem;
  color: #aaa;
  line-height: 1.5;
}

.dialog-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

.btn.danger {
  background: #e53935;
  color: #fff;
}

.btn.danger:hover {
  background: #c62828;
}
</style>
