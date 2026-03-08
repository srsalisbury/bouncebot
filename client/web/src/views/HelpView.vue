<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import HelpContent from '../components/HelpContent.vue'

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
    <div class="page-header">
      <button class="back-btn" @click="goBack">&larr;</button>
      <h1 class="page-title">How to <span class="play-text">Play</span></h1>
    </div>
    <div class="help-content">
      <HelpContent />
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
  background-size: auto 100%;
  background-position: center;
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

.page-header {
  max-width: 550px;
  margin: 0 auto 1rem;
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.back-btn {
  position: absolute;
  left: 0;
  background: none;
  border: none;
  color: #aaa;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  transition: color 0.2s;
}

.back-btn:hover {
  color: #fff;
}

.page-title {
  font-family: 'Conthrax', sans-serif;
  color: #fff;
  margin: 0;
  font-size: 2.25rem;
  text-align: center;
  text-transform: uppercase;
}

.play-text {
  color: #1e88e5;
}

@media (prefers-color-scheme: light) {
  .page-title {
    color: #000;
  }

  .help-content {
    background: #fff;
  }

  .back-btn {
    color: #666;
  }

  .back-btn:hover {
    color: #000;
  }
}
</style>
