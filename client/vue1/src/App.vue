<template>
  <router-view />
  <!-- Landscape orientation blocker for mobile -->
  <div class="landscape-blocker">
    <div class="rotate-message">
      <div class="rotate-icon">📱</div>
      <p>Please rotate your device to portrait mode</p>
    </div>
  </div>
</template>

<style scoped>
.landscape-blocker {
  display: none;
  position: fixed;
  inset: 0;
  background: #121212;
  z-index: 9999;
  align-items: center;
  justify-content: center;
}

.rotate-message {
  text-align: center;
  color: #fff;
  padding: 2rem;
}

.rotate-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  animation: rotate-hint 2s ease-in-out infinite;
}

@keyframes rotate-hint {
  0%, 100% { transform: rotate(0deg); }
  25% { transform: rotate(-20deg); }
  75% { transform: rotate(20deg); }
}

.rotate-message p {
  font-size: 1.2rem;
  margin: 0;
}

/* Show blocker on touch devices in landscape orientation */
@media (pointer: coarse) and (orientation: landscape) {
  .landscape-blocker {
    display: flex;
  }
}
</style>
