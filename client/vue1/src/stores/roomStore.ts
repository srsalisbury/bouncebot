import { ref, watch } from 'vue'
import { defineStore } from 'pinia'

const STORAGE_KEY_NAME = 'bouncebot_player_name'
const STORAGE_KEY_ID = 'bouncebot_player_id'
const STORAGE_KEY_LAST_ROOM = 'bouncebot_last_room'
const STORAGE_KEY_SINGLE_PLAYER = 'bouncebot_single_player'
const STORAGE_KEY_PUZZLES_SOLVED = 'bouncebot_puzzles_solved'
const STORAGE_KEY_PUZZLES_ATTEMPTED = 'bouncebot_puzzles_attempted'

export const useRoomStore = defineStore('room', () => {
  // Load from localStorage on init
  const storedName = localStorage.getItem(STORAGE_KEY_NAME)
  const storedId = localStorage.getItem(STORAGE_KEY_ID)
  const storedLastRoom = localStorage.getItem(STORAGE_KEY_LAST_ROOM)
  const storedSinglePlayer = localStorage.getItem(STORAGE_KEY_SINGLE_PLAYER)
  const storedPuzzlesSolved = localStorage.getItem(STORAGE_KEY_PUZZLES_SOLVED)
  const storedPuzzlesAttempted = localStorage.getItem(STORAGE_KEY_PUZZLES_ATTEMPTED)

  const currentPlayerName = ref<string | null>(storedName)
  const currentPlayerId = ref<string | null>(storedId)
  const lastRoomId = ref<string | null>(storedLastRoom)
  const isSinglePlayer = ref(storedSinglePlayer === 'true')
  const puzzlesSolved = ref(storedPuzzlesSolved ? parseInt(storedPuzzlesSolved, 10) : 0)
  const puzzlesAttempted = ref(storedPuzzlesAttempted ? parseInt(storedPuzzlesAttempted, 10) : 0)

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

  watch(isSinglePlayer, (value) => {
    if (value) {
      localStorage.setItem(STORAGE_KEY_SINGLE_PLAYER, 'true')
    } else {
      localStorage.removeItem(STORAGE_KEY_SINGLE_PLAYER)
    }
  })

  watch(puzzlesSolved, (count) => {
    localStorage.setItem(STORAGE_KEY_PUZZLES_SOLVED, count.toString())
  })

  watch(puzzlesAttempted, (count) => {
    localStorage.setItem(STORAGE_KEY_PUZZLES_ATTEMPTED, count.toString())
  })

  function setCurrentPlayer(id: string, name: string) {
    currentPlayerId.value = id
    currentPlayerName.value = name
  }

  function setLastRoom(roomId: string) {
    lastRoomId.value = roomId
  }

  function setSinglePlayer(value: boolean) {
    isSinglePlayer.value = value
    if (value) {
      puzzlesSolved.value = 0
      puzzlesAttempted.value = 0
    }
  }

  function recordPuzzleSolved() {
    puzzlesSolved.value++
  }

  function recordPuzzleAttempted() {
    puzzlesAttempted.value++
  }

  function clearLastRoom() {
    lastRoomId.value = null
  }

  function clear() {
    currentPlayerId.value = null
    currentPlayerName.value = null
    lastRoomId.value = null
    isSinglePlayer.value = false
  }

  return {
    currentPlayerId,
    currentPlayerName,
    lastRoomId,
    isSinglePlayer,
    puzzlesSolved,
    puzzlesAttempted,
    setCurrentPlayer,
    setLastRoom,
    clearLastRoom,
    setSinglePlayer,
    recordPuzzleSolved,
    recordPuzzleAttempted,
    clear,
  }
})
