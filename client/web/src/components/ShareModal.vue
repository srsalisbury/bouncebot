<script setup lang="ts">
import { ref, watch } from 'vue'
import QRCode from 'qrcode'

const props = defineProps<{
  show: boolean
  url: string
}>()

const emit = defineEmits<{
  close: []
}>()

const qrDataUrl = ref<string | null>(null)
const copyLabel = ref('Copy Link')
let copyResetTimeout: ReturnType<typeof setTimeout> | null = null

watch(
  () => [props.show, props.url] as const,
  async ([show, url]) => {
    if (!show || !url) return
    copyLabel.value = 'Copy Link'
    qrDataUrl.value = await QRCode.toDataURL(url, { margin: 1, width: 240 })
  },
  { immediate: true }
)

function handleBackdropClick(event: MouseEvent) {
  if (event.target === event.currentTarget) {
    emit('close')
  }
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(props.url)
  } catch {
    // Fallback for older browsers or when clipboard API fails
    const textArea = document.createElement('textarea')
    textArea.value = props.url
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
  }

  copyLabel.value = 'Copied!'
  if (copyResetTimeout) clearTimeout(copyResetTimeout)
  copyResetTimeout = setTimeout(() => {
    copyLabel.value = 'Copy Link'
  }, 1500)
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-backdrop" @click="handleBackdropClick">
      <div class="modal share-modal">
        <button class="close-btn" @click="emit('close')">×</button>
        <h2>Share Board</h2>

        <div class="qr-wrapper">
          <img v-if="qrDataUrl" :src="qrDataUrl" alt="QR code for the shared board link" class="qr-image" />
        </div>

        <button class="copy-link-btn" @click="copyLink">{{ copyLabel }}</button>
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
  width: 90%;
  position: relative;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
}

.share-modal {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.close-btn {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  background: none;
  border: none;
  color: #888;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

h2 {
  margin: 0 0 1.25rem 0;
  color: #43a047;
  font-size: 1.25rem;
}

.qr-wrapper {
  background: #fff;
  padding: 1rem;
  border-radius: 8px;
  width: 240px;
  height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qr-image {
  width: 100%;
  height: 100%;
  display: block;
}

.copy-link-btn {
  margin-top: 1.25rem;
  width: 100%;
  padding: 0.6rem 1rem;
  background: #43a047;
  border: none;
  border-radius: 6px;
  color: #fff;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.copy-link-btn:hover {
  background: #388e3c;
}

@media (prefers-color-scheme: light) {
  .modal {
    background: #fff;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
  }

  .close-btn {
    color: #666;
  }

  .close-btn:hover {
    color: #333;
    background: rgba(0, 0, 0, 0.1);
  }
}
</style>
