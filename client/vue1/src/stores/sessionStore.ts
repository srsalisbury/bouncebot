import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useSessionStore = defineStore('session', () => {
  const currentPlayerName = ref<string | null>(null)

  function setCurrentPlayer(name: string) {
    currentPlayerName.value = name
  }

  function clear() {
    currentPlayerName.value = null
  }

  return {
    currentPlayerName,
    setCurrentPlayer,
    clear,
  }
})
