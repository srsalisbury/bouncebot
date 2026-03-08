<script setup lang="ts">
import { ref, watch } from 'vue'
import '../diagram-styles.css' // shared SVG diagram classes

type TabId = 'basics' | 'controls' | 'features'
const activeTab = ref<TabId>('basics')
const contentRef = ref<HTMLElement | null>(null)

const isTouchDevice = 'ontouchstart' in window || navigator.maxTouchPoints > 0

watch(activeTab, () => {
  contentRef.value?.scrollTo(0, 0)
  // Also scroll parent if we're inside a scrollable container
  contentRef.value?.closest('.modal, .help-content')?.scrollTo(0, 0)
})

defineExpose({ contentRef })
</script>

<template>
  <div ref="contentRef" class="help-tabs">
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'basics' }" @click="activeTab = 'basics'">Basics</button>
      <button class="tab" :class="{ active: activeTab === 'controls' }" @click="activeTab = 'controls'">Controls</button>
      <button class="tab" :class="{ active: activeTab === 'features' }" @click="activeTab = 'features'">Features</button>
    </div>

    <!-- BASICS TAB -->
    <div v-if="activeTab === 'basics'" class="tab-content">
      <div class="section">
        <p>BounceBot is a puzzle game inspired by classic sliding-robot board games. Each puzzle presents you with a grid containing four colored robots and a target square. Your goal is to figure out how to move the robots so that the target robot lands on the target square in as few moves as possible.</p>
        <p>The twist: robots don't stop when you want them to. Once a robot starts moving, it slides in that direction until it hits something. This means you'll often need to use other robots as blockers to get the target robot where it needs to go.</p>
      </div>

      <div class="section">
        <h3>Goal</h3>
        <svg class="diagram" viewBox="0 0 192 128">
          <rect width="192" height="128" class="d-bg"/>
          <line x1="16" y1="0" x2="16" y2="128" class="d-grid"/>
          <line x1="48" y1="0" x2="48" y2="128" class="d-grid"/>
          <line x1="80" y1="0" x2="80" y2="128" class="d-grid"/>
          <line x1="112" y1="0" x2="112" y2="128" class="d-grid"/>
          <line x1="144" y1="0" x2="144" y2="128" class="d-grid"/>
          <line x1="176" y1="0" x2="176" y2="128" class="d-grid"/>
          <line x1="0" y1="16" x2="192" y2="16" class="d-grid"/>
          <line x1="0" y1="48" x2="192" y2="48" class="d-grid"/>
          <line x1="0" y1="80" x2="192" y2="80" class="d-grid"/>
          <line x1="0" y1="112" x2="192" y2="112" class="d-grid"/>
          <line x1="142" y1="80" x2="178" y2="80" class="d-wall" stroke-width="3"/>
          <line x1="176" y1="46" x2="176" y2="82" class="d-wall" stroke-width="3"/>
          <rect x="147" y="51" width="26" height="26" rx="1" fill="#e53935"/>
          <circle cx="160" cy="64" r="10" class="d-bg-fill"/>
          <text x="160" y="65" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-tnum">1</text>
          <circle cx="32" cy="64" r="11" fill="#e53935" stroke="black" stroke-width="1"/>
          <text x="32" y="65" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke">1</text>
          <line x1="45" y1="64" x2="139" y2="64" class="d-arrow" stroke-width="2" stroke-dasharray="4 3"/>
          <polygon points="139,60 147,64 139,68" class="d-arrow-fill"/>
        </svg>
        <p>Move the <strong>target robot</strong> to the matching colored square. The target square shows the number of the robot you need to move there.</p>
      </div>

      <div class="section">
        <h3>How robots move</h3>
        <svg class="diagram" viewBox="0 0 192 160">
          <rect width="192" height="160" class="d-bg"/>
          <line x1="16" y1="0" x2="16" y2="160" class="d-grid"/>
          <line x1="48" y1="0" x2="48" y2="160" class="d-grid"/>
          <line x1="80" y1="0" x2="80" y2="160" class="d-grid"/>
          <line x1="112" y1="0" x2="112" y2="160" class="d-grid"/>
          <line x1="144" y1="0" x2="144" y2="160" class="d-grid"/>
          <line x1="176" y1="0" x2="176" y2="160" class="d-grid"/>
          <line x1="0" y1="16" x2="192" y2="16" class="d-grid"/>
          <line x1="0" y1="48" x2="192" y2="48" class="d-grid"/>
          <line x1="0" y1="80" x2="192" y2="80" class="d-grid"/>
          <line x1="0" y1="112" x2="192" y2="112" class="d-grid"/>
          <line x1="0" y1="144" x2="192" y2="144" class="d-grid"/>
          <line x1="14" y1="144" x2="50" y2="144" class="d-wall" stroke-width="3"/>
          <line x1="142" y1="112" x2="178" y2="112" class="d-wall" stroke-width="3"/>
          <line x1="176" y1="110" x2="176" y2="146" class="d-wall" stroke-width="3"/>
          <rect x="147" y="115" width="26" height="26" rx="1" fill="#1e88e5"/>
          <circle cx="160" cy="128" r="10" class="d-bg-fill"/>
          <text x="160" y="129" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-tnum">2</text>
          <circle cx="32" cy="32" r="11" fill="#1e88e5" stroke="black" stroke-width="1" opacity="0.3"/>
          <text x="32" y="33" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke" opacity="0.3">2</text>
          <circle cx="32" cy="128" r="11" fill="#1e88e5" stroke="black" stroke-width="1" opacity="0.5"/>
          <text x="32" y="129" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke" opacity="0.5">2</text>
          <circle cx="160" cy="128" r="11" fill="#1e88e5" stroke="black" stroke-width="1"/>
          <text x="160" y="129" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke">2</text>
          <line x1="32" y1="45" x2="32" y2="109" class="d-arrow" stroke-width="2"/>
          <polygon points="28,109 32,117 36,109" class="d-arrow-fill"/>
          <line x1="45" y1="128" x2="141" y2="128" class="d-arrow" stroke-width="2"/>
          <polygon points="141,124 149,128 141,132" class="d-arrow-fill"/>
          <rect x="40" y="70" width="14" height="14" rx="3" class="d-badge"/>
          <text x="47" y="78" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-badge-text">1</text>
          <rect x="86" y="111" width="14" height="14" rx="3" class="d-badge"/>
          <text x="93" y="119" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-badge-text">2</text>
        </svg>
        <p>Robots slide in a straight line until they hit a <strong>wall</strong>, the <strong>board edge</strong>, or <strong>another robot</strong>. They cannot stop in the middle of the board.</p>
      </div>

      <div class="section">
        <h3>Strategy: use blockers</h3>
        <svg class="diagram" viewBox="0 0 192 160">
          <rect width="192" height="160" class="d-bg"/>
          <line x1="16" y1="0" x2="16" y2="160" class="d-grid"/>
          <line x1="48" y1="0" x2="48" y2="160" class="d-grid"/>
          <line x1="80" y1="0" x2="80" y2="160" class="d-grid"/>
          <line x1="112" y1="0" x2="112" y2="160" class="d-grid"/>
          <line x1="144" y1="0" x2="144" y2="160" class="d-grid"/>
          <line x1="176" y1="0" x2="176" y2="160" class="d-grid"/>
          <line x1="0" y1="16" x2="192" y2="16" class="d-grid"/>
          <line x1="0" y1="48" x2="192" y2="48" class="d-grid"/>
          <line x1="0" y1="80" x2="192" y2="80" class="d-grid"/>
          <line x1="0" y1="112" x2="192" y2="112" class="d-grid"/>
          <line x1="0" y1="144" x2="192" y2="144" class="d-grid"/>
          <line x1="142" y1="144" x2="178" y2="144" class="d-wall" stroke-width="3"/>
          <line x1="110" y1="48" x2="146" y2="48" class="d-wall" stroke-width="3"/>
          <line x1="112" y1="46" x2="112" y2="82" class="d-wall" stroke-width="3"/>
          <rect x="115" y="51" width="26" height="26" rx="1" fill="#e53935"/>
          <circle cx="128" cy="64" r="10" class="d-bg-fill"/>
          <text x="128" y="65" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-tnum">1</text>
          <circle cx="160" cy="32" r="11" fill="#1e88e5" stroke="black" stroke-width="1" opacity="0.3"/>
          <text x="160" y="33" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke" opacity="0.3">2</text>
          <circle cx="160" cy="128" r="11" fill="#1e88e5" stroke="black" stroke-width="1"/>
          <text x="160" y="129" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke">2</text>
          <circle cx="32" cy="128" r="11" fill="#e53935" stroke="black" stroke-width="1" opacity="0.3"/>
          <text x="32" y="129" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke" opacity="0.3">1</text>
          <circle cx="128" cy="128" r="11" fill="#e53935" stroke="black" stroke-width="1" opacity="0.5"/>
          <text x="128" y="129" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke" opacity="0.5">1</text>
          <circle cx="128" cy="64" r="11" fill="#e53935" stroke="black" stroke-width="1"/>
          <text x="128" y="65" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke">1</text>
          <line x1="160" y1="45" x2="160" y2="109" class="d-arrow" stroke-width="2"/>
          <polygon points="156,109 160,117 164,109" class="d-arrow-fill"/>
          <line x1="45" y1="128" x2="109" y2="128" class="d-arrow" stroke-width="2"/>
          <polygon points="109,124 117,128 109,132" class="d-arrow-fill"/>
          <line x1="128" y1="115" x2="128" y2="83" class="d-arrow" stroke-width="2"/>
          <polygon points="124,83 128,75 132,83" class="d-arrow-fill"/>
          <rect x="168" y="70" width="14" height="14" rx="3" class="d-badge"/>
          <text x="175" y="78" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-badge-text">1</text>
          <rect x="70" y="134" width="14" height="14" rx="3" class="d-badge"/>
          <text x="77" y="142" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-badge-text">2</text>
          <rect x="106" y="90" width="14" height="14" rx="3" class="d-badge"/>
          <text x="113" y="98" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-badge-text">3</text>
        </svg>
        <p>Move <strong>other robots</strong> into the path to act as blockers. The best solutions use as few total moves as possible.</p>
      </div>
    </div>

    <!-- CONTROLS TAB -->
    <div v-if="activeTab === 'controls'" class="tab-content">
      <template v-if="!isTouchDevice">
        <div class="section">
          <h3>Keyboard &amp; mouse</h3>
          <ul class="controls-list">
            <li>Click a robot or press <kbd>1</kbd><kbd>2</kbd><kbd>3</kbd><kbd>4</kbd> to select it</li>
            <li><kbd>&#x2191;</kbd><kbd>&#x2193;</kbd><kbd>&#x2190;</kbd><kbd>&#x2192;</kbd> or <kbd>W</kbd><kbd>A</kbd><kbd>S</kbd><kbd>D</kbd> move selected robot</li>
            <li><kbd>z</kbd> / <kbd>u</kbd> / <kbd>Esc</kbd> undo last move (or click Undo)</li>
            <li><kbd>n</kbd> / <kbd>+</kbd> start new solution attempt</li>
            <li><kbd>Shift+&#x2190;</kbd> / <kbd>Shift+&#x2192;</kbd> switch between solutions</li>
            <li><kbd>Shift+D</kbd> delete current solution</li>
            <li><kbd>p</kbd> replay current solution</li>
            <li><kbd>?</kbd> toggle this help</li>
          </ul>
        </div>

        <div class="section">
          <h3>Touch controls</h3>
          <ul class="controls-list">
            <li>Tap a robot to select it</li>
            <li>Swipe on the board to move the selected robot</li>
          </ul>
        </div>
      </template>

      <template v-else>
        <div class="section">
          <h3>Touch</h3>
          <ul class="controls-list">
            <li>Tap a robot to select it</li>
            <li>Swipe on the board to move the selected robot</li>
            <li>Tap Undo to undo the last move</li>
            <li>Tap solution pills below the board to switch solutions</li>
          </ul>
        </div>

        <div class="section">
          <h3>Keyboard shortcuts</h3>
          <ul class="controls-list">
            <li>Click a robot or press <kbd>1</kbd><kbd>2</kbd><kbd>3</kbd><kbd>4</kbd> to select it</li>
            <li><kbd>&#x2191;</kbd><kbd>&#x2193;</kbd><kbd>&#x2190;</kbd><kbd>&#x2192;</kbd> or <kbd>W</kbd><kbd>A</kbd><kbd>S</kbd><kbd>D</kbd> move selected robot</li>
            <li><kbd>z</kbd> / <kbd>u</kbd> / <kbd>Esc</kbd> undo last move</li>
            <li><kbd>n</kbd> / <kbd>+</kbd> start new solution attempt</li>
            <li><kbd>Shift+&#x2190;</kbd> / <kbd>Shift+&#x2192;</kbd> switch between solutions</li>
            <li><kbd>Shift+D</kbd> delete current solution</li>
            <li><kbd>p</kbd> replay current solution</li>
            <li><kbd>?</kbd> toggle this help</li>
          </ul>
        </div>
      </template>

      <div class="section">
        <h3>After game ends</h3>
        <ul class="controls-list">
          <li v-if="!isTouchDevice"><kbd>Shift+&#x2190;</kbd> / <kbd>Shift+&#x2192;</kbd> switch between submitted solutions</li>
          <li v-else>Swipe or tap to switch between submitted solutions</li>
          <li><kbd v-if="!isTouchDevice">p</kbd><span v-else>Tap replay</span> to replay the currently shown solution</li>
        </ul>
      </div>
    </div>

    <!-- FEATURES TAB -->
    <div v-if="activeTab === 'features'" class="tab-content">
      <div class="section">
        <h3>Playing solo</h3>
        <p>In solo mode, you can skip to the next puzzle anytime with "Next Puzzle". Your stats (puzzles solved vs attempted) are tracked in the header.</p>
      </div>

      <div class="section">
        <h3>Solutions</h3>
        <p>You can work on up to <strong>4 solution attempts</strong> at the same time. Each solution tracks its own moves independently, so you can explore different strategies in parallel.</p>
        <img src="/help/solutions-dark.png" alt="Solutions panel" class="screenshot screenshot-dark no-border" />
        <img src="/help/solutions-light.png" alt="Solutions panel" class="screenshot screenshot-light no-border" />
        <ul>
          <li>Start a new attempt with <kbd v-if="!isTouchDevice">n</kbd><span v-else>the + button</span></li>
          <li>Switch between solutions with <kbd v-if="!isTouchDevice">Shift+&#x2190;/&#x2192;</kbd><span v-else>the solution pills</span></li>
          <li>Delete an attempt with <kbd v-if="!isTouchDevice">Shift+D</kbd><span v-else>the &#x00d7; button on the solution pill</span></li>
          <li>Colored dots on the board show each robot's starting position and path</li>
        </ul>
      </div>

      <div class="section">
        <h3>Settings</h3>
        <p><strong>BBot</strong> is the built-in solver that finds the optimal solution for every puzzle. Tap the gear icon in the header to open settings, where you can choose to:</p>
        <ul>
          <li><strong>Show BBot's move count</strong> during the game, so you know what to aim for</li>
          <li><strong>Show BBot's solution</strong> in the post-game review</li>
        </ul>
        <p class="settings-note">In multiplayer, only the room creator can access settings.</p>
        <img src="/help/settings-dark.png" alt="Settings modal" class="screenshot screenshot-dark" />
        <img src="/help/settings-light.png" alt="Settings modal" class="screenshot screenshot-light" />
        <p>Shortcuts: long-press the gear icon to quickly toggle the move count. Long-press the move count display to hide it.</p>
        <p class="screenshot-caption">With BBot's move count enabled:</p>
        <img src="/help/solver-count-dark.png" alt="BBot move count in header" class="screenshot screenshot-dark no-border screenshot-small" />
        <img src="/help/solver-count-light.png" alt="BBot move count in header" class="screenshot screenshot-light no-border screenshot-small" />
      </div>

      <div class="section">
        <h3>After game ends</h3>
        <p>When a game ends, you can review the submitted solutions. In multiplayer, the top 3 are ranked by fewest moves, then fastest solve time. Your own solution is always shown, even if not in the top 3. If enabled in settings, BBot's solution appears as well.</p>
        <img src="/help/post-game-dark.png" alt="Post-game review" class="screenshot screenshot-dark no-border" />
        <img src="/help/post-game-light.png" alt="Post-game review" class="screenshot screenshot-light no-border" />
      </div>

      <div class="section">
        <h3>Multiplayer</h3>
        <p>From the home screen, enter your name and create a room. Share the room link with others, or they can join with the Room ID.</p>
        <ul>
          <li>The room creator starts each round and can access room settings</li>
          <li>Everyone solves the same puzzle simultaneously</li>
          <li>When you're done, tap <strong>"I'm Finished"</strong> to lock in your best solution</li>
          <li>Solutions are scored by fewest moves first, then fastest solve time</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<style scoped>
