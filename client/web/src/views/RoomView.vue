<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useGameStore } from '../stores/gameStore'
import { useRoomStore } from '../stores/roomStore'
import { useRoomConnection } from '../composables/useRoomConnection'
import { useGameActions } from '../composables/useGameActions'
import { bounceBotClient } from '../services/connectClient'
import GameBoard from '../components/GameBoard.vue'
import PlayersPanel from '../components/PlayersPanel.vue'
import LeaderboardModal from '../components/LeaderboardModal.vue'
import SettingsModal from '../components/SettingsModal.vue'
import ShareModal from '../components/ShareModal.vue'
import { getPlayerColor, SOLVER_PLAYER_ID, SOLO_PLAYER_ID } from '../constants'
import { create } from '@bufbuild/protobuf'
import { PlayerSolutionSchema, BotPosSchema, PositionSchema } from '../gen/bouncebot_pb'
import { encodeShareCode } from '../shareCode'
import { config } from '../config'

const props = defineProps<{
  roomId: string
}>()

const router = useRouter()
const gameStore = useGameStore()
const roomStore = useRoomStore()

const isStarting = ref(false)
const isJoining = ref(false)
const joinName = ref(roomStore.currentPlayerName ?? '')
const showProtectedSolutionDialog = ref(false)
const gameEnded = ref(false)
const showLeaderboard = ref(false)
const showLeaveConfirm = ref(false)
const showStatsDropdown = ref(false)
const showSettings = ref(false)
const showShareModal = ref(false)
const shareBoardUrl = ref('')
const showInviteModal = ref(false)

// Room connection composable
const {
  room,
  isLoading,
  error,
  normalizedRoomId,
  hasGame,
  hasJoined,
  solverSolutions,
  wasRemovedFromRoom,
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
  },
  onPlayerBooted: () => {
    // Player was booted from the room - clear state and redirect
    roomStore.clearRoom()
    router.push('/')
  },
})

// Game actions composable
const gameActions = useGameActions({
  roomId: normalizedRoomId,
  onRoomUpdated: () => loadRoom(),
  onServerRejection: (reason) => {
    // Server says room or player doesn't exist - clean up and go home
    console.log('Server rejection:', reason)
    roomStore.clearRoom()
    router.push('/')
  },
})

// Points at the server-rendered /join page rather than the SPA route
// directly, so chat apps/social platforms get a real link preview (og:image
// with the room ID) - they fetch the URL with a plain HTTP GET and don't run
// the SPA's JS, so a client-side route can't carry per-room preview content.
const shareUrl = computed(() => `${config.httpBaseUrl}/join/${normalizedRoomId.value}`)
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
const isRoomCreator = computed(() => {
  if (!room.value || !roomStore.currentPlayerId) return false
  const firstPlayer = room.value.players[0]
  return firstPlayer?.id === roomStore.currentPlayerId
})
const isPendingPlayer = computed(() => {
  if (!room.value || !roomStore.currentPlayerId) return false
  return room.value.pendingPlayers.some(p => p.id === roomStore.currentPlayerId)
})

// Game number for display - shows current game, not next game
// gamesPlayed is incremented when game ends, so we use that directly when ended
const displayedGameNumber = computed(() => {
  if (isSinglePlayer.value) {
    // puzzlesAttempted tracks completed puzzles, so current game = attempted + 1
    return roomStore.puzzlesAttempted + 1
  }
  // When game ends, gamesPlayed is already incremented to include this game
  // When game is active, gamesPlayed is the count before this game
  return gameEnded.value ? room.value?.gamesPlayed ?? 1 : (room.value?.gamesPlayed ?? 0) + 1
})

const sortedSolutions = computed(() => {
  if (!room.value) return []

  // Only player solutions, sorted by move count then time
  return [...room.value.solutions].sort((a, b) => {
    if (a.moves.length !== b.moves.length) {
      return a.moves.length - b.moves.length
    }
    const timeA = a.solvedAt?.seconds ?? 0n
    const timeB = b.solvedAt?.seconds ?? 0n
    return Number(timeA - timeB)
  })
})

