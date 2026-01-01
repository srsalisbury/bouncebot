<script setup lang="ts">
import { ref, computed } from 'vue'
import { useGameStore } from '../stores/gameStore'
import { DIRECTION_ARROWS, getRobotColor } from '../constants'
import { useSwipe } from '../composables/useSwipe'

const store = useGameStore()
const isExpanded = ref(false)
const drawerRef = ref<HTMLElement | null>(null)

// Swipe down to collapse
useSwipe({
  target: drawerRef,
  onSwipe: (direction) => {
    if (direction === 'down' && isExpanded.value) {
      isExpanded.value = false
    } else if (direction === 'up' && !isExpanded.value) {
      isExpanded.value = true
    }
  },
  minDistance: 30,
})

function toggleExpanded() {
  isExpanded.value = !isExpanded.value
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
          @click="store.switchSolution(index)"
        >
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
  border-top-left-radius: 16px;
  border-top-right-radius: 16px;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.3);
  z-index: 100;
  transition: max-height 0.3s ease;
  max-height: 60px;
  overflow: hidden;
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
  color: #42b883;
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
  overflow-x: auto;
  padding: 4px; /* Space for box-shadow outline */
  padding-bottom: 0.5rem;
}

.solution-column {
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

.solution-column:hover {
  background: #333;
}

.solution-column.active {
  background: #2a2a2a;
  box-shadow: 0 0 0 2px #42b883;
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
  color: #42b883;
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
  background: #42b883;
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
}

.move-arrow {
  font-size: 18px;
  color: #ddd;
  width: 18px;
  text-align: center;
}
</style>
