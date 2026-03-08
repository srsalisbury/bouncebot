<script setup lang="ts">
import '../diagram-styles.css'

defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  dismiss: []
}>()

function handleBackdropClick(event: MouseEvent) {
  if (event.target === event.currentTarget) {
    emit('dismiss')
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-backdrop" @click="handleBackdropClick">
      <div class="modal">
        <h2>Welcome to BounceBot!</h2>

        <div class="section">
          <h3>Goal</h3>
          <!--
            6x4 grid at 32px, half-cell (16px) edges. Interior 5x3.
            Grid lines: x=16,48,80,112,144,176  y=16,48,80,112
            Interior cell centers: x=32,64,96,128,160  y=32,64,96
          -->
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
            <!-- Corner walls (bottom-right L) at target interior cell (4,1) = center (160,64) -->
            <line x1="142" y1="80" x2="178" y2="80" class="d-wall" stroke-width="3"/>
            <line x1="176" y1="46" x2="176" y2="82" class="d-wall" stroke-width="3"/>
            <!-- Target ring at (160,64) -->
            <rect x="147" y="51" width="26" height="26" rx="1" fill="#e53935"/>
            <circle cx="160" cy="64" r="10" class="d-bg-fill"/>
            <text x="160" y="65" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-tnum">1</text>
            <!-- Red robot at interior cell (0,1) = center (32,64) -->
            <circle cx="32" cy="64" r="11" fill="#e53935" stroke="black" stroke-width="1"/>
            <text x="32" y="65" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke">1</text>
            <!-- Dashed arrow toward target -->
            <line x1="45" y1="64" x2="139" y2="64" class="d-arrow" stroke-width="2" stroke-dasharray="4 3"/>
            <polygon points="139,60 147,64 139,68" class="d-arrow-fill"/>
          </svg>
          <p>Move the target robot to the matching colored square.</p>
        </div>

        <div class="section">
          <h3>How robots move</h3>
          <!--
            6x5 grid at 32px, half-cell (16px) edges. Interior 5x4.
            Grid lines: x=16,48,80,112,144,176  y=16,48,80,112,144
            Interior cell centers: x=32,64,96,128,160  y=32,64,96,128
          -->
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
            <!-- Wall at bottom of interior cell (0,3): y=144 -->
            <line x1="14" y1="144" x2="50" y2="144" class="d-wall" stroke-width="3"/>
            <!-- Corner walls (top-right L) at target interior cell (4,3) = center (160,128) -->
            <line x1="142" y1="112" x2="178" y2="112" class="d-wall" stroke-width="3"/>
            <line x1="176" y1="110" x2="176" y2="146" class="d-wall" stroke-width="3"/>
            <!-- Target ring at (160,128) -->
            <rect x="147" y="115" width="26" height="26" rx="1" fill="#1e88e5"/>
            <circle cx="160" cy="128" r="10" class="d-bg-fill"/>
            <text x="160" y="129" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-tnum">2</text>
            <!-- Ghost robot at interior (0,0) = (32,32) -->
            <circle cx="32" cy="32" r="11" fill="#1e88e5" stroke="black" stroke-width="1" opacity="0.3"/>
            <text x="32" y="33" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke" opacity="0.3">2</text>
            <!-- Intermediate ghost at interior (0,3) = (32,128) -->
            <circle cx="32" cy="128" r="11" fill="#1e88e5" stroke="black" stroke-width="1" opacity="0.5"/>
            <text x="32" y="129" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke" opacity="0.5">2</text>
            <!-- Solid robot at target interior (4,3) = (160,128) -->
            <circle cx="160" cy="128" r="11" fill="#1e88e5" stroke="black" stroke-width="1"/>
            <text x="160" y="129" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke">2</text>
            <!-- Move 1 arrow: down -->
            <line x1="32" y1="45" x2="32" y2="109" class="d-arrow" stroke-width="2"/>
            <polygon points="28,109 32,117 36,109" class="d-arrow-fill"/>
            <!-- Move 2 arrow: right -->
            <line x1="45" y1="128" x2="141" y2="128" class="d-arrow" stroke-width="2"/>
            <polygon points="141,124 149,128 141,132" class="d-arrow-fill"/>
            <!-- Move number badges -->
            <rect x="40" y="70" width="14" height="14" rx="3" class="d-badge"/>
            <text x="47" y="78" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-badge-text">1</text>
            <rect x="86" y="111" width="14" height="14" rx="3" class="d-badge"/>
            <text x="93" y="119" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-badge-text">2</text>
          </svg>
          <p>Robots slide in a straight line until they hit a wall, the board edge, or another robot.</p>
        </div>

        <div class="section">
          <h3>Tip</h3>
          <!--
            6x5 grid at 32px, half-cell (16px) edges. Interior 5x4.
            Same grid as diagram 2.
            Blue slides down into position, red bounces off blue then up to target.
          -->
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
            <!-- Wall at bottom of interior cell (4,3): y=144 -->
            <line x1="142" y1="144" x2="178" y2="144" class="d-wall" stroke-width="3"/>
            <!-- Corner walls (top-left L) at target interior cell (3,1) = center (128,64) -->
            <line x1="110" y1="48" x2="146" y2="48" class="d-wall" stroke-width="3"/>
            <line x1="112" y1="46" x2="112" y2="82" class="d-wall" stroke-width="3"/>
            <!-- Target ring at (128,64) -->
            <rect x="115" y="51" width="26" height="26" rx="1" fill="#e53935"/>
            <circle cx="128" cy="64" r="10" class="d-bg-fill"/>
            <text x="128" y="65" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-tnum">1</text>
            <!-- Ghost blue at interior (4,0) = (160,32) -->
            <circle cx="160" cy="32" r="11" fill="#1e88e5" stroke="black" stroke-width="1" opacity="0.3"/>
            <text x="160" y="33" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke" opacity="0.3">2</text>
            <!-- Solid blue at interior (4,3) = (160,128) -->
            <circle cx="160" cy="128" r="11" fill="#1e88e5" stroke="black" stroke-width="1"/>
            <text x="160" y="129" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke">2</text>
            <!-- Ghost red at interior (0,3) = (32,128) -->
            <circle cx="32" cy="128" r="11" fill="#e53935" stroke="black" stroke-width="1" opacity="0.3"/>
            <text x="32" y="129" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke" opacity="0.3">1</text>
            <!-- Intermediate red at interior (3,3) = (128,128) stopped by blue -->
            <circle cx="128" cy="128" r="11" fill="#e53935" stroke="black" stroke-width="1" opacity="0.5"/>
            <text x="128" y="129" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke" opacity="0.5">1</text>
            <!-- Solid red at target interior (3,1) = (128,64) -->
            <circle cx="128" cy="64" r="11" fill="#e53935" stroke="black" stroke-width="1"/>
            <text x="128" y="65" text-anchor="middle" dominant-baseline="middle" font-size="10" font-weight="bold" fill="white" stroke="black" stroke-width="0.5" paint-order="stroke">1</text>
            <!-- Move 1 arrow: blue slides down -->
            <line x1="160" y1="45" x2="160" y2="109" class="d-arrow" stroke-width="2"/>
            <polygon points="156,109 160,117 164,109" class="d-arrow-fill"/>
            <!-- Move 2 arrow: red slides right into blue -->
            <line x1="45" y1="128" x2="109" y2="128" class="d-arrow" stroke-width="2"/>
            <polygon points="109,124 117,128 109,132" class="d-arrow-fill"/>
            <!-- Move 3 arrow: red slides up to target -->
            <line x1="128" y1="115" x2="128" y2="83" class="d-arrow" stroke-width="2"/>
            <polygon points="124,83 128,75 132,83" class="d-arrow-fill"/>
            <!-- Move number badges -->
            <rect x="168" y="70" width="14" height="14" rx="3" class="d-badge"/>
            <text x="175" y="78" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-badge-text">1</text>
            <rect x="70" y="134" width="14" height="14" rx="3" class="d-badge"/>
            <text x="77" y="142" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-badge-text">2</text>
            <rect x="106" y="90" width="14" height="14" rx="3" class="d-badge"/>
            <text x="113" y="98" text-anchor="middle" dominant-baseline="middle" font-size="9" font-weight="bold" class="d-badge-text">3</text>
          </svg>
          <p>Move other robots into the path as blockers.</p>
        </div>

        <button class="got-it-btn" @click="emit('dismiss')">Got it</button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 1.5rem 2rem;
  max-width: 400px;
  max-height: 80vh;
  overflow-y: auto;
  position: relative;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
}

h2 {
  margin: 0 0 1rem 0;
  color: #43a047;
  font-size: 1.5rem;
}

.section {
  margin-bottom: 1.25rem;
}

h3 {
  margin: 0 0 0.5rem 0;
  color: #eee;
  font-size: 1rem;
}

p {
  margin: 0;
  color: #bbb;
  font-size: 0.9rem;
  line-height: 1.5;
}

.got-it-btn {
  display: block;
  width: 100%;
  padding: 0.75rem;
  background: #43a047;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  margin-top: 0.5rem;
}

.got-it-btn:hover {
  background: #388e3c;
}

@media (prefers-color-scheme: light) {
  .modal {
    background: #fff;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
  }

  h3 {
    color: #333;
  }

  p {
    color: #555;
  }
}
</style>
