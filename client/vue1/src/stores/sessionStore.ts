import { ref, watch } from 'vue'
import { defineStore } from 'pinia'

const STORAGE_KEY = 'bouncebot_player_name'

export const useSessionStore = defineStore('session', () => {
  // Load from localStorage on init
  const stored = localStorage.getItem(STORAGE_KEY)
  const currentPlayerName = ref<string | null>(stored)

  // Persist to localStorage when changed
  watch(currentPlayerName, (name) => {
    if (name) {
      localStorage.setItem(STORAGE_KEY, name)
    } else {
      localStorage.removeItem(STORAGE_KEY)
    }
  })

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
