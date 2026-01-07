<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const scrollOffset = ref(0)

function goBack() {
  router.push('/')
}

function onScroll(event: Event) {
  const target = event.target as HTMLElement
  // Move pattern at 10% of scroll speed for subtle parallax
  scrollOffset.value = target.scrollTop * 0.1
}

function onKeyDown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    goBack()
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
})
</script>

<template>
  <div class="help-view" @scroll="onScroll">
    <div
      class="background-pattern"
      :style="{ transform: `translate(-50%, calc(-50% - ${scrollOffset}px)) rotate(22.5deg)` }"
    ></div>
    <h1 class="page-title">How to <span class="play-text">Play</span></h1>
    <div class="help-content">
      <section class="section">
        <h2>What is BounceBot?</h2>
        <p>
          BounceBot is a puzzle game based on Ricochet Robots. Each puzzle presents you with
          a grid containing four colored robots and a target square. Your goal is to figure out
          how to move the robots so that the target robot lands on the target square.
        </p>
        <p>
          The twist: robots don't stop when you want them to. Once a robot starts moving,
          it slides in that direction until it hits something — a wall, the edge of the board,
          or another robot. This means you'll often need to use other robots as blockers to
          get the target robot where it needs to go.
        </p>
      </section>

      <section class="section">
        <h2>Goal</h2>
        <p>
          Move the target robot to the target square in as few moves as possible.
          The target square is shown as a colored square with a number matching the robot
          you need to move there.
        </p>
      </section>

      <section class="section">
        <h2>How Robots Move</h2>
        <p>
          When you move a robot in a direction, it slides in a straight line until it hits:
        </p>
        <ul>
          <li>A wall</li>
          <li>The edge of the board</li>
          <li>Another robot</li>
        </ul>
        <p>
          Robots cannot stop in the middle of the board on their own. Plan your moves
          carefully and use other robots as blockers when needed.
        </p>
      </section>

      <section class="section">
        <h2>Playing Solo</h2>
        <p>
          Click "Play Solo" to start solving puzzles immediately. You can skip to the next
          puzzle anytime by clicking "Next Puzzle". Your stats (puzzles solved vs attempted)
          are tracked in the header.
        </p>
      </section>

      <section class="section">
        <h2>Playing with Friends</h2>
        <p>
          Create a room and share the link with friends. Everyone works on the same puzzle
          simultaneously, racing to find the shortest solution. When you've found a solution
          (or given up), click "I'm Finished". Once everyone is done, the winning solution
          is revealed and you can move on to the next puzzle.
        </p>
      </section>

      <section class="section">
        <h2>Controls</h2>
        <h3>Selecting Robots</h3>
        <ul>
          <li>Click on a robot to select it</li>
          <li><kbd>1</kbd>-<kbd>4</kbd> — Select robot by number</li>
        </ul>

        <h3>Moving Robots</h3>
        <ul>
          <li>Swipe on the board (touch devices)</li>
          <li><kbd>↑</kbd><kbd>↓</kbd><kbd>←</kbd><kbd>→</kbd> — Arrow keys</li>
          <li><kbd>W</kbd><kbd>A</kbd><kbd>S</kbd><kbd>D</kbd> — WASD keys</li>
        </ul>

        <h3>Managing Solutions</h3>
        <ul>
          <li><kbd>z</kbd> / <kbd>u</kbd> / <kbd>Esc</kbd> — Undo last move</li>
          <li>Hold <kbd>Undo</kbd> button — Reset current solution</li>
          <li><kbd>n</kbd> / <kbd>+</kbd> — Start a new solution attempt</li>
          <li><kbd>Shift</kbd>+<kbd>←</kbd>/<kbd>→</kbd> — Switch between solutions</li>
          <li><kbd>Shift</kbd>+<kbd>D</kbd> — Delete current solution</li>
        </ul>

        <h3>Other</h3>
        <ul>
          <li><kbd>?</kbd> — Show help overlay (during game)</li>
          <li><kbd>l</kbd> — Toggle leaderboard (multiplayer, after game ends)</li>
          <li><kbd>x</kbd> — Leave game and return to home</li>
        </ul>
      </section>

      <section class="section">
        <h2>Tips</h2>
        <ul>
          <li>Use other robots as blockers to stop the target robot where you need it.</li>
          <li>The colored dots show where robots started and the path they've taken.</li>
          <li>You can work on multiple solution attempts simultaneously — use the solution
              columns to track different approaches.</li>
          <li>Sometimes a longer-looking path is actually shorter in moves!</li>
        </ul>
      </section>

      <button class="back-btn" @click="goBack">← Back</button>
    </div>
  </div>