// Current player's solution (may be null if they didn't solve)
const currentPlayerSolution = computed(() => {
  if (!roomStore.currentPlayerId || !room.value) return null
  return room.value.solutions.find(s => s.playerId === roomStore.currentPlayerId) ?? null
})

// Top 3 solutions (sorted by moves, then time)
const topThreeSolutions = computed(() => {
  return sortedSolutions.value.slice(0, 3)
})

// Solver solutions as PlayerSolution-like objects (for display)
const solverPlayerSolutions = computed(() => {
  if (solverSolutions.value.length === 0) return []
  // Respect showSolverSolutions setting
  if (room.value?.settings?.showSolverSolutions === false) return []

  return solverSolutions.value.map(sol => {
    const solverMoves = sol.moves.map(m => {
      const pos = create(PositionSchema, { x: m.x, y: m.y })
      return create(BotPosSchema, { id: m.robotId, pos })
    })

    return create(PlayerSolutionSchema, {
      playerId: `${SOLVER_PLAYER_ID}:${sol.solverName}`,
      moves: solverMoves,
    })
  })
})

// Minimum move count across all solver solutions
const minSolverMoves = computed(() => {
  if (solverSolutions.value.length === 0) return null
  // Respect showSolverMoveCount setting
  if (room.value?.settings?.showSolverMoveCount === false) return null
  return Math.min(...solverSolutions.value.map(s => s.moves.length))
})

// Solo player's submitted solution (for display in review screen)
const soloPlayerSolution = computed(() => {
  if (!isSinglePlayer.value || !room.value) return []
  const playerSolution = room.value.solutions.find(s => s.playerId === roomStore.currentPlayerId)
  if (!playerSolution) return []

  return [create(PlayerSolutionSchema, {
    playerId: SOLO_PLAYER_ID,
    moves: playerSolution.moves,
  })]
})

function getPlayerName(playerId: string): string {
  if (playerId === SOLO_PLAYER_ID) {
    return 'You'
  }
  if (playerId.startsWith(SOLVER_PLAYER_ID)) {
    // If only one solver, show "BBot" instead of the actual solver name
    if (solverSolutions.value.length === 1) {
      return 'BBot'
    }
    // Extract solver name from playerId format: "__solver__:solverName"
    const solverName = playerId.split(':')[1]
    return solverName ?? 'Solver'
  }
  const player = room.value?.players.find(p => p.id === playerId)
  return player?.name ?? 'Unknown'
}

const currentPlayerColor = computed(() => getCurrentPlayerActualColor())

function getCurrentPlayerActualColor(): string {
  if (!roomStore.currentPlayerId) return '#43a047'
  const player = room.value?.players.find(p => p.id === roomStore.currentPlayerId)
    ?? room.value?.pendingPlayers.find(p => p.id === roomStore.currentPlayerId)
  return player ? getPlayerColor(player.colorIndex) : '#43a047'
}

function getPlayerColorById(playerId: string): string {
  if (playerId.startsWith(SOLVER_PLAYER_ID)) {
    return '#888888' // Gray for solver
  }
  if (playerId === SOLO_PLAYER_ID) {
    return getCurrentPlayerActualColor()
  }
  // Find player in either players or pendingPlayers
  const player = room.value?.players.find(p => p.id === playerId)
    ?? room.value?.pendingPlayers.find(p => p.id === playerId)
  if (player) {
    return getPlayerColor(player.colorIndex)
  }
  return '#888888'
}

function isSolverSolution(playerId: string): boolean {
  return playerId.startsWith(SOLVER_PLAYER_ID)
}

async function startGame() {
  isStarting.value = true
  gameActions.resetForNewGame()
  gameEnded.value = false
  await doStartGame()
  isStarting.value = false
}

