<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useGameStore } from '../stores/gameStore'
import { useDailyStore } from '../stores/dailyStore'
import { useDailyGame } from '../composables/useDailyGame'
import GameBoard from '../components/GameBoard.vue'

const props = defineProps<{
  difficulty: string
}>()

const router = useRouter()
const gameStore = useGameStore()
const dailyStore = useDailyStore()
const {
  puzzle,
  isSolved,
  isSubmitting,
  submitResult,
  submitSolution,
} = useDailyGame(props.difficulty)

const showSuccessModal = ref(false)
const hasSubmitted = ref(false)

const difficultyLabel = computed(() => {
  switch (props.difficulty) {
    case 'easy': return 'Easy'
    case 'medium': return 'Medium'
    case 'hard': return 'Hard'
    default: return props.difficulty
  }
})

const difficultyColor = computed(() => {
  switch (props.difficulty) {
    case 'easy': return '#43a047'
    case 'medium': return '#fb8c00'
    case 'hard': return '#e53935'
    default: return '#888'
  }
})

function goBack() {
  router.push('/daily')
}

async function handleSubmit() {
  if (hasSubmitted.value) return
  hasSubmitted.value = true

  const result = await submitSolution()
  if (result?.correct) {
    showSuccessModal.value = true
  }
}

function closeSuccessModal() {
  showSuccessModal.value = false
  router.push('/daily')
}

// Auto-submit when solved
watch(
  () => gameStore.isSolved,
  (solved) => {
    if (solved && !hasSubmitted.value && !isSolved.value) {
      handleSubmit()
    }
  }
)

onMounted(async () => {
  // Ensure we have the puzzle data
  if (dailyStore.puzzles.length === 0) {
    await dailyStore.fetchDaily()
  }

  // Load the puzzle into the game store
  const p = puzzle.value
  if (p) {
    // Create a unique "room ID" for daily puzzles to enable separate storage
    const dailyRoomId = `daily-${dailyStore.date}-${props.difficulty}`
    gameStore.applyGame(p.game, dailyRoomId, 1)

    // If already solved, mark as submitted
    if (p.solved) {
      hasSubmitted.value = true
    }
  }
})
</script>

<template>
  <div class="daily-game-view">
    <div v-if="!puzzle" class="loading">Loading puzzle...</div>

    <template v-else>
      <GameBoard
        :game-ended="false"
        :player-solutions="[]"
        :current-player-solution="null"
        :top-three-solutions="[]"
        :solver-solutions="[]"
        :get-player-name="() => 'You'"
        :get-player-color="() => '#43a047'"
        :is-solver-solution="() => false"
        :game-started-at="undefined"
        :room-id="''"
        :game-number="1"
        :input-blocked="showSuccessModal"
        :single-player="true"
        :get-best-submitted-index="() => null"
        :on-solution-deleted="() => {}"
      >
        <template #header>
          <div class="game-header">
            <button class="back-btn" @click="goBack">&larr;</button>
            <span class="difficulty-badge" :style="{ backgroundColor: difficultyColor }">
              {{ difficultyLabel }}
            </span>
            <span class="date-label">{{ dailyStore.date }}</span>
            <span v-if="gameStore.isSolved || isSolved" class="solved-indicator">&#10003;</span>
            <span v-if="isSolved && puzzle.optimalMoves" class="optimal-label">
              Optimal: {{ puzzle.optimalMoves }}
            </span>
            <button
              v-if="gameStore.isSolved && !hasSubmitted"
              class="btn primary submit-btn"
              :disabled="isSubmitting"
              @click="handleSubmit"
            >
              {{ isSubmitting ? 'Submitting...' : 'Submit' }}
            </button>
            <button
              v-else-if="isSolved"
              class="btn done-btn"
              @click="goBack"
            >
              Done
            </button>
          </div>
        </template>
      </GameBoard>
    </template>

    <!-- Success Modal -->
    <div v-if="showSuccessModal" class="modal-overlay" @click.self="closeSuccessModal">
      <div class="modal">
        <div class="modal-icon">&#10003;</div>
        <h2>Puzzle Solved!</h2>
        <p v-if="submitResult?.newCompletion">
          You completed the {{ difficultyLabel }} daily challenge!
        </p>
        <p v-else>
          You already solved this puzzle before.
        </p>
        <p class="move-count">
          Your solution: {{ gameStore.moveCount }} moves
          <template v-if="puzzle?.optimalMoves">
            <br />Optimal: {{ puzzle.optimalMoves }} moves
          </template>
        </p>
        <button class="btn primary" @click="closeSuccessModal">
          Continue
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.daily-game-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-height: 100vh;
  position: relative;
  overflow: hidden;
}

.daily-game-view::before {
  content: '';
  position: fixed;
  top: 50%;
  left: 50%;
  width: 185vmax;
  height: 185vmax;
  background-image: url('/pattern_dark.svg');
  background-repeat: no-repeat;
  background-size: auto 100%;
  background-position: center;
  transform: translate(-50%, -50%) rotate(22.5deg);
  z-index: -1;
  opacity: 0.7;
}

@media (prefers-color-scheme: light) {
  .daily-game-view::before {
    background-image: url('/pattern_light.svg');
    opacity: 0.4;
  }
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  color: #888;
  font-size: 1.1rem;
}

.game-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 1rem;
  background: #1a1a1a;
  border-radius: 8px;
}

.back-btn {
  background: #333;
  border: none;
  color: #fff;
  font-size: 1rem;
  padding: 0.4rem 0.6rem;
  border-radius: 4px;
  cursor: pointer;
}

.back-btn:hover {
  background: #444;
}

.difficulty-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  color: #fff;
  font-weight: 600;
  font-size: 0.85rem;
}

.date-label {
  color: #888;
  font-size: 0.85rem;
}

.solved-indicator {
  color: #43a047;
  font-size: 1.2rem;
  font-weight: bold;
}

.optimal-label {
  color: #888;
  font-size: 0.85rem;
}

.submit-btn {
  margin-left: auto;
  padding: 0.4rem 1rem;
  font-size: 0.9rem;
}

.done-btn {
  margin-left: auto;
  padding: 0.4rem 1rem;
  font-size: 0.9rem;
  background: #333;
  color: #fff;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.done-btn:hover {
  background: #444;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  background: #333;
  color: #fff;
}

.btn:hover {
  background: #444;
}

.btn.primary {
  background: #43a047;
  color: #fff;
}

.btn.primary:hover:not(:disabled) {
  background: #388e3c;
}

.btn.primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: #1a1a1a;
  border-radius: 16px;
  padding: 2rem;
  max-width: 360px;
  text-align: center;
  margin: 1rem;
}

.modal-icon {
  font-size: 3rem;
  color: #43a047;
  margin-bottom: 1rem;
}

.modal h2 {
  margin: 0 0 1rem;
  color: #fff;
}

.modal p {
  margin: 0 0 1rem;
  color: #aaa;
  line-height: 1.5;
}

.move-count {
  font-size: 0.9rem;
  color: #888;
}

@media (prefers-color-scheme: light) {
  .modal {
    background: #fff;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
  }

  .modal h2 {
    color: #000;
  }
}
</style>
