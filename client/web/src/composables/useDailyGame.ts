import { ref, computed } from 'vue'
import { useGameStore } from '../stores/gameStore'
import { useDailyStore } from '../stores/dailyStore'
import { create } from '@bufbuild/protobuf'
import { BotPosSchema, PositionSchema } from '../gen/bouncebot_pb'

export function useDailyGame(difficulty: string) {
  const gameStore = useGameStore()
  const dailyStore = useDailyStore()

  const isSubmitting = ref(false)
  const submitResult = ref<{ correct: boolean; newCompletion: boolean } | null>(null)

  const puzzle = computed(() => dailyStore.getPuzzleByDifficulty(difficulty))
  const isSolved = computed(() => puzzle.value?.solved ?? false)

  async function submitSolution() {
    if (isSubmitting.value) return
    if (!gameStore.isSolved) return

    isSubmitting.value = true

    const moves = gameStore.moves.map(move =>
      create(BotPosSchema, {
        id: move.robotId,
        pos: create(PositionSchema, { x: move.toX, y: move.toY }),
      })
    )

    const result = await dailyStore.submitSolution(difficulty, moves)
    submitResult.value = result
    isSubmitting.value = false

    return result
  }

  function resetSubmitResult() {
    submitResult.value = null
  }

  return {
    puzzle,
    isSolved,
    isSubmitting,
    submitResult,
    submitSolution,
    resetSubmitResult,
  }
}
