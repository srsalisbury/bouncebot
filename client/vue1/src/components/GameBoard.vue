<script setup lang="ts">
const BOARD_SIZE = 16
const CELL_SIZE = 32

const boardPixelSize = BOARD_SIZE * CELL_SIZE
</script>

<template>
  <div
    class="board"
    :style="{
      width: `${boardPixelSize}px`,
      height: `${boardPixelSize}px`,
      gridTemplateColumns: `repeat(${BOARD_SIZE}, ${CELL_SIZE}px)`,
      gridTemplateRows: `repeat(${BOARD_SIZE}, ${CELL_SIZE}px)`,
    }"
  >
    <div
      v-for="i in BOARD_SIZE * BOARD_SIZE"
      :key="i"
      class="cell"
    />
  </div>
</template>

<style scoped>
.board {
  display: grid;
  background: #dddddd;
  border: 4px solid #546e7a;
}

.cell {
  border: 1px solid #37474f;
  box-sizing: border-box;
}
</style>
