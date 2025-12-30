# Mobile Support Implementation Plan

## Overview

Add mobile/touch support to the BounceBot Vue.js client, enabling gameplay on phones and tablets with various screen sizes.

## Current State

- **Board**: Fixed 512x512px (16x16 grid, 32px cells)
- **Controls**: Keyboard-only (arrow keys/WASD, number keys 1-4)
- **Layout**: Side-by-side board + solutions panel (desktop-optimized)
- **Styling**: Vanilla scoped CSS, minimal media queries

---

## Phase 1: Responsive Board Scaling

### 1.1 Make board scale to viewport

**File**: `client/vue1/src/components/GameBoard.vue`

- Use CSS `transform: scale()` to fit board in available space
- Calculate scale factor based on viewport width/height
- Preserve aspect ratio (board is always square)
- Add container with `overflow: hidden` to clip during transitions

```vue
const boardScale = computed(() => {
  const maxWidth = viewportWidth.value - padding
  const maxHeight = viewportHeight.value - headerHeight
  return Math.min(1, maxWidth / 512, maxHeight / 512)
})
```

### 1.2 Add viewport size tracking

**File**: `client/vue1/src/composables/useViewport.ts` (new)

- Track `window.innerWidth` and `window.innerHeight`
- Debounce resize events
- Provide reactive viewport dimensions

---

## Phase 2: Touch Controls

### 2.1 Robot selection via tap

**File**: `client/vue1/src/components/GameBoard.vue`

- Add `@click` / `@touchend` handlers to robot elements
- Tap on robot to select it
- Visual feedback: selected robot has distinct border/glow

### 2.2 Movement via swipe gestures

**File**: `client/vue1/src/composables/useSwipe.ts` (new)

- Detect swipe direction (up/down/left/right) on board area
- Minimum swipe distance threshold (30-50px)
- Call existing `moveRobot(direction)` function
- Visual feedback: brief directional indicator on swipe

---

## Phase 3: Responsive Layout

### 3.1 Stack layout on narrow screens

**File**: `client/vue1/src/views/RoomView.vue`

- Media query breakpoint: `max-width: 768px`
- Stack board above solutions panel (column layout)
- Full-width panels on mobile

### 3.2 Collapsible solutions drawer

**File**: `client/vue1/src/components/SolutionsDrawer.vue` (new)

- Bottom drawer that slides up when tapped
- Collapsed state: thin bar showing "Solutions (3)" count + undo button
- Expanded state: full solutions list with switch/delete controls
- Swipe down or tap header to collapse

### 3.3 Compact player header

**File**: `client/vue1/src/components/PlayersPanel.vue`

- Horizontal scrolling player list on mobile
- Smaller avatars/names
- Timer remains prominent

---

## Phase 4: Touch-Friendly UI

### 4.1 Larger tap targets

- Minimum 44x44px touch targets (Apple HIG)
- Increase button padding on mobile
- Add spacing between interactive elements

### 4.2 Prevent unwanted browser behaviors

**File**: `client/vue1/src/components/GameBoard.vue`

```css
.game-board {
  touch-action: none;  /* Prevent scroll/zoom on board */
  user-select: none;   /* Prevent text selection */
  -webkit-touch-callout: none;
}
```

### 4.3 Disable viewport zooming during gameplay

**File**: `client/vue1/index.html`

```html
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
```

---

## Files to Create

| File | Purpose |
|------|---------|
| `src/composables/useViewport.ts` | Reactive viewport dimensions |
| `src/composables/useSwipe.ts` | Swipe gesture detection |
| `src/components/SolutionsDrawer.vue` | Collapsible mobile solutions panel |

## Files to Modify

| File | Changes |
|------|---------|
| `src/components/GameBoard.vue` | Scaling, touch events, tap selection |
| `src/views/RoomView.vue` | Responsive stacked layout |
| `src/components/PlayersPanel.vue` | Compact mobile mode |
| `src/style.css` | Mobile media queries, touch styles |
| `index.html` | Viewport meta tag update |

---

## Implementation Order

1. `useViewport.ts` - Viewport tracking composable
2. `GameBoard.vue` - Add board scaling
3. `GameBoard.vue` - Add tap-to-select robot
4. `useSwipe.ts` - Swipe gesture composable
5. `GameBoard.vue` - Add swipe-to-move
6. `RoomView.vue` - Responsive stacked layout
7. `SolutionsDrawer.vue` - Collapsible solutions drawer
8. `PlayersPanel.vue` - Compact mobile mode
9. CSS polish and touch behavior fixes

---

## Testing Considerations

- Test on iOS Safari, Android Chrome
- Test various screen sizes (320px - 768px width)
- Test both portrait and landscape orientations
- Verify touch gestures don't conflict with browser gestures
