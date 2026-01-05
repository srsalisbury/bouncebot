<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function goBack() {
  router.push('/')
}
</script>

<template>
  <div class="help-view">
    <div class="help-content">
      <button class="back-btn" @click="goBack">← Back</button>

      <h1>How to Play BounceBot</h1>

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
    </div>
  </div>
</template>

<style scoped>
.help-view {
  height: 100vh;
  height: 100dvh;
  padding: 2rem;
  background: #f5f5f5;
  overflow-y: auto;
  box-sizing: border-box;
}

.help-content {
  max-width: 550px;
  margin: 0 auto;
  background: #fff;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.back-btn {
  background: none;
  border: none;
  color: #666;
  font-size: 1rem;
  cursor: pointer;
  padding: 0.5rem 0;
  margin-bottom: 1rem;
}

.back-btn:hover {
  color: #42b883;
}

h1 {
  color: #2d7a5a;
  margin: 0 0 2rem 0;
  font-size: 1.75rem;
}

.section {
  margin-bottom: 2rem;
}

h2 {
  color: #222;
  margin: 0 0 0.75rem 0;
  font-size: 1.2rem;
  border-bottom: 1px solid #ddd;
  padding-bottom: 0.5rem;
}

h3 {
  color: #333;
  margin: 1rem 0 0.5rem 0;
  font-size: 1rem;
}

p {
  color: #333;
  margin: 0 0 0.75rem 0;
  font-size: 0.95rem;
  line-height: 1.6;
  text-align: left;
}

ul {
  margin: 0 0 0.75rem 0;
  padding-left: 1.5rem;
  color: #333;
  font-size: 0.95rem;
  line-height: 1.8;
  text-align: left;
}

li {
  margin-bottom: 0.25rem;
}

kbd {
  background: #e8e8e8;
  color: #333;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: inherit;
  font-size: 0.85rem;
  border: 1px solid #ccc;
}
</style>