async function nextPuzzle() {
  // For single-player with showSolverSolutions: show review screen first
  if (!gameEnded.value && room.value?.settings?.showSolverSolutions) {
    // Record stats before showing review
    if (gameStore.hasAnySolvedSolution) {
      roomStore.recordPuzzleSolved()
    }
    gameEnded.value = true
    return
  }

  // Start a new game (either from review screen or when showSolverSolutions is off)
  if (!gameEnded.value && gameStore.hasAnySolvedSolution) {
    // Only record if we didn't already record above
    roomStore.recordPuzzleSolved()
  }
  isStarting.value = true
  gameActions.resetForNewGame()
  gameStore.resetSolutions()
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

function openShareModal() {
  if (!room.value?.currentGame) return
  try {
    const code = encodeShareCode(room.value.currentGame)
    const path = router.resolve({ name: 'share', params: { code } }).href
    shareBoardUrl.value = `${window.location.origin}${path}`
    showShareModal.value = true
  } catch (e) {
    console.error('Failed to create a share link for this board', e)
  }
}

async function updateSettings(settings: {
  showSolverMoveCount: boolean
  showSolverSolutions: boolean
  minSolutionLength: number
}) {
  if (!roomStore.currentPlayerId) return

  // Persist settings locally for future sessions
  roomStore.setSettings(settings)

  if (!roomStore.currentSessionToken) return

  try {
    await bounceBotClient.updateRoomSettings({
      roomId: normalizedRoomId.value,
      sessionToken: roomStore.currentSessionToken,
      settings,
    })
  } catch (e) {
    console.error('Failed to update settings:', e)
  }
}

function goHome() {
  router.push('/')
}

// Called by GameBoard before modifying a solved solution
// If it's the best submitted solution, show a blocking dialog
function onBeforeModifyBest(solutionIndex: number, action: () => void) {
  // Check if this solution is the best submitted one
  if (gameActions.isBestSubmittedSolution(solutionIndex)) {
    // Block the action and show info dialog
    showProtectedSolutionDialog.value = true
  } else {
    // Allow the action for non-best solutions
    action()
  }
}

function dismissProtectedDialog() {
  showProtectedSolutionDialog.value = false
}

// Submit solution when puzzle is solved (or improved)
watch(
  () => gameStore.isSolved,
  (solved) => {
    if (solved && hasGame.value && !gameEnded.value) {
      gameActions.submitSolution()
    }
  }
)

// Handle dialog keyboard events at window level
function protectedDialogKeyHandler(event: KeyboardEvent) {
  if (!showProtectedSolutionDialog.value) return
  if (event.key === 'Escape' || event.key === 'Enter') {
    event.preventDefault()
    event.stopPropagation()
    dismissProtectedDialog()
  }
}

watch(showProtectedSolutionDialog, (show) => {
  if (show) {
    window.addEventListener('keydown', protectedDialogKeyHandler, true)
  } else {
    window.removeEventListener('keydown', protectedDialogKeyHandler, true)
  }
})

// Leaderboard toggle
function toggleLeaderboard() {
  showLeaderboard.value = !showLeaderboard.value
}

// Long-press gear to toggle showSolverMoveCount
let gearLongPressTimer: ReturnType<typeof setTimeout> | null = null
let gearLongPressFired = false

function onGearPointerDown() {
  gearLongPressFired = false
  gearLongPressTimer = setTimeout(() => {
    gearLongPressFired = true
    const current = room.value?.settings?.showSolverMoveCount ?? true
    updateSettings({
      showSolverMoveCount: !current,
      showSolverSolutions: room.value?.settings?.showSolverSolutions ?? true,
      minSolutionLength: room.value?.settings?.minSolutionLength ?? 1,
    })
  }, 500)
}

function onGearPointerUp() {
  if (gearLongPressTimer) {
    clearTimeout(gearLongPressTimer)
    gearLongPressTimer = null
  }
}

function onGearClick(e: Event) {
  if (gearLongPressFired) {
    e.preventDefault()
    return
  }
  showSettings.value = true
}

// Long-press solver move count to hide it
let solverLongPressTimer: ReturnType<typeof setTimeout> | null = null

function onSolverPointerDown() {
  solverLongPressTimer = setTimeout(() => {
    solverLongPressTimer = null
    updateSettings({
      showSolverMoveCount: false,
      showSolverSolutions: room.value?.settings?.showSolverSolutions ?? true,
      minSolutionLength: room.value?.settings?.minSolutionLength ?? 1,
    })
  }, 500)
}

function onSolverPointerUp() {
  if (solverLongPressTimer) {
    clearTimeout(solverLongPressTimer)
    solverLongPressTimer = null
  }
}

// Leave game
function promptLeaveGame() {
  showLeaveConfirm.value = true
}

async function confirmLeave() {
  showLeaveConfirm.value = false
  if (roomStore.currentSessionToken) {
    try {
      await bounceBotClient.leaveRoom({
        roomId: normalizedRoomId.value,
        sessionToken: roomStore.currentSessionToken,
      })
    } catch (e) {
      // Best-effort: server will clean up via disconnect timeout if this fails
    }
  }
  roomStore.clearRoom()
  router.push('/')
}

function cancelLeave() {
  showLeaveConfirm.value = false
}

// Boot player (host only)
async function bootPlayer(targetPlayerId: string) {
  if (!roomStore.currentSessionToken) return

  try {
    await bounceBotClient.bootPlayer({
      roomId: normalizedRoomId.value,
      sessionToken: roomStore.currentSessionToken,
      targetPlayerId,
    })
    // If host booted themselves, redirect to home
    if (targetPlayerId === roomStore.currentPlayerId) {
      roomStore.clearRoom()
      router.push('/')
    }
    // Otherwise, room will update via player_left WebSocket event
  } catch (e) {
    console.error('Failed to boot player:', e)
  }
}

function globalKeyHandler(event: KeyboardEvent) {
  // Don't handle if any dialog is open
  if (showProtectedSolutionDialog.value || showLeaveConfirm.value || showSettings.value || showShareModal.value || showInviteModal.value) return

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

function settingsKeyHandler(event: KeyboardEvent) {
  if (!showSettings.value) return
  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    showSettings.value = false
  }
}

watch(showSettings, (show) => {
  if (show) {
    window.addEventListener('keydown', settingsKeyHandler, true)
  } else {
    window.removeEventListener('keydown', settingsKeyHandler, true)
  }
})

function shareModalKeyHandler(event: KeyboardEvent) {
  if (!showShareModal.value) return
  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    showShareModal.value = false
  }
}

