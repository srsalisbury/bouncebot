<script setup lang="ts">
import { ref, watch } from 'vue'
import QRCode from 'qrcode'

const props = withDefaults(defineProps<{
  show: boolean
  url: string
  title?: string
  description?: string
  code?: string
}>(), {
  title: 'Share Board',
})

const emit = defineEmits<{
  close: []
}>()

const qrDataUrl = ref<string | null>(null)
const copyLabel = ref('Copy Link')
let copyResetTimeout: ReturnType<typeof setTimeout> | null = null

const QR_SIZE = 240

// logo_color_halo.svg is a generated asset (scripts/generate-logo-halo.js)
// combining the logo with a white halo already sized to its own silhouette -
// see that script for why. Regenerate it if logo_color.svg changes.
let logoImagePromise: Promise<HTMLImageElement> | null = null
function loadLogoImage(): Promise<HTMLImageElement> {
  if (!logoImagePromise) {
    logoImagePromise = new Promise((resolve, reject) => {
      const img = new Image()
      img.onload = () => resolve(img)
      img.onerror = reject
      img.src = '/logo_color_halo.svg'
    })
  }
  return logoImagePromise
}

async function generateQrWithLogo(url: string): Promise<string> {
  const canvas = document.createElement('canvas')
  // High error correction tolerates roughly up to ~30% of the code being
  // obscured - the logo below (plus its halo) stays well inside that budget
  // so the code remains reliably scannable.
  await QRCode.toCanvas(canvas, url, { margin: 1, width: QR_SIZE, errorCorrectionLevel: 'H' })

  const ctx = canvas.getContext('2d')
  if (!ctx) return canvas.toDataURL('image/png')

  try {
    const logo = await loadLogoImage()
    // ~13% of the QR's area (width² since the overlay is square). Verified
    // empirically (BarcodeDetector) that a pristine render stops decoding
    // around 50-55% width (~30% area, matching level H's theoretical
    // recovery budget) - staying well under that leaves real margin for
    // real-world scan noise a phone camera adds (glare, tilt, print quality)
    // that a perfect digital decode doesn't have to contend with.
    const logoSize = QR_SIZE * 0.36
    const logoOffset = (QR_SIZE - logoSize) / 2
    ctx.drawImage(logo, logoOffset, logoOffset, logoSize, logoSize)
  } catch {
    // Logo failed to load - fall back to a plain QR code rather than blocking sharing.
  }

  return canvas.toDataURL('image/png')
}

watch(
  () => [props.show, props.url] as const,
  async ([show, url]) => {
    if (!show || !url) return
    copyLabel.value = 'Copy Link'
    qrDataUrl.value = await generateQrWithLogo(url)
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
        <h2>{{ title }}</h2>
        <p v-if="description" class="modal-description">{{ description }}</p>

        <code v-if="code" class="code-display">{{ code }}</code>

        <div class="qr-wrapper">
          <img v-if="qrDataUrl" :src="qrDataUrl" alt="QR code for the shared link" class="qr-image" />
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

.modal-description {
  margin: -0.75rem 0 1.25rem 0;
  color: #aaa;
  font-size: 0.85rem;
}

.code-display {
  display: block;
  margin: 0 0 1.25rem 0;
  background: #1e88e5;
  padding: 0.5rem 1.25rem;
  border-radius: 6px;
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  color: #fff;
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

  .modal-description {
    color: #666;
  }
}
</style>
