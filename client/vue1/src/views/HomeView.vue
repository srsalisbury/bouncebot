<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { bounceBotClient } from '../services/connectClient'
import { useRoomStore } from '../stores/roomStore'

const router = useRouter()
const roomStore = useRoomStore()

const playerName = ref(roomStore.currentPlayerName ?? '')
const joinRoomId = ref('')
const isCreating = ref(false)
const isJoining = ref(false)
const error = ref<string | null>(null)

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
      <div class="form-group">
        <label for="playerName">Your Name</label>
        <input
          id="playerName"
          v-model="playerName"
          type="text"
          placeholder="Enter your name"
          maxlength="20"
          @keyup.enter="createRoom"
        />
      </div>

      <div class="actions">
        <button
          class="btn primary"
          :disabled="isCreating || isJoining"
          @click="createRoom"
        >
          {{ isCreating ? 'Creating...' : 'Create Room' }}
        </button>
      </div>

      <div class="divider">
        <span>or join existing</span>
      </div>

      <div class="form-group">
        <label for="roomId">Room ID</label>
        <input
          id="roomId"
          v-model="joinRoomId"
          type="text"
          placeholder="Enter room ID"
          @keyup.enter="joinRoom"
        />
      </div>

      <div class="actions">
        <button
          class="btn secondary"
          :disabled="isCreating || isJoining"
          @click="joinRoom"
        >
          {{ isJoining ? 'Joining...' : 'Join Room' }}
        </button>
      </div>

      <div v-if="error" class="error">{{ error }}</div>
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

.divider {
  display: flex;
  align-items: center;
  margin: 1.5rem 0;
  color: #666;
  font-size: 0.85rem;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: #333;
}

.divider span {
  padding: 0 1rem;
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
</style>
