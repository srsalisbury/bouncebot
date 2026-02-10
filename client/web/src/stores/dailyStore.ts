import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { bounceBotClient } from '../services/connectClient'
import type { DailyPuzzleInfo, BotPos, Game } from '../gen/bouncebot_pb'

const STORAGE_KEY_DAILY_PLAYER_ID = 'bouncebot_daily_player_id'

function generateUUID(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID()
  }
  // Fallback for non-secure contexts
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

export interface DailyPuzzle {
  difficulty: string
  game: Game
  solved: boolean
  optimalMoves: number
}

export interface DailyState {
  date: string
  puzzles: DailyPuzzle[]
  secondsUntilReset: number
  loading: boolean
  error: string | null
}

export const useDailyStore = defineStore('daily', () => {
  // Get or create a player ID for daily challenges
  const storedPlayerId = localStorage.getItem(STORAGE_KEY_DAILY_PLAYER_ID)
  const playerId = ref<string>(storedPlayerId || generateUUID())

  // Persist player ID
  if (!storedPlayerId) {
    localStorage.setItem(STORAGE_KEY_DAILY_PLAYER_ID, playerId.value)
  }

  const date = ref<string>('')
  const puzzles = ref<DailyPuzzle[]>([])
  const secondsUntilReset = ref<number>(0)
  const loading = ref<boolean>(false)
  const error = ref<string | null>(null)

  const easyPuzzle = computed(() => puzzles.value.find(p => p.difficulty === 'easy'))
  const mediumPuzzle = computed(() => puzzles.value.find(p => p.difficulty === 'medium'))
  const hardPuzzle = computed(() => puzzles.value.find(p => p.difficulty === 'hard'))

  const allSolved = computed(() =>
    puzzles.value.length === 3 && puzzles.value.every(p => p.solved)
  )

  const solvedCount = computed(() =>
    puzzles.value.filter(p => p.solved).length
  )

  async function fetchDaily(): Promise<void> {
    loading.value = true
    error.value = null

    try {
      // Get timezone offset (minutes behind UTC, e.g., PST is 480)
      const timezoneOffsetMinutes = new Date().getTimezoneOffset()

      const response = await bounceBotClient.getDailyChallenge({
        playerId: playerId.value,
        timezoneOffsetMinutes,
      })

      date.value = response.date
      secondsUntilReset.value = response.secondsUntilReset
      puzzles.value = response.puzzles.map((p: DailyPuzzleInfo) => ({
        difficulty: p.difficulty,
        game: p.game!,
        solved: p.solved,
        optimalMoves: p.optimalMoves,
      }))
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load daily challenges'
    } finally {
      loading.value = false
    }
  }

  async function submitSolution(difficulty: string, moves: BotPos[]): Promise<{ correct: boolean; newCompletion: boolean }> {
    try {
      const response = await bounceBotClient.submitDailySolution({
        playerId: playerId.value,
        date: date.value,
        difficulty,
        moves,
      })

      // Update local state if correct
      if (response.correct) {
        const puzzle = puzzles.value.find(p => p.difficulty === difficulty)
        if (puzzle) {
          puzzle.solved = true
        }
      }

      return {
        correct: response.correct,
        newCompletion: response.newCompletion,
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to submit solution'
      return { correct: false, newCompletion: false }
    }
  }

  function getPuzzleByDifficulty(difficulty: string): DailyPuzzle | undefined {
    return puzzles.value.find(p => p.difficulty === difficulty)
  }

  return {
    playerId,
    date,
    puzzles,
    secondsUntilReset,
    loading,
    error,
    easyPuzzle,
    mediumPuzzle,
    hardPuzzle,
    allSolved,
    solvedCount,
    fetchDaily,
    submitSolution,
    getPuzzleByDifficulty,
  }
})
