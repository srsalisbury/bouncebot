import { ref, watch } from 'vue'
import { defineStore } from 'pinia'

const STORAGE_KEY_NAME = 'bouncebot_player_name'
const STORAGE_KEY_ID = 'bouncebot_player_id'
const STORAGE_KEY_SESSION_TOKEN = 'bouncebot_session_token'
const STORAGE_KEY_LAST_ROOM = 'bouncebot_last_room'
const STORAGE_KEY_SINGLE_PLAYER = 'bouncebot_single_player'
const STORAGE_KEY_PUZZLES_SOLVED = 'bouncebot_puzzles_solved'
const STORAGE_KEY_PUZZLES_ATTEMPTED = 'bouncebot_puzzles_attempted'
const STORAGE_KEY_SHOW_SOLVER_MOVE_COUNT = 'bouncebot_show_solver_move_count'
const STORAGE_KEY_SHOW_SOLVER_SOLUTIONS = 'bouncebot_show_solver_solutions'
const STORAGE_KEY_MIN_SOLUTION_LENGTH = 'bouncebot_min_solution_length'

// Must match server/room.MinMinSolutionLength (the no-op default).
const DEFAULT_MIN_SOLUTION_LENGTH = 1

export const useRoomStore = defineStore('room', () => {
  // Load from localStorage on init
  const storedName = localStorage.getItem(STORAGE_KEY_NAME)
  const storedId = localStorage.getItem(STORAGE_KEY_ID)
  const storedSessionToken = localStorage.getItem(STORAGE_KEY_SESSION_TOKEN)
  const storedLastRoom = localStorage.getItem(STORAGE_KEY_LAST_ROOM)
  const storedSinglePlayer = localStorage.getItem(STORAGE_KEY_SINGLE_PLAYER)
  const storedPuzzlesSolved = localStorage.getItem(STORAGE_KEY_PUZZLES_SOLVED)
  const storedPuzzlesAttempted = localStorage.getItem(STORAGE_KEY_PUZZLES_ATTEMPTED)
  const storedShowSolverMoveCount = localStorage.getItem(STORAGE_KEY_SHOW_SOLVER_MOVE_COUNT)
  const storedShowSolverSolutions = localStorage.getItem(STORAGE_KEY_SHOW_SOLVER_SOLUTIONS)
  const storedMinSolutionLength = localStorage.getItem(STORAGE_KEY_MIN_SOLUTION_LENGTH)

  const currentPlayerName = ref<string | null>(storedName)
  const currentPlayerId = ref<string | null>(storedId)
  const currentSessionToken = ref<string | null>(storedSessionToken)
  const lastRoomId = ref<string | null>(storedLastRoom)
  const isSinglePlayer = ref(storedSinglePlayer === 'true')
  const puzzlesSolved = ref(storedPuzzlesSolved ? parseInt(storedPuzzlesSolved, 10) : 0)
  const puzzlesAttempted = ref(storedPuzzlesAttempted ? parseInt(storedPuzzlesAttempted, 10) : 0)
  const showSolverMoveCount = ref(storedShowSolverMoveCount === 'true')
  const showSolverSolutions = ref(storedShowSolverSolutions === 'true')
  const minSolutionLength = ref(
    storedMinSolutionLength ? parseInt(storedMinSolutionLength, 10) : DEFAULT_MIN_SOLUTION_LENGTH
  )

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

  watch(currentSessionToken, (token) => {
    if (token) {
      localStorage.setItem(STORAGE_KEY_SESSION_TOKEN, token)
    } else {
      localStorage.removeItem(STORAGE_KEY_SESSION_TOKEN)
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

  watch(showSolverMoveCount, (value) => {
    localStorage.setItem(STORAGE_KEY_SHOW_SOLVER_MOVE_COUNT, value.toString())
  })

  watch(showSolverSolutions, (value) => {
    localStorage.setItem(STORAGE_KEY_SHOW_SOLVER_SOLUTIONS, value.toString())
  })

  watch(minSolutionLength, (value) => {
    localStorage.setItem(STORAGE_KEY_MIN_SOLUTION_LENGTH, value.toString())
  })

  function setCurrentPlayer(id: string, name: string, sessionToken: string) {
    currentPlayerId.value = id
    currentPlayerName.value = name
    currentSessionToken.value = sessionToken
  }

  function setCurrentPlayerId(id: string, sessionToken: string) {
    currentPlayerId.value = id
    currentSessionToken.value = sessionToken
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

  function setSettings(settings: {
    showSolverMoveCount: boolean
    showSolverSolutions: boolean
    minSolutionLength: number
  }) {
    showSolverMoveCount.value = settings.showSolverMoveCount
    showSolverSolutions.value = settings.showSolverSolutions
    minSolutionLength.value = settings.minSolutionLength
  }

  function clearLastRoom() {
    lastRoomId.value = null
  }

  function clearRoom() {
    currentPlayerId.value = null
    currentSessionToken.value = null
    lastRoomId.value = null
    isSinglePlayer.value = false
    // Note: currentPlayerName is intentionally preserved for future sessions
  }

  return {
    currentPlayerId,
    currentPlayerName,
    currentSessionToken,
    lastRoomId,
    isSinglePlayer,
    puzzlesSolved,
    puzzlesAttempted,
    showSolverMoveCount,
    showSolverSolutions,
    minSolutionLength,
    setCurrentPlayer,
    setCurrentPlayerId,
    setLastRoom,
    clearLastRoom,
    clearRoom,
    setSinglePlayer,
    recordPuzzleSolved,
    recordPuzzleAttempted,
    setSettings,
  }
})
