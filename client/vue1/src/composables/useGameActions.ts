import { ref, type Ref } from 'vue'
import { bounceBotClient } from '../services/connectClient'
import { useGameStore } from '../stores/gameStore'
import { useRoomStore } from '../stores/roomStore'
import { create } from '@bufbuild/protobuf'
import { BotPosSchema, PositionSchema } from '../gen/bouncebot_pb'

export interface GameActionsOptions {
  roomId: Ref<string>
  onRoomUpdated?: () => void
}

export function useGameActions(options: GameActionsOptions) {
  const { roomId, onRoomUpdated } = options

  const gameStore = useGameStore()
  const roomStore = useRoomStore()

  const bestSubmittedMoveCount = ref<number | null>(null)

  async function submitSolution() {
    if (!roomStore.currentPlayerId) return
    const moveCount = gameStore.moveCount

    // Only submit if this is better than our previous best (or first submission)
    if (bestSubmittedMoveCount.value !== null && moveCount >= bestSubmittedMoveCount.value) return

    const moves = gameStore.moves.map(move =>
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
      bestSubmittedMoveCount.value = moveCount
      onRoomUpdated?.()
    } catch (e) {
      console.error('Failed to submit solution:', e)
    }
  }

  async function retractSolution() {
    if (!roomStore.currentPlayerId) return

    try {
      await bounceBotClient.retractSolution({
        roomId: roomId.value,
        playerId: roomStore.currentPlayerId,
      })
      onRoomUpdated?.()
    } catch (e) {
      console.error('Failed to retract solution:', e)
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
      console.error('Failed to mark finished:', e)
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
      console.error('Failed to mark ready:', e)
    }
  }

  function resetForNewGame() {
    bestSubmittedMoveCount.value = null
  }

  function restoreBestMoveCount(moveCount: number) {
    bestSubmittedMoveCount.value = moveCount
  }

  function clearBestMoveCount() {
    bestSubmittedMoveCount.value = null
  }

  return {
    bestSubmittedMoveCount,
    submitSolution,
    retractSolution,
    markFinishedSolving,
    markReadyForNext,
    resetForNewGame,
    restoreBestMoveCount,
    clearBestMoveCount,
  }
}
