<script setup lang="ts">
import GameBoard from './components/GameBoard.vue'
</script>

<template>
  <div class="app">
    <GameBoard />
  </div>
</template>

<style scoped>
.app {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
}
</style>
