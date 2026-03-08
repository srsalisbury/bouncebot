<script setup lang="ts">
import HelpContent from './HelpContent.vue'

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
        <button class="close-btn" @click="emit('close')">&#x00d7;</button>
        <h2>How to Play</h2>
        <HelpContent />
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
  align-items: flex-start;
  justify-content: center;
  z-index: 100;
}

.modal {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 1.5rem 2rem;
  max-width: 500px;
  width: 90%;
  margin-top: 5vh;
  max-height: calc(100vh - 5vh - 5vh);
  box-sizing: border-box;
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
  color: #43a047;
  font-size: 1.5rem;
}

@media (prefers-color-scheme: light) {
  .modal {
    background: #fff;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
  }

  .close-btn {
    color: #666;
  }

  .close-btn:hover {
    color: #333;
    background: rgba(0, 0, 0, 0.1);
  }
}
</style>
