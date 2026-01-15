<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { bounceBotClient } from '../services/connectClient'
import { translateJoinRoomError } from '../services/errorMessages'
import { useRoomStore } from '../stores/roomStore'

const router = useRouter()
const roomStore = useRoomStore()

const storedName = roomStore.currentPlayerName
const playerName = ref(storedName && storedName !== 'Player' ? storedName : '')
const joinRoomId = ref('')
const isStartingSolo = ref(false)
const isCreating = ref(false)
const isJoining = ref(false)
const showMultiplayer = ref(false)
const error = ref<string | null>(null)

const lastRoom = computed(() => roomStore.lastRoomId)
const isLoading = computed(() => isStartingSolo.value || isCreating.value || isJoining.value)
const appVersion = __APP_VERSION__

function returnToGame() {
  if (lastRoom.value) {
    router.push(`/room/${lastRoom.value}`)
  }
}

async function startSoloGame() {
  isStartingSolo.value = true
  error.value = null

  try {
    // Create room with default name
    const room = await bounceBotClient.createRoom({
      playerName: 'Player',
      isSinglePlayer: true,
    })
    const player = room.players[0]
    if (player) {
      roomStore.setCurrentPlayer(player.id, player.name)
    }
    roomStore.setSinglePlayer(true)

    // Start game immediately
    await bounceBotClient.startGame({ roomId: room.id })

    router.push(`/room/${room.id}`)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to start game'
  } finally {
    isStartingSolo.value = false
  }
}

async function createRoom() {
  if (!playerName.value.trim()) {
    error.value = 'Please enter your name'
    return
  }

  isCreating.value = true
  error.value = null

  try {
    const room = await bounceBotClient.createRoom({
      playerName: playerName.value.trim(),
    })
    // Creator is the first (and only) player in the new room
    const player = room.players[0]
    if (player) {
      roomStore.setCurrentPlayer(player.id, player.name)
    }
    roomStore.setSinglePlayer(false)
    router.push(`/room/${room.id}`)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to create room'
  } finally {
    isCreating.value = false
  }
}

async function joinRoom() {
  if (!playerName.value.trim()) {
    error.value = 'Please enter your name'
    return
  }
  if (!joinRoomId.value.trim()) {
    error.value = 'Please enter a Room ID'
    return
  }

  isJoining.value = true
  error.value = null

  try {
    const response = await bounceBotClient.joinRoom({
      roomId: joinRoomId.value.trim(),
      playerName: playerName.value.trim(),
    })
    roomStore.setCurrentPlayer(response.playerId, playerName.value.trim())
    roomStore.setSinglePlayer(false)
    router.push(`/room/${response.room?.id}`)
  } catch (e) {
    error.value = translateJoinRoomError(e)
  } finally {
    isJoining.value = false
  }
}
</script>

<template>
  <div class="home">
    <img src="/logo_color.svg" alt="BounceBot" class="logo" />
    <img src="/name_light.svg" alt="BounceBot" class="name name-light" />
    <img src="/name_dark.svg" alt="BounceBot" class="name name-dark" />

    <div class="card">
      <!-- Play Solo - Primary action -->
      <div class="actions">
        <button
          class="btn primary solo-btn"
          :disabled="isLoading"
          @click="startSoloGame"
        >
          {{ isStartingSolo ? 'Starting...' : 'Play Solo' }}
        </button>
      </div>

      <!-- Play with Friends - Secondary action -->
      <div class="multiplayer-section">
        <button
          class="btn secondary multiplayer-toggle"
          :class="{ expanded: showMultiplayer }"
          :disabled="isLoading"
          @click="showMultiplayer = !showMultiplayer; if (!showMultiplayer) error = null"
        >
          Play with Friends
          <span class="toggle-icon">{{ showMultiplayer ? '▲' : '▼' }}</span>
        </button>

        <div v-if="showMultiplayer" class="multiplayer-options">
          <div class="form-group">
            <input
              id="playerName"
              v-model="playerName"
              type="text"
              placeholder="Enter your name"
              maxlength="20"
            />
          </div>

          <div class="multiplayer-buttons">
            <button
              class="btn"
              :disabled="isLoading"
              @click="createRoom"
            >
              {{ isCreating ? 'Creating...' : 'Create New Room' }}
            </button>
            <p class="join-text">or</p>
            <div class="join-row">
              <button
                class="btn join-room-btn"
                :disabled="isLoading"
                @click="joinRoom()"
              >
                {{ isJoining ? 'Joining...' : 'Join Room' }}
              </button>
              <input
                ref="roomIdInput"
                v-model="joinRoomId"
                type="text"
                placeholder="Room ID"
                class="room-id-inline"
                :disabled="isLoading"
                @keyup.enter="joinRoom"
              />
            </div>
          </div>
        </div>
      </div>

      <div v-if="error" class="error">{{ error }}</div>

      <div class="help-link">
        <router-link to="/help">How to Play</router-link>
      </div>

      <div v-if="lastRoom" class="return-section">
        <button class="btn return-btn" @click="returnToGame">
          Return to Game
        </button>
      </div>
    </div>

    <div class="version">{{ appVersion }}</div>
  </div>
