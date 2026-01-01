import { onMounted, onUnmounted, type Ref } from 'vue'
import type { Direction } from '../constants'

export interface SwipeOptions {
  /** Minimum distance in pixels to register as a swipe */
  minDistance?: number
  /** Element to attach touch listeners to */
  target: Ref<HTMLElement | null>
  /** Called when a swipe is detected */
  onSwipe: (direction: Direction) => void
  /** Whether swipe detection is enabled */
  enabled?: Ref<boolean>
}

export function useSwipe(options: SwipeOptions) {
  const { target, onSwipe, minDistance = 30, enabled } = options

  let startX = 0
  let startY = 0
  let isSwiping = false

  function handleTouchStart(event: TouchEvent) {
    if (enabled?.value === false) return
    const touch = event.touches[0]
    if (!touch) return
    startX = touch.clientX
    startY = touch.clientY
    isSwiping = true
  }

  function handleTouchMove(event: TouchEvent) {
    if (!isSwiping || enabled?.value === false) return
    const touch = event.touches[0]
    if (!touch) return

    const deltaY = touch.clientY - startY
    const absY = Math.abs(deltaY)

    // If vertical movement exceeds threshold, prevent default to stop pull-to-refresh
    if (absY > 10) {
      event.preventDefault()
    }
  }

  function handleTouchEnd(event: TouchEvent) {
    if (enabled?.value === false) {
      isSwiping = false
      return
    }
    const touch = event.changedTouches[0]
    if (!touch) {
      isSwiping = false
      return
    }
    const deltaX = touch.clientX - startX
    const deltaY = touch.clientY - startY

    const absX = Math.abs(deltaX)
    const absY = Math.abs(deltaY)

    isSwiping = false

    // Must exceed minimum distance
    if (absX < minDistance && absY < minDistance) {
      return
    }

    // Determine primary direction
    let direction: Direction
    if (absX > absY) {
      direction = deltaX > 0 ? 'right' : 'left'
    } else {
      direction = deltaY > 0 ? 'down' : 'up'
    }

    onSwipe(direction)
  }

  function attach() {
    const el = target.value
    if (!el) return
    el.addEventListener('touchstart', handleTouchStart, { passive: true })
    el.addEventListener('touchmove', handleTouchMove, { passive: false })
    el.addEventListener('touchend', handleTouchEnd, { passive: true })
  }

  function detach() {
    const el = target.value
    if (!el) return
    el.removeEventListener('touchstart', handleTouchStart)
    el.removeEventListener('touchmove', handleTouchMove)
    el.removeEventListener('touchend', handleTouchEnd)
  }

  onMounted(() => {
    attach()
  })

  onUnmounted(() => {
    detach()
  })

  return {
    attach,
    detach,
  }
}