</template>

<style scoped>
@font-face {
  font-family: 'Conthrax';
  src: url('/fonts/ConthraxRg-Bold.eot');
  src: url('/fonts/ConthraxRg-Bold.eot?#iefix') format('embedded-opentype'),
      url('/fonts/ConthraxRg-Bold.woff2') format('woff2'),
      url('/fonts/ConthraxRg-Bold.woff') format('woff'),
      url('/fonts/ConthraxRg-Bold.ttf') format('truetype'),
      url('/fonts/ConthraxRg-Bold.otf') format('opentype');
  font-weight: bold;
  font-style: normal;
  font-display: swap;
}

.help-view {
  height: 100vh;
  height: 100dvh;
  padding: 2rem;
  overflow-y: auto;
  box-sizing: border-box;
  position: relative;
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE/Edge */
}

.help-view::-webkit-scrollbar {
  display: none; /* Chrome/Safari */
}

.background-pattern {
  position: fixed;
  top: 50%;
  left: 50%;
  width: 185vmax;
  height: 185vmax;
  background-image: url('/pattern_dark.svg');
  background-repeat: no-repeat;
  background-size: 100% 100%;
  z-index: 0;
  opacity: 0.7;
  pointer-events: none;
}

@media (prefers-color-scheme: light) {
  .background-pattern {
    background-image: url('/pattern_light.svg');
    opacity: 0.4;
  }
}

.help-content {
  max-width: 550px;
  margin: 0 auto;
  background: #1a1a1a;
  padding: 2rem;
  border-radius: 12px;
  position: relative;
  z-index: 1;
}

.back-btn {
  display: inline-block;
  background: #333;
  border: none;
  border-radius: 6px;
  color: #ccc;
  font-size: 0.9rem;
  cursor: pointer;
  padding: 0.5rem 1.25rem;
  margin-top: 1rem;
  transition: all 0.2s;
}

.back-btn:hover {
  color: #fff;
  background: #444;
}

.page-title {
  font-family: 'Conthrax', sans-serif;
  color: #fff;
  margin: 0 auto 1rem;
  font-size: 2.25rem;
  text-align: center;
  text-transform: uppercase;
  position: relative;
  z-index: 1;
}

.play-text {
  color: #1e88e5;
}

@media (prefers-color-scheme: light) {
  .page-title {
    color: #000;
  }
}

.section {
  margin-bottom: 2rem;
}

h2 {
  color: #eee;
  margin: 0 0 0.75rem 0;
  font-size: 1.2rem;
  border-bottom: 1px solid #444;
  padding-bottom: 0.5rem;
}

h3 {
  color: #ddd;
  margin: 1rem 0 0.5rem 0;
  font-size: 1rem;
}

p {
  color: #bbb;
  margin: 0 0 0.75rem 0;
  font-size: 0.95rem;
  line-height: 1.6;
  text-align: left;
}

ul {
  margin: 0 0 0.75rem 0;
  padding-left: 1.5rem;
  color: #bbb;
  font-size: 0.95rem;
  line-height: 1.8;
  text-align: left;
}

li {
  margin-bottom: 0.25rem;
}

kbd {
  background: #333;
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: inherit;
  font-size: 0.85rem;
  border: 1px solid #555;
}
</style>
