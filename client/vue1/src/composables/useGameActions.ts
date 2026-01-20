import { ref, watch, type Ref } from 'vue'
import { bounceBotClient } from '../services/connectClient'
import { useGameStore } from '../stores/gameStore'
import { useRoomStore } from '../stores/roomStore'
import { create } from '@bufbuild/protobuf'
import { BotPosSchema, PositionSchema } from '../gen/bouncebot_pb'
import { getServerRejectionReason, type ServerRejectionReason } from '../services/errorUtils'

const STORAGE_KEY_PREFIX = 'bouncebot_best_solution_'

type BestSolutionInfo = {
  index: number
  moveCount: number
}

function getBestSolutionStorageKey(roomId: string): string {
  return `${STORAGE_KEY_PREFIX}${roomId}`
}

function loadBestSolutionInfo(roomId: string): BestSolutionInfo | null {
  try {
    const stored = localStorage.getItem(getBestSolutionStorageKey(roomId))
    if (!stored) return null
    return JSON.parse(stored) as BestSolutionInfo
  } catch {
    return null
  }
}

function saveBestSolutionInfo(roomId: string, info: BestSolutionInfo): void {
  try {
    localStorage.setItem(getBestSolutionStorageKey(roomId), JSON.stringify(info))
  } catch {
    // Ignore storage errors
  }
}

function clearBestSolutionStorage(roomId: string): void {
  try {
    localStorage.removeItem(getBestSolutionStorageKey(roomId))
  } catch {
    // Ignore storage errors
  }
}

export interface GameActionsOptions {
  roomId: Ref<string>
  onRoomUpdated?: () => void
  onServerRejection?: (reason: ServerRejectionReason) => void
}

export function useGameActions(options: GameActionsOptions) {
  const { roomId, onRoomUpdated, onServerRejection } = options

  const gameStore = useGameStore()
  const roomStore = useRoomStore()

  // Track the best submitted solution by its index and move count
  const bestSubmittedInfo = ref<BestSolutionInfo | null>(null)

  // Load best solution from localStorage on init
  function loadBestSolutionFromStorage() {
    if (!roomId.value) return
    bestSubmittedInfo.value = loadBestSolutionInfo(roomId.value)
  }

  // Load on init
  loadBestSolutionFromStorage()

  // Reload when roomId changes
  watch(roomId, () => {
    loadBestSolutionFromStorage()
  })

  function handleError(e: unknown, action: string) {
    console.error(`Failed to ${action}:`, e)
    const reason = getServerRejectionReason(e)
    if (reason) {
      onServerRejection?.(reason)
    }
  }

  async function submitSolution() {
    if (!roomStore.currentPlayerId) return
    const moveCount = gameStore.moveCount
    const currentMoves = gameStore.moves
    const currentIndex = gameStore.activeSolutionIndex

    // Only submit if this is better than our previous best (or first submission)
    if (bestSubmittedInfo.value !== null && moveCount >= bestSubmittedInfo.value.moveCount) return

    const moves = currentMoves.map(move =>
      create(BotPosSchema, {
        id: move.robotId,
        pos: create(PositionSchema, { x: move.toX, y: move.toY }),
      })
    )

    try {
      await bounceBotClient.submitSolution({
        roomId: roomId.value,
        playerId: roomStore.currentPlayerId,
        moves,
      })
      const info: BestSolutionInfo = { index: currentIndex, moveCount }
      bestSubmittedInfo.value = info
      saveBestSolutionInfo(roomId.value, info)
      onRoomUpdated?.()
    } catch (e) {
      handleError(e, 'submit solution')
    }
  }

  async function markFinishedSolving() {
    if (!roomStore.currentPlayerId) return

    try {
      await bounceBotClient.markFinishedSolving({
        roomId: roomId.value,
        playerId: roomStore.currentPlayerId,
      })
      onRoomUpdated?.()
    } catch (e) {
      handleError(e, 'mark finished')
    }
  }

  async function markReadyForNext() {
    if (!roomStore.currentPlayerId) return

    try {
      await bounceBotClient.markReadyForNext({
        roomId: roomId.value,
        playerId: roomStore.currentPlayerId,
      })
      onRoomUpdated?.()
    } catch (e) {
      handleError(e, 'mark ready')
    }
  }

  function resetForNewGame() {
    bestSubmittedInfo.value = null
    if (roomId.value) {
      clearBestSolutionStorage(roomId.value)
    }
  }

  // Check if the given solution index is the best submitted solution
  function isBestSubmittedSolution(solutionIndex: number): boolean {
    if (bestSubmittedInfo.value === null) return false
    return solutionIndex === bestSubmittedInfo.value.index
  }

  // Get the current best submitted solution index (or null if none submitted)
  function getBestSubmittedIndex(): number | null {
    return bestSubmittedInfo.value?.index ?? null
  }

  // Notify that a solution at the given index was deleted.
  // This adjusts the best submitted index if needed.
  function notifySolutionDeleted(deletedIndex: number): void {
    if (bestSubmittedInfo.value === null) return

    if (deletedIndex === bestSubmittedInfo.value.index) {
      // The best submitted solution was deleted
      bestSubmittedInfo.value = null
      if (roomId.value) {
        clearBestSolutionStorage(roomId.value)
      }
    } else if (deletedIndex < bestSubmittedInfo.value.index) {
      // A solution before the best submitted was deleted, shift index down
      bestSubmittedInfo.value = {
        ...bestSubmittedInfo.value,
        index: bestSubmittedInfo.value.index - 1,
      }
      if (roomId.value) {
        saveBestSolutionInfo(roomId.value, bestSubmittedInfo.value)
      }
    }
  }

  return {
    submitSolution,
    markFinishedSolving,
    markReadyForNext,
    resetForNewGame,
    isBestSubmittedSolution,
    getBestSubmittedIndex,
    notifySolutionDeleted,
  }
}
