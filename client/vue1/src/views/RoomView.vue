<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useGameStore } from '../stores/gameStore'
import { useRoomStore } from '../stores/roomStore'
import { useRoomConnection } from '../composables/useRoomConnection'
import { useGameActions } from '../composables/useGameActions'
import GameBoard from '../components/GameBoard.vue'
import PlayersPanel from '../components/PlayersPanel.vue'
import LeaderboardModal from '../components/LeaderboardModal.vue'
import { getPlayerColor } from '../constants'

const props = defineProps<{
  roomId: string
}>()

const router = useRouter()
const gameStore = useGameStore()
const roomStore = useRoomStore()

const isStarting = ref(false)
const isJoining = ref(false)
const joinName = ref(roomStore.currentPlayerName ?? '')
const showRetractConfirm = ref(false)
const pendingRetractAction = ref<(() => void) | null>(null)
const gameEnded = ref(false)
const showLeaderboard = ref(false)
const showLeaveConfirm = ref(false)
const showStatsDropdown = ref(false)

// Room connection composable
const {
  room,
  isLoading,
  error,
  normalizedRoomId,
  hasGame,
  hasJoined,
  loadRoom,
  joinRoom: doJoinRoom,
  startGame: doStartGame,
} = useRoomConnection({
  roomId: computed(() => props.roomId),
  onGameStarted: () => {
    gameActions.resetForNewGame()
    gameEnded.value = false
  },
  onGameEnded: () => {
    gameEnded.value = true
  },
  onRoomUpdated: (rm) => {
    if (rm.currentGame) {
      gameStore.applyGame(rm.currentGame, rm.id, rm.gamesPlayed + 1)
    }
    // Restore gameEnded state from server
    if (rm.currentGame && rm.finishedSolving.length === rm.players.length && rm.players.length > 0) {
      gameEnded.value = true
    }
    // Restore bestSubmittedMoveCount from player's solution
    if (roomStore.currentPlayerId && gameActions.bestSubmittedMoveCount.value === null) {
      const mySolution = rm.solutions.find(s => s.playerId === roomStore.currentPlayerId)
      if (mySolution) {
        gameActions.restoreBestMoveCount(mySolution.moves.length)
      }
    }
  },
})

// Game actions composable
const gameActions = useGameActions({
  roomId: normalizedRoomId,
  onRoomUpdated: () => loadRoom(),
  onServerRejection: (reason) => {
    // Server says room or player doesn't exist - clean up and go home
    console.log('Server rejection:', reason)
    roomStore.clearLastRoom()
    roomStore.clear()
    router.push('/')
  },
})

const shareUrl = computed(() => window.location.href)
const isPlayerFinished = computed(() => {
  if (!roomStore.currentPlayerId || !room.value) return false
  return room.value.finishedSolving.includes(roomStore.currentPlayerId)
})

const isPlayerReady = computed(() => {
  if (!roomStore.currentPlayerId || !room.value) return false
  return room.value.readyForNext.includes(roomStore.currentPlayerId)
})

const readyCount = computed(() => room.value?.readyForNext.length ?? 0)
const playerCount = computed(() => room.value?.players.length ?? 0)
const isSinglePlayer = computed(() => roomStore.isSinglePlayer)

const sortedSolutions = computed(() => {
  if (!room.value) return []
  return [...room.value.solutions].sort((a, b) => {
    if (a.moves.length !== b.moves.length) {
      return a.moves.length - b.moves.length
    }
    const timeA = a.solvedAt?.seconds ?? 0n
    const timeB = b.solvedAt?.seconds ?? 0n
    return Number(timeA - timeB)
  })
})

function getPlayerName(playerId: string): string {
  const player = room.value?.players.find(p => p.id === playerId)
  return player?.name ?? 'Unknown'
}

function getPlayerColorById(playerId: string): string {
  const index = room.value?.players.findIndex(p => p.id === playerId) ?? -1
  return index >= 0 ? getPlayerColor(index) : '#888888'
}

async function startGame() {
  isStarting.value = true
  gameActions.resetForNewGame()
  gameEnded.value = false
  // Track puzzle attempt in single-player (before game starts so number is correct)
  if (isSinglePlayer.value) {
    roomStore.recordPuzzleAttempted()
  }
  await doStartGame()
  isStarting.value = false
}

