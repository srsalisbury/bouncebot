<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { bounceBotClient } from '../services/connectClient'
import { useRoomStore } from '../stores/roomStore'

const router = useRouter()
const roomStore = useRoomStore()

const playerName = ref(roomStore.currentPlayerName ?? '')
const joinRoomId = ref('')
const isStartingSolo = ref(false)
const isCreating = ref(false)
const isJoining = ref(false)
const showMultiplayer = ref(false)
const error = ref<string | null>(null)

const lastRoom = computed(() => roomStore.lastRoomId)
const isLoading = computed(() => isStartingSolo.value || isCreating.value || isJoining.value)

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
    const room = await bounceBotClient.joinRoom({
      roomId: joinRoomId.value.trim(),
      playerName: playerName.value.trim(),
    })
    // Find ourselves in the players list (we're the last one added)
    const player = room.players[room.players.length - 1]
    if (player) {
      roomStore.setCurrentPlayer(player.id, player.name)
    }
    roomStore.setSinglePlayer(false)
    router.push(`/room/${room.id}`)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to join room'
  } finally {
    isJoining.value = false
  }
}
</script>

<template>
  <div class="home">
    <h1 class="title">BounceBot</h1>
    <p class="subtitle">A Ricochet Robots puzzle game</p>

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
          @click="showMultiplayer = !showMultiplayer"
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
              <input
                id="roomId"
                v-model="joinRoomId"
                type="text"
                placeholder="Room ID"
                class="room-id-input"
                @keyup.enter="joinRoom"
              />
              <button
                class="btn"
                :disabled="isLoading"
                @click="joinRoom"
              >
                {{ isJoining ? 'Joining...' : 'Join Room' }}
              </button>
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
  </div>
</template>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
  min-height: 100vh;
}

.title {
  color: #42b883;
  margin: 0;
  font-size: 2.5rem;
}

.subtitle {
  color: #888;
  margin: 0.5rem 0 2rem;
}

.card {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 2rem;
  width: 100%;
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
}

.form-group input:focus {
  outline: none;
  border-color: #42b883;
}

.actions {
  margin-top: 1rem;
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
  background: #42b883;
  color: #fff;
}

.btn.primary:hover:not(:disabled) {
  background: #3aa876;
}

.btn.secondary {
  background: #333;
  color: #fff;
}

.btn.secondary:hover:not(:disabled) {
  background: #444;
}

.solo-btn {
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

.toggle-icon {
  font-size: 0.7rem;
}

.multiplayer-options {
  margin-top: 1rem;
  padding: 1rem;
  background: #242424;
  border-radius: 8px;
}

.multiplayer-buttons {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.join-row {
  display: flex;
  gap: 0.5rem;
}

.join-text {
  color: #888;
  font-size: 0.9rem;
  margin: 0 0 0;
}

.room-id-input {
  flex: 1;
  padding: 0.75rem;
  border: 1px solid #333;
  border-radius: 6px;
  background: #1a1a1a;
  color: #fff;
  font-size: 1rem;
  box-sizing: border-box;
}

.room-id-input:focus {
  outline: none;
  border-color: #42b883;
}

.join-row .btn {
  width: auto;
  padding: 0.75rem 1.25rem;
}

.help-link {
  margin-top: 1.5rem;
  text-align: center;
}

.help-link a {
  color: #888;
  text-decoration: none;
  font-size: 0.9rem;
}

.help-link a:hover {
  color: #42b883;
}

.error {
  margin-top: 1rem;
  padding: 0.75rem;
  background: rgba(229, 57, 53, 0.1);
  border: 1px solid #e53935;
  border-radius: 6px;
  color: #e53935;
  font-size: 0.9rem;
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
</style>