</template>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
  min-height: 100vh;
  position: relative;
  overflow: hidden;
}

.home::before {
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
  .home::before {
    background-image: url('/pattern_light.svg');
    opacity: 0.4;
  }
}

.logo {
  width: 120px;
  height: auto;
  margin-bottom: 0.75rem;
}

.name {
  width: 320px;
  height: auto;
  margin-bottom: 1rem;
}

/* Base (dark mode): show dark name */
.name-light {
  display: none;
}

.name-dark {
  display: block;
}

/* Light mode: show light name */
@media (prefers-color-scheme: light) {
  .name-light {
    display: block;
  }

  .name-dark {
    display: none;
  }
}

.card {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 2rem;
  width: calc(100% - 2rem);
  max-width: 360px;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #aaa;
  font-size: 0.9rem;
}

.form-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #333;
  border-radius: 6px;
  background: #242424;
  color: #fff;
  font-size: 1rem;
  box-sizing: border-box;
  text-align: center;
}

.form-group input:focus {
  outline: none;
  border-color: #43a047;
}

.form-group input:-webkit-autofill,
.form-group input:-webkit-autofill:hover,
.form-group input:-webkit-autofill:focus {
  -webkit-box-shadow: 0 0 0 1000px #242424 inset;
  -webkit-text-fill-color: #fff;
  caret-color: #fff;
}

.actions {
  margin-top: 0;
}

.btn {
  width: 100%;
  padding: 0.75rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn.primary {
  background: #43a047;
  color: #fff;
}

.btn.primary:hover:not(:disabled) {
  background: #388e3c;
}

.btn.secondary {
  background: #1e88e5;
  color: #fff;
}

.btn.secondary:hover:not(:disabled) {
  background: #1976d2;
}

.solo-btn,
.multiplayer-toggle {
  font-size: 1.1rem;
  padding: 1rem;
}

.multiplayer-section {
  margin-top: 1rem;
}

.multiplayer-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.multiplayer-toggle.expanded {
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
}

.toggle-icon {
  font-size: 0.7rem;
}

.multiplayer-options {
  margin-top: 0;
  padding: 1rem;
  background: #1565c0;
  border-radius: 0 0 8px 8px;
}

.multiplayer-buttons {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.multiplayer-buttons .btn {
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
  font-size: 1rem;
  height: 44px;
}

.multiplayer-buttons .btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.3);
}

.join-text {
  color: rgba(255, 255, 255, 0.8);
  font-size: 0.9rem;
  margin: 0 0 0;
}

.join-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.join-room-btn {
  flex: 1;
}

.room-id-inline {
  padding: 0.25rem 0.5rem;
  border: none;
  border-radius: 4px;
  background: #1a1a1a;
  color: #fff;
  font-size: 1rem;
  font-family: inherit;
  width: 80px;
  text-align: center;
}

.room-id-inline::placeholder {
  color: #888;
}

.room-id-inline:focus {
  outline: none;
  background: #242424;
}

.help-link {
  margin-top: 1.5rem;
  text-align: center;
}

.help-link a {
  display: inline-block;
  color: #ccc;
  text-decoration: none;
  font-size: 0.9rem;
  padding: 0.5rem 1.25rem;
  background: #333;
  border: none;
  border-radius: 6px;
  transition: all 0.2s;
}

.help-link a:hover {
  color: #fff;
  background: #444;
}

.error {
  margin-top: 1rem;
  color: #ff6b6b;
  font-size: 0.85rem;
  text-align: center;
}

.return-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid #333;
}

.btn.return-btn {
  background: #1e88e5;
  color: #fff;
}

.btn.return-btn:hover:not(:disabled) {
  background: #1976d2;
}

.version {
  position: fixed;
  bottom: 0.5rem;
  right: 0.75rem;
  color: #666;
  font-size: 0.75rem;
}
</style>
