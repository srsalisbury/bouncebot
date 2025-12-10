<script setup lang="ts">
defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

function handleBackdropClick(event: MouseEvent) {
  if (event.target === event.currentTarget) {
    emit('close')
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-backdrop" @click="handleBackdropClick">
      <div class="modal">
        <button class="close-btn" @click="emit('close')">×</button>
        <h2>How to Play</h2>

        <div class="section">
          <h3>Goal</h3>
          <p>The colored circles represent robots that can move around on the board. Your goal is to move the target robot to the target cell. The target cell is shown as a colored square with a number matching the target robot you need to move there.</p>
        </div>

        <div class="section">
          <h3>Robot Movement</h3>
          <p>Robots slide in a straight line until they hit a wall, the edge of the board, or another robot. They cannot stop in the middle of the board.</p>
        </div>

        <div class="section">
          <h3>Controls</h3>
          <ul>
            <li><kbd>1</kbd>-<kbd>4</kbd> — Select robot by number</li>
            <li><kbd>↑</kbd><kbd>↓</kbd><kbd>←</kbd><kbd>→</kbd> or <kbd>W</kbd><kbd>A</kbd><kbd>S</kbd><kbd>D</kbd> — Move selected robot</li>
            <li><kbd>Z</kbd> / <kbd>U</kbd> / <kbd>Esc</kbd> — Undo last move</li>
            <li><kbd>Shift+R</kbd> — Reset puzzle</li>
          </ul>
          <p>You can also click on a robot to select it.</p>
        </div>

        <div class="section">
          <h3>Tips</h3>
          <ul>
            <li>Use other robots as blockers to stop the target robot where you need it.</li>
            <li>The colored dots show where robots started and where they've been.</li>
          </ul>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 1.5rem 2rem;
  max-width: 450px;
  max-height: 80vh;
  overflow-y: auto;
  position: relative;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
}

.close-btn {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  background: none;
  border: none;
  color: #888;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

h2 {
  margin: 0 0 1rem 0;
  color: #42b883;
  font-size: 1.5rem;
}

.section {
  margin-bottom: 1.25rem;
}

.section:last-child {
  margin-bottom: 0;
}

h3 {
  margin: 0 0 0.5rem 0;
  color: #eee;
  font-size: 1rem;
}

p {
  margin: 0;
  color: #bbb;
  font-size: 0.9rem;
  line-height: 1.5;
}

ul {
  margin: 0;
  padding-left: 1.25rem;
  color: #bbb;
  font-size: 0.9rem;
  line-height: 1.7;
}

kbd {
  background: #333;
  color: #fff;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: inherit;
  font-size: 0.8rem;
}
</style>
