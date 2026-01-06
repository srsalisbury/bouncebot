import { ref, type Ref } from 'vue'
import { bounceBotClient } from '../services/connectClient'
import { useGameStore } from '../stores/gameStore'
import { useRoomStore } from '../stores/roomStore'
import { create } from '@bufbuild/protobuf'
import { BotPosSchema, PositionSchema } from '../gen/bouncebot_pb'
import { getServerRejectionReason, type ServerRejectionReason } from '../services/errorUtils'

export interface GameActionsOptions {
  roomId: Ref<string>
  onRoomUpdated?: () => void
  onServerRejection?: (reason: ServerRejectionReason) => void
}

export function useGameActions(options: GameActionsOptions) {
  const { roomId, onRoomUpdated, onServerRejection } = options

  const gameStore = useGameStore()
  const roomStore = useRoomStore()

  const bestSubmittedMoveCount = ref<number | null>(null)

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
      handleError(e, 'submit solution')
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
      handleError(e, 'retract solution')
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
