import { ref, computed, watch, onMounted, onUnmounted, type Ref } from 'vue'
import { bounceBotClient } from '../services/connectClient'
import { translateJoinRoomError } from '../services/errorMessages'
import { websocketService, type WebSocketEvent, type SolverCompletePayload, type PlayerLeftPayload } from '../services/websocket'
import { useRoomStore } from '../stores/roomStore'
import { isRoomNotFoundError } from '../services/errorUtils'
import { ROOM_POLL_INTERVAL_MS } from '../constants'
import { timestampToSeconds } from '../services/timeUtils'
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
  onPlayerBooted?: () => void
}

export function useRoomConnection(options: RoomConnectionOptions) {
  const { roomId, onGameStarted, onGameEnded, onRoomUpdated, onPlayerBooted } = options

  const roomStore = useRoomStore()

  const room = ref<Room | null>(null)
  const isLoading = ref(true)
  const error = ref<string | null>(null)
  const pollInterval = ref<number | null>(null)
  const solverSolutions = ref<SolverSolution[]>([])
  const wasRemovedFromRoom = ref(false)

  const normalizedRoomId = computed(() => roomId.value.toUpperCase())
  const hasGame = computed(() => room.value?.currentGame != null)
  const hasJoined = computed(() => roomStore.currentPlayerId != null && roomStore.currentSessionToken != null)

  // Guards against out-of-order responses: two events fired in quick
  // succession (e.g. a player's own "ready for next" broadcast, followed
  // shortly by "game started" once board generation finishes) each trigger a
  // loadRoom call, and the network makes no guarantee they resolve in the
  // order they were issued. Without this, a slow response to the earlier
  // call can land after the later one and silently overwrite fresher state
  // with stale data (e.g. reverting a promoted pending player back to
  // "watching" the game that just ended).
  let loadRoomSeq = 0

  async function loadRoom(forceApplyGame = false) {
    const seq = ++loadRoomSeq
    try {
      const rm = await bounceBotClient.getRoom({ roomId: normalizedRoomId.value })
      if (seq !== loadRoomSeq) return // superseded by a newer loadRoom call
      const hadGame = hasGame.value
      // room.currentGame is never nulled out between rounds server-side (it's
      // replaced in place when the next round commits), so hadGame alone can't
      // tell "still the old, finished game" from "a new round already started".
      // gameStartedAt is set fresh on every round, so a change in it is a
      // reliable signal even if this call's forceApplyGame flag got lost
      // because an earlier/later loadRoom call superseded it (see loadRoomSeq
      // comment above).
      const prevGameStartedAt = timestampToSeconds(room.value?.gameStartedAt)
      const isNewRound = timestampToSeconds(rm.gameStartedAt) !== prevGameStartedAt
      room.value = rm

      // Check if current player is still in the room (handle stale localStorage)
      // Must check before updating lastRoomId
      if (roomStore.currentPlayerId) {
        const player = rm.players.find(p => p.id === roomStore.currentPlayerId)
          || rm.pendingPlayers.find(p => p.id === roomStore.currentPlayerId)
        if (!player) {
          // Player not found in room - only show "removed" message if this is the same room
          // they were in before (not when visiting a different room)
          if (roomStore.lastRoomId === normalizedRoomId.value) {
            wasRemovedFromRoom.value = true
          }
          roomStore.clearRoom()
        }
        // If player is found, keep the existing session token from localStorage
        // (getRoom doesn't return session tokens since they're secret)
      }

      // Remember this room for easy return if user navigates away
      roomStore.setLastRoom(normalizedRoomId.value)

      // Notify when game first appears, a new round has started, or when forced
      if (rm.currentGame && (!hadGame || forceApplyGame || isNewRound)) {
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
      if (seq !== loadRoomSeq) return // superseded by a newer loadRoom call
      error.value = e instanceof Error ? e.message : 'Failed to load room'
      // If server says room doesn't exist, clear stored room state (preserve player name)
      if (isRoomNotFoundError(e)) {
        roomStore.clearRoom()
      }
    } finally {
      if (seq === loadRoomSeq) {
        isLoading.value = false
      }
    }
  }

  async function joinRoom(playerName: string) {
    if (!playerName.trim()) {
      error.value = 'Please enter your name'
      return false
    }

    error.value = null
    wasRemovedFromRoom.value = false

    try {
      const trimmedName = playerName.trim()
      const response = await bounceBotClient.joinRoom({
        roomId: normalizedRoomId.value,
        playerName: trimmedName,
      })

      // Update room.value BEFORE setting currentPlayerId to avoid race condition
      // where hasJoined becomes true but isPendingPlayer is still false
      room.value = response.room!

      // Use the player ID and session token from the response
      roomStore.setCurrentPlayer(response.playerId, trimmedName, response.sessionToken)

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
      const payload = event.payload as PlayerLeftPayload
      // Check if the current player was booted
      if (payload.playerId === roomStore.currentPlayerId) {
        onPlayerBooted?.()
        return
      }
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

  async function handleWebSocketDisconnect() {
    // When WebSocket disconnects persistently, check if player is still in the room
    console.log('WebSocket: persistent disconnect, checking room membership')

    try {
      const rm = await bounceBotClient.getRoom({ roomId: normalizedRoomId.value })
      const stillInRoom = rm.players.some(p => p.id === roomStore.currentPlayerId)
        || rm.pendingPlayers.some(p => p.id === roomStore.currentPlayerId)

      if (!stillInRoom && roomStore.currentPlayerId) {
        // Player was removed from room (disconnect timeout)
        console.log('Player was removed from room due to disconnect timeout')
        wasRemovedFromRoom.value = true
        roomStore.clearRoom()
        room.value = rm // Update room state so UI shows current players
      } else {
        // Still in room, just network issues - reload and continue
        room.value = rm
      }
    } catch (e) {
      // Room may not exist anymore
      if (isRoomNotFoundError(e)) {
        roomStore.clearRoom()
      }
    }
  }

  function connectWebSocket() {
    if (hasJoined.value && roomStore.currentSessionToken) {
      websocketService.connect(
        normalizedRoomId.value,
        roomStore.currentSessionToken,
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

  function handleVisibilityChange() {
    // When page becomes visible again (e.g., phone wakes up), check if still in room
    if (document.visibilityState === 'visible' && hasJoined.value) {
      loadRoom()
    }
  }

  onMounted(async () => {
    await loadRoom()

    if (hasJoined.value) {
      connectWebSocket()
    } else {
      pollInterval.value = window.setInterval(loadRoom, ROOM_POLL_INTERVAL_MS)
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
  })

  onUnmounted(() => {
    if (pollInterval.value) {
      clearInterval(pollInterval.value)
    }
    websocketService.disconnect()
    document.removeEventListener('visibilitychange', handleVisibilityChange)
  })

  return {
    room,
    isLoading,
    error,
    normalizedRoomId,
    hasGame,
    hasJoined,
    solverSolutions,
    wasRemovedFromRoom,
    loadRoom,
    joinRoom,
    startGame,
  }
}
