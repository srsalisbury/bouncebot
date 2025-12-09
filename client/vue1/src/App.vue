<script setup lang="ts">
import GameBoard from './components/GameBoard.vue'
</script>

<template>
  <div class="app">
    <h1>BounceBot</h1>
    <GameBoard />
  </div>
</template>

<style scoped>
.app {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
  gap: 1.5rem;
}

h1 {
  color: #42b883;
  font-size: 2.5rem;
  margin: 0;
  letter-spacing: -0.02em;
}
</style>
