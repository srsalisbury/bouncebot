<script setup lang="ts">
import { ref, watch } from 'vue'
import type { RoomSettings } from '../gen/bouncebot_pb'

const props = defineProps<{
  show: boolean
  settings: RoomSettings | undefined
}>()

const emit = defineEmits<{
  close: []
  update: [settings: { showSolverMoveCount: boolean; showSolverSolutions: boolean }]
}>()

// Local state for the toggles
const showSolverMoveCount = ref(false)
const showSolverSolutions = ref(false)

// Sync local state when props change
watch(() => props.settings, (newSettings) => {
  if (newSettings) {
    showSolverMoveCount.value = newSettings.showSolverMoveCount
    showSolverSolutions.value = newSettings.showSolverSolutions
  }
}, { immediate: true })

function handleBackdropClick(event: MouseEvent) {
  if (event.target === event.currentTarget) {
    emit('close')
  }
}

function updateSetting(key: 'showSolverMoveCount' | 'showSolverSolutions', value: boolean) {
  if (key === 'showSolverMoveCount') {
    showSolverMoveCount.value = value
  } else {
    showSolverSolutions.value = value
  }
  emit('update', {
    showSolverMoveCount: showSolverMoveCount.value,
    showSolverSolutions: showSolverSolutions.value,
  })
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-backdrop" @click="handleBackdropClick">
      <div class="modal">
        <button class="close-btn" @click="emit('close')">×</button>
        <h2>Room Settings</h2>

        <div class="setting-item">
          <label class="toggle-label">
            <input
              type="checkbox"
              :checked="showSolverMoveCount"
              @change="updateSetting('showSolverMoveCount', ($event.target as HTMLInputElement).checked)"
            />
            <span class="toggle-text">Show solver move count during game</span>
          </label>
          <p class="setting-description">Display the bot's solution move count in the header while playing.</p>
        </div>

        <div class="setting-item">
          <label class="toggle-label">
            <input
              type="checkbox"
              :checked="showSolverSolutions"
              @change="updateSetting('showSolverSolutions', ($event.target as HTMLInputElement).checked)"
            />
            <span class="toggle-text">Show solver solutions after game</span>
          </label>
          <p class="setting-description">Include the bot's solution in the end-game solution viewer.</p>
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
  max-width: 400px;
  width: 90%;
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
  margin: 0 0 1.25rem 0;
  color: #43a047;
  font-size: 1.25rem;
}

.setting-item {
  margin-bottom: 1.25rem;
}

.setting-item:last-child {
  margin-bottom: 0;
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
}

.toggle-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #43a047;
}

.toggle-text {
  color: #eee;
  font-size: 0.95rem;
}

.setting-description {
  margin: 0.4rem 0 0 1.7rem;
  color: #888;
  font-size: 0.8rem;
  line-height: 1.4;
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

  .toggle-text {
    color: #333;
  }

  .setting-description {
    color: #666;
  }
}
</style>