async function nextPuzzle() {
  // For single-player: record stats and start a new game
  if (gameStore.hasAnySolvedSolution) {
    roomStore.recordPuzzleSolved()
  }
  isStarting.value = true
  gameActions.resetForNewGame()
  gameEnded.value = false
  // Increment before game starts so number is correct when rendered
  roomStore.recordPuzzleAttempted()
  await doStartGame()
  isStarting.value = false
}

async function joinRoom() {
  isJoining.value = true
  await doJoinRoom(joinName.value)
  isJoining.value = false
}

function copyShareUrl() {
  navigator.clipboard.writeText(shareUrl.value)
}

function goHome() {
  router.push('/')
}

// Called by GameBoard before undoing/deleting a solved solution
function onBeforeRetract(action: () => void) {
  if (gameActions.bestSubmittedMoveCount.value !== null) {
    pendingRetractAction.value = action
    showRetractConfirm.value = true
  } else {
    action()
  }
}

async function confirmRetract() {
  showRetractConfirm.value = false
  await gameActions.retractSolution()
  if (pendingRetractAction.value) {
    pendingRetractAction.value()
    pendingRetractAction.value = null
  }
  gameActions.clearBestMoveCount()
}

function cancelRetract() {
  showRetractConfirm.value = false
  pendingRetractAction.value = null
}

// Submit solution when puzzle is solved (or improved)
watch(
  () => gameStore.isSolved,
  (solved) => {
    if (solved && hasGame.value) {
      gameActions.submitSolution()
    }
  }
)

// Handle dialog keyboard events at window level
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

watch(showRetractConfirm, (show) => {
  if (show) {
    window.addEventListener('keydown', globalKeydownHandler, true)
  } else {
    window.removeEventListener('keydown', globalKeydownHandler, true)
  }
})

// Leaderboard toggle
function toggleLeaderboard() {
  showLeaderboard.value = !showLeaderboard.value
}

// Leave game
function promptLeaveGame() {
  showLeaveConfirm.value = true
}

function confirmLeave() {
  showLeaveConfirm.value = false
  roomStore.clear()
  router.push('/')
}

function cancelLeave() {
  showLeaveConfirm.value = false
}

function globalKeyHandler(event: KeyboardEvent) {
  // Don't handle if any dialog is open
  if (showRetractConfirm.value || showLeaveConfirm.value) return

  if (event.key === 'l' && (hasGame.value || gameEnded.value)) {
    event.preventDefault()
    toggleLeaderboard()
  } else if (event.key === 'x') {
    event.preventDefault()
    promptLeaveGame()
  }
}

function leaveDialogKeyHandler(event: KeyboardEvent) {
  if (!showLeaveConfirm.value) return
  if (event.key === 'Enter') {
    event.preventDefault()
    event.stopPropagation()
    confirmLeave()
  } else if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    cancelLeave()
  }
}

watch(showLeaveConfirm, (show) => {
  if (show) {
    window.addEventListener('keydown', leaveDialogKeyHandler, true)
  } else {
    window.removeEventListener('keydown', leaveDialogKeyHandler, true)
  }
})

onMounted(() => {
  window.addEventListener('keydown', globalKeyHandler)
})

onUnmounted(() => {
  window.removeEventListener('keydown', globalKeyHandler)
})
</script>

