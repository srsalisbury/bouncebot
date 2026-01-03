// Animation timing constants
export const ANIMATION_TIMING = {
  MOVE_DELAY: 150,      // Delay between animated moves (ms)
  REPLAY_DELAY: 500,    // Delay before starting replay after reset (ms)
} as const

// Schedule a sequence of callbacks with delays between them
export function animateSequence<T>(
  items: T[],
  onStep: (item: T, index: number) => void,
  delayMs: number = ANIMATION_TIMING.MOVE_DELAY
): number {
  items.forEach((item, i) => {
    setTimeout(() => onStep(item, i), i * delayMs)
  })
  return items.length * delayMs
}

// Schedule a sequence in reverse order
export function animateSequenceReverse<T>(
  items: T[],
  onStep: (item: T, index: number) => void,
  delayMs: number = ANIMATION_TIMING.MOVE_DELAY
): number {
  const reversed = items.slice().reverse()
  reversed.forEach((item, i) => {
    const originalIndex = items.length - 1 - i
    setTimeout(() => onStep(item, originalIndex), i * delayMs)
  })
  return items.length * delayMs
}

// Schedule a callback after a delay
export function scheduleAfter(delayMs: number, callback: () => void): void {
  setTimeout(callback, delayMs)
}

// Chain two animations: first runs, then after it completes, second runs
export function chainAnimations(
  firstDurationMs: number,
  secondAnimation: () => number
): number {
  let secondDuration = 0
  setTimeout(() => {
    secondDuration = secondAnimation()
  }, firstDurationMs)
  // Return estimated total time (can't know second duration ahead of time)
  return firstDurationMs + secondDuration
}
