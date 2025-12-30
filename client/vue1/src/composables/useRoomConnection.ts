import { ref, computed, watch, onMounted, onUnmounted, type Ref } from 'vue'
import { bounceBotClient } from '../services/connectClient'
import { websocketService, type WebSocketEvent } from '../services/websocket'
import { useRoomStore } from '../stores/roomStore'
import type { Room } from '../gen/bouncebot_pb'

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

  const normalizedRoomId = computed(() => roomId.value.toUpperCase())
  const hasGame = computed(() => room.value?.currentGame != null)
  const hasJoined = computed(() => roomStore.currentPlayerId != null)

  async function loadRoom(forceApplyGame = false) {
    try {
      const rm = await bounceBotClient.getRoom({ roomId: normalizedRoomId.value })
      const hadGame = hasGame.value
      room.value = rm

      // Check if current player is still in the room (handle stale localStorage)
      if (roomStore.currentPlayerId) {
        const isPlayerInRoom = rm.players.some(p => p.id === roomStore.currentPlayerId)
        if (!isPlayerInRoom) {
          roomStore.clear()
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

      error.value = null
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load room'
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
      const rm = await bounceBotClient.joinRoom({
        roomId: normalizedRoomId.value,
        playerName: playerName.trim(),
      })
      const player = rm.players[rm.players.length - 1]
      if (player) {
        roomStore.setCurrentPlayer(player.id, player.name)
      }
      await loadRoom()
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to join room'
      return false
    }
  }

  async function startGame(useFixedBoard = false) {
    error.value = null

    try {
      const rm = await bounceBotClient.startGame({ roomId: normalizedRoomId.value, useFixedBoard })
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
      loadRoom(true)
    } else if (event.type === 'player_solved') {
      loadRoom()
    } else if (event.type === 'solution_retracted') {
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
    }
  }

  function connectWebSocket() {
    if (hasJoined.value && roomStore.currentPlayerId) {
      websocketService.connect(normalizedRoomId.value, roomStore.currentPlayerId, handleWebSocketEvent)
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
    loadRoom,
    joinRoom,
    startGame,
  }
}
