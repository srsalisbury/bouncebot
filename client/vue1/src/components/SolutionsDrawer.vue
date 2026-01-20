<script setup lang="ts">
import { ref, computed } from 'vue'
import { useGameStore } from '../stores/gameStore'
import { DIRECTION_ARROWS, getRobotColor, MOBILE_ASPECT_RATIO, MOBILE_WIDTH_BREAKPOINT } from '../constants'
import { useSwipe } from '../composables/useSwipe'

const props = defineProps<{
  getBestSubmittedIndex?: () => number | null
  onSolutionDeleted?: (index: number) => void
}>()

const store = useGameStore()
const isExpanded = ref(false)
const drawerRef = ref<HTMLElement | null>(null)

// Check if we're on mobile/vertical layout
function isMobile(): boolean {
  const aspectRatio = window.innerWidth / window.innerHeight
  return aspectRatio < MOBILE_ASPECT_RATIO || window.innerWidth <= MOBILE_WIDTH_BREAKPOINT
}

// Switch solution and auto-collapse on mobile
function handleSolutionClick(index: number) {
  store.switchSolution(index)
  if (isMobile()) {
    isExpanded.value = false
  }
}

// Replay solution and auto-collapse on mobile
function handleReplayClick() {
  store.replaySolution()
  if (isMobile()) {
    isExpanded.value = false
  }
}

// Swipe to expand/collapse and switch solutions
useSwipe({
  target: drawerRef,
  onSwipe: (direction) => {
    if (direction === 'down' && isExpanded.value) {
      isExpanded.value = false
    } else if (direction === 'up' && !isExpanded.value) {
      isExpanded.value = true
    } else if (!isExpanded.value) {
      // Swipe left/right on collapsed drawer to switch solutions
      if (direction === 'left' && store.activeSolutionIndex < store.solutions.length - 1) {
        store.switchSolution(store.activeSolutionIndex + 1)
      } else if (direction === 'right' && store.activeSolutionIndex > 0) {
        store.switchSolution(store.activeSolutionIndex - 1)
      }
    }
  },
  minDistance: 30,
})

function toggleExpanded() {
  isExpanded.value = !isExpanded.value
}

// Start a new solution, handling auto-delete if at max capacity
function handleNewSolution() {
  const bestIndex = props.getBestSubmittedIndex?.() ?? null
  const result = store.startNewSolution(bestIndex)
  if (result.deletedIndex !== null) {
    props.onSolutionDeleted?.(result.deletedIndex)
  }
}

// Delete a solution and notify parent
function handleDeleteSolution(index: number) {
  store.deleteSolution(index)
  props.onSolutionDeleted?.(index)
}

const activeSolution = computed(() => store.solutions[store.activeSolutionIndex])
const moveCount = computed(() => activeSolution.value?.moves.length ?? 0)
const isSolved = computed(() => activeSolution.value?.isSolved ?? false)
const solutionCount = computed(() => store.solutions.length)
</script>

<template>
  <div
    ref="drawerRef"
    class="solutions-drawer"
    :class="{ expanded: isExpanded }"
  >
    <!-- Collapsed header bar -->
    <div class="drawer-header" @click="toggleExpanded">
      <div class="drawer-handle" />
      <div class="header-content">
        <span class="solution-count">
          {{ moveCount }} {{ moveCount === 1 ? 'move' : 'moves' }}
          <span v-if="isSolved" class="solved-indicator">✓</span>
        </span>
        <span class="solution-label">Solution {{ store.activeSolutionIndex + 1 }}/{{ solutionCount }}</span>
      </div>
    </div>

    <!-- Expanded content -->
    <div v-if="isExpanded" class="drawer-content">
      <div class="solutions-columns">
        <div
          v-for="(solution, index) in store.solutions"
          :key="index"
          class="solution-column"
          :class="{ active: index === store.activeSolutionIndex }"
          @click="handleSolutionClick(index)"
        >
          <!-- Replay button on active solution -->
          <button
            v-if="index === store.activeSolutionIndex && solution.moves.length > 0"
            class="replay-btn"
            @click.stop="handleReplayClick()"
          >
            <span class="play-icon">▶</span>
          </button>
          <!-- Delete button on active solution (only if more than 1 solution) -->
          <button
            v-if="index === store.activeSolutionIndex && store.solutions.length > 1"
            class="delete-btn"
            @click.stop="handleDeleteSolution(index)"
          >
            ×
          </button>
          <div class="solution-header">
            <span class="solution-moves">{{ solution.moves.length }}</span>
            <span class="solved-check" :class="{ visible: solution.isSolved }">✓</span>
          </div>
          <div class="move-list">
            <div
              v-for="(move, i) in solution.moves"
              :key="i"
              class="move-item"
              :class="{ animating: index === store.activeSolutionIndex && store.animatingMoveIndex === i }"
            >
              <span class="move-robot" :style="{ backgroundColor: getRobotColor(move.robotId) }">
                {{ move.robotId + 1 }}
              </span>
              <span class="move-arrow">{{ DIRECTION_ARROWS[move.direction] }}</span>
            </div>
          </div>
        </div>
        <!-- Add new solution button (always visible, auto-deletes worst if at max) -->
        <button
          class="add-solution-btn"
          @click="handleNewSolution()"
        >
          +
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.solutions-drawer {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: #1a1a1a;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.3);
  z-index: 100;
  transition: max-height 0.3s ease;
  max-height: calc(60px + env(safe-area-inset-bottom, 0px));
  overflow: hidden;
  padding-bottom: env(safe-area-inset-bottom, 0px);
}

