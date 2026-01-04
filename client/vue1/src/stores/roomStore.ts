import { ref, watch } from 'vue'
import { defineStore } from 'pinia'

const STORAGE_KEY_NAME = 'bouncebot_player_name'
const STORAGE_KEY_ID = 'bouncebot_player_id'
const STORAGE_KEY_LAST_ROOM = 'bouncebot_last_room'

export const useRoomStore = defineStore('room', () => {
  // Load from localStorage on init
  const storedName = localStorage.getItem(STORAGE_KEY_NAME)
  const storedId = localStorage.getItem(STORAGE_KEY_ID)
  const storedLastRoom = localStorage.getItem(STORAGE_KEY_LAST_ROOM)
  const currentPlayerName = ref<string | null>(storedName)
  const currentPlayerId = ref<string | null>(storedId)
  const lastRoomId = ref<string | null>(storedLastRoom)

  // Persist to localStorage when changed
  watch(currentPlayerName, (name) => {
    if (name) {
      localStorage.setItem(STORAGE_KEY_NAME, name)
    } else {
      localStorage.removeItem(STORAGE_KEY_NAME)
    }
  })

  watch(currentPlayerId, (id) => {
    if (id) {
      localStorage.setItem(STORAGE_KEY_ID, id)
    } else {
      localStorage.removeItem(STORAGE_KEY_ID)
    }
  })

  watch(lastRoomId, (roomId) => {
    if (roomId) {
      localStorage.setItem(STORAGE_KEY_LAST_ROOM, roomId)
    } else {
      localStorage.removeItem(STORAGE_KEY_LAST_ROOM)
    }
  })

  function setCurrentPlayer(id: string, name: string) {
    currentPlayerId.value = id
    currentPlayerName.value = name
  }

  function setLastRoom(roomId: string) {
    lastRoomId.value = roomId
  }

  function clear() {
    currentPlayerId.value = null
    currentPlayerName.value = null
  }

  return {
    currentPlayerId,
    currentPlayerName,
    lastRoomId,
    setCurrentPlayer,
    setLastRoom,
    clear,
  }
})