<template>
  <div class="room-view">
    <!-- Loading state -->
    <div v-if="isLoading" class="loading">Loading room...</div>

    <!-- Error state -->
    <div v-else-if="error && !room" class="error-container">
      <div class="error-icon">⚠</div>
      <div class="error-message">{{ error }}</div>
      <button class="btn" @click="goHome">Back to Home</button>
    </div>

    <!-- Join form (for users who navigated directly to room URL) -->
    <div v-else-if="room && !hasJoined" class="join-view">
      <h1 class="title">BounceBot</h1>
      <p class="subtitle">Join Room</p>

      <div class="card">
        <div class="players-section">
          <h3>Players in room ({{ room.players.length }})</h3>
          <PlayersPanel :players="room.players" />
        </div>

        <div class="form-group">
          <label for="joinName">Your Name</label>
          <input
            id="joinName"
            v-model="joinName"
            type="text"
            placeholder="Enter your name"
            maxlength="20"
            @keyup.enter="joinRoom"
          />
        </div>

        <div v-if="error" class="error">{{ error }}</div>

        <button
          class="btn primary join-btn"
          :disabled="isJoining"
          @click="joinRoom"
        >
          {{ isJoining ? 'Joining...' : 'Join Room' }}
        </button>
      </div>
    </div>

    <!-- Game in progress -->
    <div v-else-if="hasGame && room" class="game-wrapper">
      <GameBoard
        :on-before-retract="onBeforeRetract"
        :game-ended="gameEnded"
        :player-solutions="isSinglePlayer ? [] : sortedSolutions"
        :get-player-name="getPlayerName"
        :get-player-color="getPlayerColorById"
        :game-started-at="room.gameStartedAt"
        :room-id="room.id"
        :game-number="(isSinglePlayer ? roomStore.puzzlesAttempted : room.gamesPlayed) + 1"
        :input-blocked="showLeaderboard"
        :single-player="isSinglePlayer"
      >
        <template #header>
          <div class="game-header">
            <!-- Single-player mode: simplified header -->
            <template v-if="isSinglePlayer">
              <div class="solo-stats-wrapper">
                <button
                  class="solo-stats-btn"
                  @click="showStatsDropdown = !showStatsDropdown"
                >
                  Solo · {{ roomStore.puzzlesSolved }}/{{ roomStore.puzzlesAttempted }}
                  <span class="stats-toggle">{{ showStatsDropdown ? '▲' : '▼' }}</span>
                </button>
                <div v-if="showStatsDropdown" class="stats-dropdown">
                  <div class="stats-row">
                    <span class="stats-label">Puzzles solved</span>
                    <span class="stats-value">{{ roomStore.puzzlesSolved }}</span>
                  </div>
                  <div class="stats-row">
                    <span class="stats-label">Puzzles attempted</span>
                    <span class="stats-value">{{ roomStore.puzzlesAttempted }}</span>
                  </div>
                </div>
              </div>
              <span v-if="gameStore.isSolved" class="solved-indicator">✓</span>
              <button
                class="btn primary next-puzzle-btn"
                :disabled="isStarting"
                @click="nextPuzzle"
              >
                {{ isStarting ? 'Loading...' : 'Next Puzzle' }}
              </button>
            </template>
            <!-- Multiplayer mode: full header -->
            <template v-else-if="!gameEnded">
              <PlayersPanel :players="room.players" :solutions="room.solutions" :scores="room.scores" :game-started-at="room.gameStartedAt" :finished-solving="room.finishedSolving" compact />
              <button
                v-if="!isPlayerFinished"
                class="btn done-btn"
                @click="gameActions.markFinishedSolving"
              >
                I'm Finished
              </button>
              <span v-else class="done-indicator">Finished</span>
            </template>
            <template v-else>
              <button
                class="btn leaderboard-btn desktop-only"
                @click="toggleLeaderboard"
              >
                Leaderboard
              </button>
              <LeaderboardModal
                class="mobile-only"
                :show="showLeaderboard"
                :players="room.players"
                :scores="room.scores"
                :games-played="room.gamesPlayed"
                @close="toggleLeaderboard"
              />
              <button
                class="btn ready-btn"
                :class="{ pressed: isPlayerReady }"
                :disabled="isPlayerReady"
                @click="gameActions.markReadyForNext"
              >
                Next Game ({{ readyCount }}/{{ playerCount }})
              </button>
            </template>
          </div>
        </template>
      </GameBoard>
    </div>

    <!-- Waiting room -->
    <div v-else-if="room && hasJoined" class="waiting-room">
      <h1 class="title">BounceBot</h1>
      <p class="subtitle">Waiting Room</p>

      <div class="card">
        <div class="room-info">
          <div class="info-row">
            <span class="label">Room ID:</span>
            <code class="room-id">{{ room.id }}</code>
          </div>
          <button class="btn-small" @click="copyShareUrl">Copy Link</button>
        </div>

        <div class="players-section">
          <h3>Players ({{ room.players.length }})</h3>
          <PlayersPanel :players="room.players" />
        </div>

        <div v-if="error" class="error">{{ error }}</div>

        <div class="start-options">
          <button
            class="btn primary start-btn"
            :disabled="isStarting"
            @click="startGame"
          >
            {{ isStarting ? 'Starting...' : 'Start Game' }}
          </button>
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

    <!-- Leave confirmation dialog -->
    <div v-if="showLeaveConfirm" class="dialog-overlay" @click.self="cancelLeave">
      <div class="dialog">
        <h3>Leave this game?</h3>
        <p>
          You will be removed from this room and returned to the home screen.
        </p>
        <div class="dialog-actions">
          <button class="btn" @click="cancelLeave">Cancel</button>
          <button class="btn danger" @click="confirmLeave">Leave</button>
        </div>
      </div>
    </div>

    <!-- Leaderboard modal -->
    <LeaderboardModal
      :show="showLeaderboard"
      :players="room?.players ?? []"
      :scores="room?.scores ?? []"
      :games-played="room?.gamesPlayed ?? 0"
      @close="showLeaderboard = false"
    />
  </div>
