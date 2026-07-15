<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/roomStore'
import { bounceBotClient } from '../services/connectClient'

const props = defineProps<{
  code: string
}>()

const router = useRouter()
const roomStore = useRoomStore()

const error = ref<string | null>(null)

onMounted(async () => {
  try {
    // Matches HomeView.vue's startSoloGame bootstrap: fixed default player
    // name, no prompt, straight into solo play - just seeded with the
    // shared board instead of a random one.
    const response = await bounceBotClient.createRoomFromSharedBoard({
      playerName: 'Player',
      shareCode: props.code,
    })
    roomStore.setCurrentPlayerId(response.playerId, response.sessionToken)
    roomStore.setSinglePlayer(true)

    // Apply saved display settings, same as any other new solo room.
    if (roomStore.showSolverMoveCount || roomStore.showSolverSolutions || roomStore.minSolutionLength > 1) {
      await bounceBotClient.updateRoomSettings({
        roomId: response.room!.id,
        sessionToken: response.sessionToken,
        settings: {
          showSolverMoveCount: roomStore.showSolverMoveCount,
          showSolverSolutions: roomStore.showSolverSolutions,
          minSolutionLength: roomStore.minSolutionLength,
        },
      })
    }

    // replace, not push - this landing route is a one-shot redirect, not
    // somewhere the back button should return to.
    router.replace(`/room/${response.room!.id}`)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to open this shared board'
  }
})
</script>

<template>
  <div class="share-landing">
    <div v-if="error" class="error-card">
      <h1>Link Unavailable</h1>
      <p>This link looks broken, or it came from a newer version of BounceBot.</p>
      <router-link to="/" class="btn">Back to Home</router-link>
    </div>
    <div v-else class="loading">
      Loading shared board&hellip;
    </div>
  </div>
</template>

<style scoped>
.share-landing {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 2rem;
  text-align: center;
}

.loading {
  color: #aaa;
  font-size: 1.1rem;
}

.error-card {
  max-width: 360px;
}

.error-card h1 {
  color: #ff6b6b;
  font-size: 1.4rem;
  margin: 0 0 0.75rem;
}

.error-card p {
  color: #aaa;
  margin: 0 0 1.5rem;
  line-height: 1.5;
}

.btn {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  background: #43a047;
  color: #fff;
  text-decoration: none;
}

.btn:hover {
  background: #388e3c;
}

@media (prefers-color-scheme: light) {
  .loading {
    color: #666;
  }

  .error-card p {
    color: #666;
  }
}
</style>