.help-tabs {
  text-align: left;
}

.tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid #333;
  margin-bottom: 1.25rem;
}

.tab {
  flex: 1;
  padding: 0.5rem 0.75rem;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: #888;
  font-size: 0.9rem;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
}

.tab:hover {
  color: #ccc;
}

.tab.active {
  color: #43a047;
  border-bottom-color: #43a047;
}

.tab-content {
  min-height: 200px;
}

.section {
  margin-bottom: 1.25rem;
}

.section:last-child {
  margin-bottom: 0;
}

h3 {
  margin: 0 0 0.5rem 0;
  color: #eee;
  font-size: 1rem;
}

p {
  margin: 0 0 0.5rem 0;
  color: #bbb;
  font-size: 0.9rem;
  line-height: 1.5;
}

p:last-child {
  margin-bottom: 0;
}

.screenshot-caption {
  margin-top: 0.75rem;
  margin-bottom: 0;
  font-size: 0.85rem;
  font-style: italic;
  color: #999;
}

.settings-note {
  font-size: 0.8rem;
  color: #999;
  font-style: italic;
}

ul {
  margin: 0;
  padding-left: 1.25rem;
  color: #bbb;
  font-size: 0.9rem;
  line-height: 1.7;
}

.controls-list {
  list-style: none;
  padding-left: 0;
}