</template>

<style scoped>
.room-view {
  min-height: 100vh;
  position: relative;
  overflow: hidden;
}

.room-view::before {
  content: '';
  position: fixed;
  top: 50%;
  left: 50%;
  width: 170vmax;
  height: 170vmax;
  background-image: url('/pattern_dark.svg');
  background-repeat: no-repeat;
  background-size: 100% 100%;
  transform: translate(-50%, -50%) rotate(22.5deg);
  z-index: -1;
  opacity: 0.7;
}

@media (prefers-color-scheme: light) {
  .room-view::before {
    background-image: url('/pattern_light.svg');
    opacity: 0.4;
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
  align-items: center;
  padding: 1rem 0;
  min-height: 100vh;
  box-sizing: border-box;
}

.game-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.5rem 1rem;
  background: #1a1a1a;
  border-radius: 8px;
}

.done-btn {
  margin-left: auto;
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
  white-space: nowrap;
  background: #43a047;
}

.done-btn:hover {
  background: #388e3c;
}

.done-indicator {
  margin-left: auto;
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
  color: #43a047;
  font-weight: 500;
}

.leaderboard-btn {
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
  white-space: nowrap;
  background: #333;
  color: #ddd;
  border: 1px solid #555;
}

.leaderboard-btn:hover {
  background: #444;
  border-color: #666;
}

.ready-btn {
  margin-left: auto;
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
  white-space: nowrap;
  background: #43a047;
  color: white;
  border: 1px solid #43a047;
}

.ready-btn:hover:not(.pressed) {
  background: #388e3c;
  border-color: #388e3c;
}

.ready-btn.pressed {
  background: #1a2e1a;
  color: #43a047;
  border-color: #43a047;
  cursor: default;
  opacity: 1;
}

.next-game-btn {
  margin-left: auto;
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
  white-space: nowrap;
}

.solo-stats-wrapper {
  position: relative;
}

.solo-stats-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.8rem;
  background: #242424;
  border: 1px solid #333;
  border-radius: 6px;
  color: #aaa;
  font-size: 0.85rem;
  cursor: pointer;
}

.solo-stats-btn:hover {
  background: #2a2a2a;
  border-color: #444;
}

.stats-toggle {
  font-size: 0.6rem;
  color: #666;
}

.stats-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  background: #1a1a1a;
  border: 1px solid #333;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  min-width: 180px;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.stats-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.25rem 0;
}

.stats-row:not(:last-child) {
  border-bottom: 1px solid #2a2a2a;
  padding-bottom: 0.5rem;
  margin-bottom: 0.25rem;
}

.stats-label {
  color: #888;
  font-size: 0.85rem;
}

.stats-value {
  color: #fff;
  font-weight: 500;
  font-size: 0.9rem;
}

.solved-indicator {
  color: #43a047;
  font-size: 1.2rem;
  font-weight: bold;
}

.next-puzzle-btn {
  margin-left: auto;
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
  white-space: nowrap;
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
  border-color: #43a047;
}

.join-btn {
  width: 100%;
  padding: 1rem;
  font-size: 1.1rem;
  margin-top: 0.5rem;
}

.title {
  color: #43a047;
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

.room-info {
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

.room-id {
  background: #242424;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.85rem;
  color: #43a047;
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
  background: #43a047;
  color: #fff;
}

.btn.primary:hover:not(:disabled) {
  background: #388e3c;
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

/* Desktop/mobile visibility */
.mobile-only {
  display: none;
}

/* Vertical layout responsive styles */
@media (max-aspect-ratio: 6/5), (max-width: 1050px) {
  .desktop-only {
    display: none;
  }

  .mobile-only {
    display: block;
  }

  .game-wrapper {
    padding: 0.5rem;
  }

  .game-header {
    flex-wrap: nowrap;
    gap: 0.5rem;
    padding: 0.5rem;
  }

  .ready-btn {
    margin-left: auto;
  }

  .done-btn,
  .done-indicator {
    margin-left: auto;
  }
}
</style>