watch(showShareModal, (show) => {
  if (show) {
    window.addEventListener('keydown', shareModalKeyHandler, true)
  } else {
    window.removeEventListener('keydown', shareModalKeyHandler, true)
  }
})

function inviteModalKeyHandler(event: KeyboardEvent) {
  if (!showInviteModal.value) return
  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    showInviteModal.value = false
  }
}

watch(showInviteModal, (show) => {
  if (show) {
    window.addEventListener('keydown', inviteModalKeyHandler, true)
  } else {
    window.removeEventListener('keydown', inviteModalKeyHandler, true)
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

      <div v-if="wasRemovedFromRoom" class="removed-message">
        You were disconnected for too long and removed from the room.
        Please rejoin to continue playing.
      </div>

      <div class="card">
        <div class="players-section">
          <h3>Players in room</h3>
          <PlayersPanel :players="room.players" show-host />
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

        <div class="join-row">
          <button
            class="btn primary join-btn"
            :disabled="isJoining"
            @click="joinRoom"
          >
            {{ isJoining ? 'Joining...' : 'Join Room' }}
          </button>
          <span class="room-id-display">Room ID: {{ room.id }}</span>
        </div>
      </div>
    </div>

    <!-- Game in progress (but not for pending players) -->
    <div v-else-if="hasGame && room && !isPendingPlayer" class="game-wrapper">
      <GameBoard
        :on-before-modify-best="onBeforeModifyBest"
        :game-ended="gameEnded"
        :player-solutions="isSinglePlayer ? soloPlayerSolution : sortedSolutions"
        :current-player-solution="currentPlayerSolution"
        :top-three-solutions="topThreeSolutions"
        :solver-solutions="solverPlayerSolutions"
        :get-player-name="getPlayerName"
        :get-player-color="getPlayerColorById"
        :is-solver-solution="isSolverSolution"
        :game-started-at="room.gameStartedAt"
        :room-id="room.id"
        :game-number="displayedGameNumber"
        :input-blocked="showLeaderboard"
        :single-player="isSinglePlayer"
        :get-best-submitted-index="gameActions.getBestSubmittedIndex"
        :on-solution-deleted="gameActions.notifySolutionDeleted"
        :player-color="currentPlayerColor"
        @share="openShareModal"
      >
        <template #header="{ toggleHelp }">
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
              <div v-if="minSolverMoves !== null" class="solver-status" @pointerdown="onSolverPointerDown" @pointerup="onSolverPointerUp" @pointerleave="onSolverPointerUp" @contextmenu.prevent>
                <img src="/favicon_dark.svg" alt="" class="solver-icon" />
                <span class="solver-moves">{{ minSolverMoves }}</span>
              </div>
              <span v-if="gameStore.isSolved" class="solved-indicator">✓</span>
              <button class="btn-icon settings-btn-header" @click="onGearClick" @pointerdown="onGearPointerDown" @pointerup="onGearPointerUp" @pointerleave="onGearPointerUp" @contextmenu.prevent title="Settings">
                <img src="/gear.svg" alt="Settings" class="gear-icon" />
              </button>
              <button class="help-btn-header" @click="toggleHelp" title="How to Play">?</button>
              <button
                class="btn primary next-puzzle-btn"
                :disabled="isStarting"
                @click="nextPuzzle"
              >
                {{ isStarting ? 'Loading...' : (gameEnded || !room?.settings?.showSolverSolutions ? 'Next Puzzle' : "I'm Finished") }}
              </button>
            </template>
            <!-- Multiplayer mode: full header -->
            <template v-else-if="!gameEnded">
              <PlayersPanel :players="room.players" :solutions="room.solutions" :scores="room.scores" :game-started-at="room.gameStartedAt" :finished-solving="room.finishedSolving" compact />
              <div v-if="minSolverMoves !== null" class="solver-status" @pointerdown="onSolverPointerDown" @pointerup="onSolverPointerUp" @pointerleave="onSolverPointerUp" @contextmenu.prevent>
                <img src="/favicon_dark.svg" alt="" class="solver-icon" />
                <span class="solver-moves">{{ minSolverMoves }}</span>
              </div>
              <button class="btn-icon help-btn-header" @click="toggleHelp" title="How to Play">?</button>
              <button v-if="isRoomCreator" class="btn-icon settings-btn-header" @click="onGearClick" @pointerdown="onGearPointerDown" @pointerup="onGearPointerUp" @pointerleave="onGearPointerUp" @contextmenu.prevent title="Room Settings">
                <img src="/gear.svg" alt="Settings" class="gear-icon" />
              </button>
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
                :show="showLeaderboard"
                :players="room.players"
                :scores="room.scores"
                :games-played="room.gamesPlayed"
                @close="toggleLeaderboard"
              />
              <button class="btn-icon help-btn-header" @click="toggleHelp" title="How to Play">?</button>
              <button v-if="isRoomCreator" class="btn-icon settings-btn-header" @click="onGearClick" @pointerdown="onGearPointerDown" @pointerup="onGearPointerUp" @pointerleave="onGearPointerUp" @contextmenu.prevent title="Room Settings">
                <img src="/gear.svg" alt="Settings" class="gear-icon" />
              </button>
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

    <!-- Pending player watching the in-progress round they'll join next -->
    <div v-else-if="room && isPendingPlayer" class="game-wrapper">
      <GameBoard
        spectator-mode
        :game-ended="gameEnded"
        :player-solutions="sortedSolutions"
        :current-player-solution="currentPlayerSolution"
        :top-three-solutions="topThreeSolutions"
        :solver-solutions="solverPlayerSolutions"
        :is-solver-solution="isSolverSolution"
        :get-player-name="getPlayerName"
        :get-player-color="getPlayerColorById"
        :game-started-at="room.gameStartedAt"
        :room-id="room.id"
        :game-number="displayedGameNumber"
        @share="openShareModal"
      >
        <template #header>
          <div class="game-header">
            <PlayersPanel :players="room.players" :solutions="room.solutions" :scores="room.scores" :game-started-at="room.gameStartedAt" :finished-solving="room.finishedSolving" compact />
            <div v-if="minSolverMoves !== null" class="solver-status">
              <img src="/favicon_dark.svg" alt="" class="solver-icon" />
              <span class="solver-moves">{{ minSolverMoves }}</span>
            </div>
          </div>
        </template>
      </GameBoard>

      <div class="spectator-footer">
        <div class="room-info">
          <div class="info-row">
            <span class="label">Room ID</span>
            <code class="room-id">{{ room.id }}</code>
          </div>
          <div class="room-actions">
            <button class="btn-small" @click="showInviteModal = true">Invite</button>
          </div>
        </div>

        <div v-if="room.pendingPlayers.length > 1" class="players-section">
          <h3>Also waiting to join</h3>
          <PlayersPanel :players="room.pendingPlayers" hide-waiting-message />
        </div>
      </div>
    </div>

    <!-- Waiting room -->
    <div v-else-if="room && hasJoined" class="waiting-room">
      <h1 class="waiting-title">WAITING <span class="room-text">ROOM</span></h1>

      <div class="card">
        <div class="room-info">
          <div class="info-row">
            <span class="label">Room ID</span>
            <code class="room-id">{{ room.id }}</code>
          </div>
          <div class="room-actions">
            <button class="btn-small" @click="showInviteModal = true">Invite</button>
            <button v-if="isRoomCreator" class="btn-small settings-btn" @click="showSettings = true" title="Room Settings">
              <img src="/gear.svg" alt="Settings" class="gear-icon" />
            </button>
          </div>
        </div>

        <div class="players-section">
          <h3>Players ({{ room.players.length }})</h3>
          <PlayersPanel :players="room.players" show-host />
        </div>

        <div v-if="error" class="error">{{ error }}</div>

        <div class="start-options">
          <button
            v-if="isSinglePlayer || isRoomCreator"
            class="btn primary start-btn"
            :disabled="isStarting"
            @click="startGame"
          >
            {{ isStarting ? 'Starting...' : 'Start Game' }}
          </button>
          <p v-else class="waiting-text">Waiting for room creator to start game...</p>
        </div>

        <p class="hint">Click Invite to share this room with friends!</p>
      </div>
    </div>

    <!-- Best solution protected dialog -->
    <div v-if="showProtectedSolutionDialog" class="dialog-overlay" @click.self="dismissProtectedDialog">
      <div class="dialog">
        <h3>Solution Protected</h3>
        <p>
          This is your best submitted solution. You can't undo or delete it,
          but you can create another solution to try a different strategy.
        </p>
        <div class="dialog-actions">
          <button class="btn primary" @click="dismissProtectedDialog">OK</button>
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

    <!-- Leaderboard modal (only during game, not for pending players) -->
    <LeaderboardModal
      v-if="room?.currentGame && hasJoined && !isPendingPlayer"
      :show="showLeaderboard"
      :players="room?.players ?? []"
      :scores="room?.scores ?? []"
      :games-played="room?.gamesPlayed ?? 0"
      @close="showLeaderboard = false"
    />

    <!-- Settings modal (only for room creator) -->
    <SettingsModal
      :show="showSettings"
      :settings="room?.settings"
      :players="room?.players ?? []"
      :show-boot-player="!isSinglePlayer"
      @close="showSettings = false"
      @update="updateSettings"
      @boot-player="bootPlayer"
    />

    <!-- Share board modal -->
    <ShareModal
      :show="showShareModal"
      :url="shareBoardUrl"
      @close="showShareModal = false"
    />

    <!-- Invite players modal -->
    <ShareModal
      :show="showInviteModal"
      :url="shareUrl"
      title="Invite Players"
      :code="room?.id"
      @close="showInviteModal = false"
    />
  </div>
</template>

<style scoped>
@font-face {
  font-family: 'Conthrax';
  src: url('/fonts/ConthraxRg-Bold.eot');
  src: url('/fonts/ConthraxRg-Bold.eot?#iefix') format('embedded-opentype'),
      url('/fonts/ConthraxRg-Bold.woff2') format('woff2'),
      url('/fonts/ConthraxRg-Bold.woff') format('woff'),
      url('/fonts/ConthraxRg-Bold.ttf') format('truetype'),
      url('/fonts/ConthraxRg-Bold.otf') format('opentype');
  font-weight: bold;
  font-style: normal;
  font-display: swap;
}

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
  width: 185vmax;
  height: 185vmax;
  background-image: url('/pattern_dark.svg');
  background-repeat: no-repeat;
  background-size: auto 100%;
  background-position: center;
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

.spectator-footer {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 1.5rem 2rem;
  width: calc(100% - 2rem);
  max-width: 360px;
  margin-top: 1.5rem;
  box-sizing: border-box;
}

.spectator-footer .room-info {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.spectator-footer .players-section {
  margin-top: 1.5rem;
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

.solver-status {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.3rem 0.6rem;
  background: #242424;
  border: 1px solid #333;
  border-radius: 6px;
  font-size: 0.8rem;
  color: #aaa;
}

.solver-status .solver-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

.solver-moves {
  font-weight: 600;
  color: #fff;
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

.removed-message {
  background: #5c3a1e;
  border: 1px solid #8b5a2b;
  color: #ffd699;
  padding: 0.75rem 1rem;
  border-radius: 6px;
  margin-bottom: 1rem;
  text-align: center;
  max-width: 400px;
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

.join-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-top: 0.5rem;
}

.join-btn {
  padding: 1rem 2rem;
  font-size: 1.1rem;
}

.room-id-display {
  font-size: 0.8rem;
  color: #888;
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

.waiting-title {
  font-family: 'Conthrax', sans-serif;
  color: #fff;
  margin: 0 0 1.5rem;
  font-size: 2.25rem;
  text-align: center;
  white-space: nowrap;
}

.room-text {
  color: #1e88e5;
}

@media (prefers-color-scheme: light) {
  .waiting-title {
    color: #000;
  }

  .removed-message {
    background: #fff3cd;
    border-color: #ffc107;
    color: #856404;
  }
}

.card {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 2rem;
  width: calc(100% - 2rem);
  max-width: 360px;
}

.room-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid #333;
  padding-bottom: 1.5rem;
  margin-bottom: 1.5rem;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.label {
  color: #aaa;
  font-size: 0.9rem;
}

.room-id {
  background: #1e88e5;
  padding: 0.4rem 0.75rem;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  color: #fff;
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

.room-actions {
  display: flex;
  gap: 0.5rem;
}

.settings-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.4rem;
}

.gear-icon {
  width: 16px;
  height: 16px;
  opacity: 0.8;
}

.settings-btn:hover .gear-icon {
  opacity: 1;
}

.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.4rem;
  background: transparent;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-icon:hover {
  background: rgba(255, 255, 255, 0.1);
}

.help-btn-header {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 1.5px solid #888;
  background: transparent;
  color: #888;
  font-size: 0.8rem;
  font-weight: bold;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  margin-left: 0.25rem;
  transition: color 0.15s, border-color 0.15s;
}

.help-btn-header:hover {
  color: #fff;
  border-color: #fff;
}

.settings-btn-header .gear-icon {
  width: 18px;
  height: 18px;
  opacity: 0.6;
}

.settings-btn-header:hover .gear-icon {
  opacity: 1;
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

.waiting-text {
  color: #888;
  text-align: center;
  padding: 1rem;
  font-style: italic;
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
  color: #999;
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