.controls-list li {
  padding: 0.15rem 0;
}

kbd {
  background: #333;
  color: #fff;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: inherit;
  font-size: 0.8rem;
  margin-right: 1px;
}

.secondary-section {
  padding-top: 1rem;
  border-top: 1px solid #333;
}

.secondary-section h3 {
  color: #888;
  font-size: 0.9rem;
}

.screenshot {
  width: 100%;
  max-width: 300px;
  max-height: 250px;
  object-fit: contain;
  display: block;
  margin: 0.75rem auto;
  border-radius: 6px;
  border: 1px solid #333;
}

.screenshot.no-border {
  border: none;
}

.screenshot.screenshot-small {
  max-width: 120px;
  max-height: 50px;
}

.screenshot-light {
  display: none;
}

.screenshot-dark {
  display: block;
}

@media (prefers-color-scheme: light) {
  .tabs {
    border-bottom-color: #ddd;
  }

  .tab {
    color: #999;
  }

  .tab:hover {
    color: #555;
  }

  .tab.active {
    color: #43a047;
  }

  h3 {
    color: #333;
  }

  p, ul {
    color: #555;
  }

  kbd {
    background: #e0e0e0;
    color: #333;
  }

  .secondary-section {
    border-top-color: #ddd;
  }

  .secondary-section h3 {
    color: #999;
  }

  .screenshot {
    border-color: #ddd;
  }

  .screenshot-dark {
    display: none;
  }

  .screenshot-light {
    display: block;
  }
}
</style>
