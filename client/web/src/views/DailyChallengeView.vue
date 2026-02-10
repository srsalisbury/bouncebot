<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDailyStore } from '../stores/dailyStore'

const router = useRouter()
const dailyStore = useDailyStore()

const countdown = ref('')
let countdownInterval: ReturnType<typeof setInterval> | null = null

// Difficulty display info
const difficulties = [
  { key: 'easy', label: 'Easy', description: '4-6 moves', color: '#43a047' },
  { key: 'medium', label: 'Medium', description: '7-11 moves', color: '#fb8c00' },
  { key: 'hard', label: 'Hard', description: '12+ moves', color: '#e53935' },
]

const puzzleMap = computed(() => {
  const map: Record<string, typeof dailyStore.puzzles[0]> = {}
  for (const puzzle of dailyStore.puzzles) {
    map[puzzle.difficulty] = puzzle
  }
  return map
})

function formatCountdown(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60
  return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

function updateCountdown() {
  if (dailyStore.secondsUntilReset > 0) {
    dailyStore.secondsUntilReset--
    countdown.value = formatCountdown(dailyStore.secondsUntilReset)
    if (dailyStore.secondsUntilReset <= 0) {
      // Refresh when day resets
      dailyStore.fetchDaily()
    }
  }
}

function playPuzzle(difficulty: string) {
  router.push(`/daily/${difficulty}`)
}

function goHome() {
  router.push('/')
}

onMounted(async () => {
  await dailyStore.fetchDaily()
  countdown.value = formatCountdown(dailyStore.secondsUntilReset)
  countdownInterval = setInterval(updateCountdown, 1000)
})

onUnmounted(() => {
  if (countdownInterval) {
    clearInterval(countdownInterval)
  }
})
</script>

<template>
  <div class="daily-view">
    <div class="header">
      <button class="back-btn" @click="goHome">&larr;</button>
      <h1>Daily Challenge</h1>
    </div>

    <div v-if="dailyStore.loading" class="loading">Loading puzzles...</div>
    <div v-else-if="dailyStore.error" class="error">{{ dailyStore.error }}</div>
    <template v-else>
      <p class="date-info">{{ dailyStore.date }}</p>

      <div class="puzzles-grid">
        <button
          v-for="diff in difficulties"
          :key="diff.key"
          class="puzzle-card"
          :class="{ solved: puzzleMap[diff.key]?.solved }"
          :style="{ '--accent-color': diff.color }"
          @click="playPuzzle(diff.key)"
        >
          <div class="card-header">
            <span class="difficulty-label">{{ diff.label }}</span>
            <span v-if="puzzleMap[diff.key]?.solved" class="solved-badge">&#10003;</span>
          </div>
          <p class="difficulty-desc">{{ diff.description }}</p>
          <div v-if="puzzleMap[diff.key]?.solved" class="optimal-moves">
            Optimal: {{ puzzleMap[diff.key]?.optimalMoves }} moves
          </div>
        </button>
      </div>

      <div class="progress-section">
        <div class="progress-bar">
          <div
            class="progress-fill"
            :style="{ width: `${(dailyStore.solvedCount / 3) * 100}%` }"
          />
        </div>
        <p class="progress-text">{{ dailyStore.solvedCount }} / 3 completed</p>
      </div>

      <div class="reset-info">
        <p>New puzzles in</p>
        <p class="countdown">{{ countdown }}</p>
      </div>
    </template>
  </div>
</template>

<style scoped>
.daily-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
  min-height: 100vh;
  position: relative;
  overflow: hidden;
}

.daily-view::before {
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
  .daily-view::before {
    background-image: url('/pattern_light.svg');
    opacity: 0.4;
  }
}

.header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.5rem;
}

.back-btn {
  background: #333;
  border: none;
  color: #fff;
  font-size: 1.25rem;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  cursor: pointer;
}

.back-btn:hover {
  background: #444;
}

h1 {
  color: #fff;
  margin: 0;
  font-size: 1.75rem;
}

.date-info {
  color: #888;
  margin: 0 0 1.5rem;
}

.loading {
  color: #888;
  font-size: 1.1rem;
}

.error {
  color: #e53935;
  background: rgba(229, 57, 53, 0.1);
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid #e53935;
}

.puzzles-grid {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  max-width: 360px;
}

.puzzle-card {
  background: #1a1a1a;
  border: 2px solid #333;
  border-radius: 12px;
  padding: 1.25rem;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.puzzle-card:hover {
  border-color: var(--accent-color);
  transform: translateY(-2px);
}

.puzzle-card.solved {
  border-color: var(--accent-color);
  background: rgba(var(--accent-color), 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.difficulty-label {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--accent-color);
}

.solved-badge {
  background: var(--accent-color);
  color: #fff;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9rem;
}

.difficulty-desc {
  color: #888;
  margin: 0;
  font-size: 0.9rem;
}

.optimal-moves {
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid #333;
  color: #aaa;
  font-size: 0.85rem;
}

.progress-section {
  margin-top: 2rem;
  width: 100%;
  max-width: 360px;
}

.progress-bar {
  height: 8px;
  background: #333;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: #43a047;
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-text {
  text-align: center;
  color: #888;
  font-size: 0.9rem;
  margin-top: 0.5rem;
}

.reset-info {
  margin-top: 2rem;
  text-align: center;
}

.reset-info p {
  margin: 0;
  color: #888;
  font-size: 0.9rem;
}

.countdown {
  font-size: 1.5rem;
  font-weight: 600;
  color: #fff;
  font-family: monospace;
  margin-top: 0.25rem;
}

@media (prefers-color-scheme: light) {
  h1 {
    color: #000;
  }

  .puzzle-card {
    background: #fff;
    border-color: #ddd;
  }

  .puzzle-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .optimal-moves {
    border-color: #eee;
  }

  .countdown {
    color: #000;
  }
}
</style>
