import { onMounted, onUnmounted, type Ref } from 'vue'
import type { Direction } from '../constants'

export interface SwipeStartInfo {
  /** X position relative to target element (0-1 normalized) */
  relativeX: number
  /** Y position relative to target element (0-1 normalized) */
  relativeY: number
}

export interface SwipeOptions {
  /** Minimum distance in pixels to register as a swipe */
  minDistance?: number
  /** Element to attach touch listeners to */
  target: Ref<HTMLElement | null>
  /** Called when a swipe is detected */
  onSwipe: (direction: Direction) => void
  /** Called when a swipe gesture starts, before direction is determined */
  onSwipeStart?: (info: SwipeStartInfo) => void
  /** Whether swipe detection is enabled */
  enabled?: Ref<boolean>
}

export function useSwipe(options: SwipeOptions) {
  const { target, onSwipe, onSwipeStart, minDistance = 50, enabled } = options

  let originX = 0
  let originY = 0
  let isSwiping = false
  let hasFiredSwipe = false  // Track if we've fired at least one swipe this gesture

  function handleTouchStart(event: TouchEvent) {
    if (enabled?.value === false) return
    const touch = event.touches[0]
    if (!touch) return
    originX = touch.clientX
    originY = touch.clientY
    isSwiping = true
    hasFiredSwipe = false

    // Call onSwipeStart with normalized position relative to target
    if (onSwipeStart && target.value) {
      const rect = target.value.getBoundingClientRect()
      const relativeX = (touch.clientX - rect.left) / rect.width
      const relativeY = (touch.clientY - rect.top) / rect.height
      onSwipeStart({ relativeX, relativeY })
    }
  }

  function handleTouchMove(event: TouchEvent) {
    if (!isSwiping || enabled?.value === false) return
    const touch = event.touches[0]
    if (!touch) return

    const deltaX = touch.clientX - originX
    const deltaY = touch.clientY - originY
    const absX = Math.abs(deltaX)
    const absY = Math.abs(deltaY)

    // Prevent default to stop pull-to-refresh and scrolling
    if (absX > 10 || absY > 10) {
      event.preventDefault()
    }

    // Check if we've moved enough to trigger a swipe
    if (absX >= minDistance || absY >= minDistance) {
      // Determine primary direction
      let direction: Direction
      if (absX > absY) {
        direction = deltaX > 0 ? 'right' : 'left'
      } else {
        direction = deltaY > 0 ? 'down' : 'up'
      }

      // Fire the swipe
      onSwipe(direction)
      hasFiredSwipe = true

      // Reset origin to current position for next potential swipe
      originX = touch.clientX
      originY = touch.clientY
    }
  }

  function handleTouchEnd(event: TouchEvent) {
    if (enabled?.value === false) {
      isSwiping = false
      return
    }

    // If we haven't fired any swipe yet, check if we should fire one on release
    if (!hasFiredSwipe) {
      const touch = event.changedTouches[0]
      if (touch) {
        const deltaX = touch.clientX - originX
        const deltaY = touch.clientY - originY
        const absX = Math.abs(deltaX)
        const absY = Math.abs(deltaY)

        // Fire swipe if exceeded threshold
        if (absX >= minDistance || absY >= minDistance) {
          let direction: Direction
          if (absX > absY) {
            direction = deltaX > 0 ? 'right' : 'left'
          } else {
            direction = deltaY > 0 ? 'down' : 'up'
          }
          onSwipe(direction)
        }
      }
    }

    isSwiping = false
    hasFiredSwipe = false
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
