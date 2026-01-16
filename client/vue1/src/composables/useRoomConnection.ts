import { ref, computed, watch, onMounted, onUnmounted, type Ref } from 'vue'
import { bounceBotClient } from '../services/connectClient'
import { translateJoinRoomError } from '../services/errorMessages'
import { websocketService, type WebSocketEvent, type SolverCompletePayload } from '../services/websocket'
import { useRoomStore } from '../stores/roomStore'
import { isRoomNotFoundError } from '../services/errorUtils'
import type { Room } from '../gen/bouncebot_pb'

export interface SolverSolution {
  solverName: string
  moves: { robotId: number; x: number; y: number }[]
  completed: boolean
}

export interface RoomConnectionOptions {
  roomId: Ref<string>
  onGameStarted?: () => void
  onGameEnded?: () => void
  onRoomUpdated?: (room: Room) => void
}

export function useRoomConnection(options: RoomConnectionOptions) {
  const { roomId, onGameStarted, onGameEnded, onRoomUpdated } = options

  const roomStore = useRoomStore()

  const room = ref<Room | null>(null)
  const isLoading = ref(true)
  const error = ref<string | null>(null)
  const pollInterval = ref<number | null>(null)
  const solverSolutions = ref<SolverSolution[]>([])

  const normalizedRoomId = computed(() => roomId.value.toUpperCase())
  const hasGame = computed(() => room.value?.currentGame != null)
  const hasJoined = computed(() => roomStore.currentPlayerId != null)

  async function loadRoom(forceApplyGame = false) {
    try {
      const rm = await bounceBotClient.getRoom({ roomId: normalizedRoomId.value })
      const hadGame = hasGame.value
      room.value = rm

      // Remember this room for easy return if user navigates away
      roomStore.setLastRoom(normalizedRoomId.value)

      // Check if current player is still in the room (handle stale localStorage)
      if (roomStore.currentPlayerId) {
        const player = rm.players.find(p => p.id === roomStore.currentPlayerId)
          || rm.pendingPlayers.find(p => p.id === roomStore.currentPlayerId)
        if (player) {
          // Re-save player info to ensure it's persisted (e.g., on page reload in lobby)
          // For solo mode, only save the ID to avoid overwriting saved multiplayer name
          if (rm.isSinglePlayer) {
            roomStore.setCurrentPlayerId(player.id)
          } else {
            roomStore.setCurrentPlayer(player.id, player.name)
          }
        } else {
          roomStore.clearRoom()
        }
      }

      // Notify when game first appears or when forced
      if (rm.currentGame && (!hadGame || forceApplyGame)) {
        onRoomUpdated?.(rm)
        if (pollInterval.value) {
          clearInterval(pollInterval.value)
          pollInterval.value = null
        }
      }

      // Restore solver solutions from room state (for page reloads)
      if (rm.solverResults && rm.solverResults.length > 0) {
        solverSolutions.value = rm.solverResults
          .filter(sr => sr.completed && sr.moves.length > 0)
          .map(sr => ({
            solverName: sr.solverName,
            moves: sr.moves.map(m => ({
              robotId: m.id,
              x: m.pos?.x ?? 0,
              y: m.pos?.y ?? 0,
            })),
            completed: sr.completed,
          }))
      }

      error.value = null
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load room'
      // If server says room doesn't exist, clear stored room state (preserve player name)
      if (isRoomNotFoundError(e)) {
        roomStore.clearRoom()
      }
    } finally {
      isLoading.value = false
    }
  }

  async function joinRoom(playerName: string) {
    if (!playerName.trim()) {
      error.value = 'Please enter your name'
      return false
    }

    error.value = null

    try {
      const trimmedName = playerName.trim()
      const response = await bounceBotClient.joinRoom({
        roomId: normalizedRoomId.value,
        playerName: trimmedName,
      })

      // Update room.value BEFORE setting currentPlayerId to avoid race condition
      // where hasJoined becomes true but isPendingPlayer is still false
      room.value = response.room!

      // Use the player ID from the response directly
      roomStore.setCurrentPlayer(response.playerId, trimmedName)

      // Still call loadRoom to trigger game state callbacks (onRoomUpdated)
      await loadRoom(true)
      return true
    } catch (e) {
      error.value = translateJoinRoomError(e)
      return false
    }
  }

  async function startGame() {
    error.value = null

    try {
      const rm = await bounceBotClient.startGame({ roomId: normalizedRoomId.value })
      room.value = rm
      onRoomUpdated?.(rm)
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to start game'
      return false
    }
  }

  function handleWebSocketEvent(event: WebSocketEvent) {
    if (event.type === 'player_joined') {
      loadRoom()
    } else if (event.type === 'game_started') {
      onGameStarted?.()
      solverSolutions.value = [] // Clear solver solutions when new game starts
      loadRoom(true)
    } else if (event.type === 'player_solved') {
      loadRoom()
    } else if (event.type === 'player_finished_solving') {
      loadRoom()
    } else if (event.type === 'player_ready_for_next') {
      loadRoom()
    } else if (event.type === 'game_ended') {
      onGameEnded?.()
      loadRoom()
    } else if (event.type === 'player_left') {
      loadRoom()
    } else if (event.type === 'solver_complete') {
      const payload = event.payload as SolverCompletePayload
      if (payload.completed && payload.moves.length > 0) {
        const newSolution: SolverSolution = {
          solverName: payload.solverName,
          moves: payload.moves,
          completed: payload.completed,
        }
        // Upsert by solver name
        const existingIndex = solverSolutions.value.findIndex(s => s.solverName === payload.solverName)
        if (existingIndex >= 0) {
          solverSolutions.value[existingIndex] = newSolution
        } else {
          solverSolutions.value = [...solverSolutions.value, newSolution]
        }
      }
    } else if (event.type === 'settings_changed') {
      loadRoom()
    }
  }

  function handleWebSocketDisconnect() {
    // When WebSocket disconnects persistently, check if room still exists
    // This will trigger state cleanup if server returns 404
    console.log('WebSocket disconnected, checking if room still exists')
    loadRoom()
  }

  function connectWebSocket() {
    if (hasJoined.value && roomStore.currentPlayerId) {
      websocketService.connect(
        normalizedRoomId.value,
        roomStore.currentPlayerId,
        handleWebSocketEvent,
        handleWebSocketDisconnect
      )
    }
  }

  // Connect to WebSocket when user joins
  watch(hasJoined, (joined) => {
    if (joined) {
      connectWebSocket()
      if (pollInterval.value) {
        clearInterval(pollInterval.value)
        pollInterval.value = null
      }
    }
  })

  onMounted(async () => {
    await loadRoom()

    if (hasJoined.value) {
      connectWebSocket()
    } else {
      pollInterval.value = window.setInterval(loadRoom, 3000)
    }
  })

  onUnmounted(() => {
    if (pollInterval.value) {
      clearInterval(pollInterval.value)
    }
    websocketService.disconnect()
  })

  return {
    room,
    isLoading,
    error,
    normalizedRoomId,
    hasGame,
    hasJoined,
    solverSolutions,
    loadRoom,
    joinRoom,
    startGame,
  }
}