.solutions-drawer.expanded {
  max-height: 50vh;
  overflow-y: auto;
}

.drawer-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  cursor: pointer;
  user-select: none;
  min-height: 44px;
}

.drawer-handle {
  position: absolute;
  top: 8px;
  left: 50%;
  transform: translateX(-50%);
  width: 40px;
  height: 4px;
  background: #444;
  border-radius: 2px;
}

.header-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.solution-count {
  font-size: 1.1rem;
  font-weight: 600;
  color: #fff;
}

.solved-indicator {
  color: #43a047;
  margin-left: 0.3rem;
}

.solution-label {
  font-size: 0.8rem;
  color: #888;
}

.drawer-content {
  padding: 0 1rem 1rem;
}

.solutions-columns {
  display: flex;
  flex-direction: row;
  gap: 0.5rem;
  justify-content: center;
  overflow-x: auto;
  padding: 12px; /* Space for delete/replay buttons that overflow */
  padding-bottom: 0.5rem;
}

.solution-column {
  position: relative;
  min-width: 54px;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.4rem;
  border-radius: 6px;
  background: #2a2a2a;
  cursor: pointer;
  flex-shrink: 0;
}

.delete-btn {
  /* Reset iOS button styling */
  -webkit-appearance: none;
  appearance: none;
  position: absolute;
  top: -8px;
  right: -8px;
  width: 22px;
  min-width: 22px;
  max-width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #e53935;
  color: white;
  border: 2px solid #1a1a1a;
  font-size: 16px;
  font-weight: bold;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  box-sizing: border-box;
  padding: 0;
}

.delete-btn:hover {
  background: #c62828;
}

.replay-btn {
  /* Reset iOS button styling */
  -webkit-appearance: none;
  appearance: none;
  position: absolute;
  top: -8px;
  left: -8px;
  width: 22px;
  min-width: 22px;
  max-width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #43a047;
  color: #333;
  border: 2px solid #1a1a1a;
  font-size: 10px;
  font-weight: bold;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  box-sizing: border-box;
  padding: 0;
}

.replay-btn:hover {
  background: #3aa876;
}

.play-icon {
  margin-left: 2px;
}

.add-solution-btn {
  /* Reset iOS button styling */
  -webkit-appearance: none;
  appearance: none;
  width: 44px;
  min-width: 44px;
  max-width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #333;
  color: #43a047;
  border: 2px solid #43a047;
  font-size: 24px;
  font-weight: bold;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 44px;
  align-self: flex-start;
  box-sizing: border-box;
  padding: 0;
}

.add-solution-btn:hover {
  background: #3a3a3a;
}

.solution-column:hover {
  background: #333;
}

.solution-column.active {
  background: #2a2a2a;
  box-shadow: 0 0 0 2px #43a047;
}

.solution-header {
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  font-weight: 600;
  font-size: 1.2rem;
  padding-bottom: 0.4rem;
  margin-bottom: 0.25rem;
  border-bottom: 1px solid #444;
}

.solution-moves {
  color: #fff;
}

.solved-check {
  position: absolute;
  right: 0;
  color: #43a047;
  opacity: 0;
}

.solved-check.visible {
  opacity: 1;
}

.move-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  max-height: 200px;
  overflow-y: auto;
}

.move-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 2px 4px;
  border-radius: 4px;
}

.move-item.animating {
  background: #43a047;
}

.move-robot {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 12px;
  color: white;
  border: 0.5px solid black;
  text-shadow: -0.3px -0.3px 0 black, 0.3px -0.3px 0 black, -0.3px 0.3px 0 black, 0.3px 0.3px 0 black;
}

.move-arrow {
  font-size: 18px;
  color: #ddd;
  width: 18px;
  text-align: center;
}
</style>
