import { ref } from 'vue'
import { defineStore } from 'pinia'
import { bounceBotClient } from '../services/connectClient'

export const useFeatureStore = defineStore('features', () => {
  const dailyChallengeEnabled = ref(false)
  const loaded = ref(false)

  async function fetchFeatures() {
    try {
      const resp = await bounceBotClient.getServerInfo({})
      dailyChallengeEnabled.value = resp.dailyChallengeEnabled
    } catch {
      dailyChallengeEnabled.value = false
    }
    loaded.value = true
  }

  return { dailyChallengeEnabled, loaded, fetchFeatures }
})
