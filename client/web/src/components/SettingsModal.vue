<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { RoomSettings, Player } from '../gen/bouncebot_pb'
import { getPlayerColor } from '../constants'

const props = defineProps<{
  show: boolean
  settings: RoomSettings | undefined
  players?: Player[]
  showBootPlayer?: boolean
}>()

const emit = defineEmits<{
  close: []
  update: [settings: { showSolverMoveCount: boolean; showSolverSolutions: boolean }]
  bootPlayer: [playerId: string]
}>()

// Local state for the toggles
const showSolverMoveCount = ref(false)
const showSolverSolutions = ref(false)

// Boot player state
const showPlayerList = ref(false)
const showBootConfirm = ref(false)
const playerToBootId = ref<string | null>(null)

const playerToBootName = computed(() => {
  if (!playerToBootId.value || !props.players) return ''
  const player = props.players.find(p => p.id === playerToBootId.value)
  return player?.name ?? 'this player'
})

// Map player IDs to their color index
const playerColorMap = computed(() => {
  const map = new Map<string, number>()
  props.players?.forEach((player, index) => {
    map.set(player.id, index)
  })
  return map
})

function getPlayerColorById(playerId: string): string {
  const index = playerColorMap.value.get(playerId) ?? 0
  return getPlayerColor(index)
}

function openPlayerList() {
  showPlayerList.value = true
}

function closePlayerList() {
  showPlayerList.value = false
}

function selectPlayerToBoot(playerId: string) {
  playerToBootId.value = playerId
  showPlayerList.value = false
  showBootConfirm.value = true
}

function confirmBoot() {
  if (playerToBootId.value) {
    emit('bootPlayer', playerToBootId.value)
  }
  showBootConfirm.value = false
  playerToBootId.value = null
  emit('close')
}

function cancelBoot() {
  showBootConfirm.value = false
  playerToBootId.value = null
}

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
            <span class="toggle-text">Show BBot's move count during game</span>
          </label>
          <p class="setting-description">Display BBot's optimal move count in the header while playing.</p>
        </div>

        <div class="setting-item">
          <label class="toggle-label">
            <input
              type="checkbox"
              :checked="showSolverSolutions"
              @change="updateSetting('showSolverSolutions', ($event.target as HTMLInputElement).checked)"
            />
            <span class="toggle-text">Show BBot's solution after game</span>
          </label>
          <p class="setting-description">Include BBot's solution in the post-game review.</p>
        </div>

        <!-- Boot Player button -->
        <div v-if="showBootPlayer && players && players.length > 0" class="boot-section">
          <button class="boot-player-btn" @click="openPlayerList">
            Boot Player
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Player selection modal -->
  <Teleport to="body">
    <div v-if="showPlayerList" class="modal-backdrop player-list-backdrop" @click.self="closePlayerList">
      <div class="modal player-list-modal">
        <button class="close-btn" @click="closePlayerList">×</button>
        <h2>Select Player to Remove</h2>
        <div class="player-list">
          <button
            v-for="player in players"
            :key="player.id"
            class="player-item"
            @click="selectPlayerToBoot(player.id)"
          >
            <span class="player-dot" :style="{ backgroundColor: getPlayerColorById(player.id) }" />
            <span class="player-name">{{ player.name }}</span>
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Boot confirmation dialog -->
  <Teleport to="body">
    <div v-if="showBootConfirm" class="modal-backdrop confirm-backdrop" @click.self="cancelBoot">
      <div class="modal confirm-modal">
        <h3>Remove Player</h3>
        <p>Are you sure you want to remove <strong>{{ playerToBootName }}</strong> from the room?</p>
        <div class="dialog-actions">
          <button class="btn" @click="cancelBoot">Cancel</button>
          <button class="btn danger" @click="confirmBoot">Remove</button>
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

/* Boot player section */
.boot-section {
  margin-top: 1.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid #333;
}

.boot-player-btn {
  width: 100%;
  padding: 0.6rem 1rem;
  background: #333;
  border: 1px solid #444;
  border-radius: 6px;
  color: #eee;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.boot-player-btn:hover {
  background: #444;
  border-color: #555;
}

/* Player list modal */
.player-list-backdrop,
.confirm-backdrop {
  z-index: 200;
}

.player-list-modal {
  max-width: 320px;
}

.player-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.player-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.75rem;
  background: #242424;
  border: 1px solid transparent;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.player-item:hover {
  background: #2a2a2a;
  border-color: #e53935;
}

.player-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.player-name {
  color: #eee;
  font-size: 0.95rem;
  text-align: left;
}

/* Confirmation dialog */
.confirm-modal {
  max-width: 400px;
}

.confirm-modal h3 {
  margin: 0 0 1rem;
  color: #eee;
  font-size: 1.1rem;
}

.confirm-modal p {
  margin: 0 0 1.5rem;
  color: #aaa;
  line-height: 1.5;
}

.dialog-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  background: #333;
  color: #fff;
  font-size: 0.9rem;
}

.btn:hover {
  background: #444;
}

.btn.danger {
  background: #e53935;
}

.btn.danger:hover {
  background: #c62828;
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
