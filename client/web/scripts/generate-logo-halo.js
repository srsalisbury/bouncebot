// Generates public/logo_color_halo.svg: the BounceBot mark with a white
// plus-shaped "halo" layered behind it, for embedding in the center of a QR
// code without the code's dark modules showing through.
//
// The halo is a simple "+", not a copy of the logo's own clover silhouette -
// but it sits as a uniform-width margin around the clover: the plus's 4
// inner (concave) corners sit CLOVER_CORNER_RATIO's corner distance further
// out from center (still a gap from the black outline, not touching it),
// and the plus's 4 outer arm tips sit that *same absolute margin* further
// out from the clover's own N/E/S/W peaks - not a scaled-up copy, which
// would leave a tight margin at the corners and a generous one at the
// peaks. CLOVER_CORNER_RATIO is measured from the logo's first <path> (the
// solid blob every other shape in the logo sits inside): sampling it with
// getPointAtLength/getCTM in a browser and looking for sharp tangent-angle
// jumps finds its 4 concave corners sit at (~=0.2583 * viewBoxWidth) from
// center along both axes - see git history on this file for the one-off
// script used to measure it. Re-measure and update the constant if
// logo_color.svg's silhouette changes shape.
//
// This is a manual, build-independent step - like `npm run generate` (buf)
// for the proto bindings, its output is committed to git, not regenerated on
// every `dev`/`build`. Run after editing public/logo_color.svg:
//   npm run generate:logo-halo
import { readFileSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { JSDOM } from 'jsdom'

const __dirname = dirname(fileURLToPath(import.meta.url))
const SRC_PATH = join(__dirname, '../public/logo_color.svg')
const OUT_PATH = join(__dirname, '../public/logo_color_halo.svg')

// How far the halo's margin bleeds beyond the mark's own outline, expressed
// as the scale-up of the clover's own inner-corner distance from center -
// sized to roughly match the width of the logo's own white strokes (the
// negative space carved out of the black ring), not the thinner black
// outline around them. The resulting absolute margin (see haloMargin below)
// is then reused for the outer arm reach too, so the halo's width reads as
// constant all the way around instead of tight at the corners and wide at
// the peaks.
const HALO_SCALE = 1.35

// The clover's own concave corners, as a fraction of the logo viewBox
// width, offset (±this, ±this) from center - see the file comment above for
// how this was measured.
const CLOVER_CORNER_RATIO = 0.2583

function main() {
  const svgText = readFileSync(SRC_PATH, 'utf8')

  const { window } = new JSDOM()
  const root = new window.DOMParser().parseFromString(svgText, 'image/svg+xml').documentElement

  const viewBox = root.getAttribute('viewBox')
  if (!viewBox) throw new Error('logo svg: no viewBox attribute')
  const rootStyle = root.getAttribute('style') ?? ''
  const [vx, vy, vw, vh] = viewBox.split(/\s+/).map(Number)
  const cx = vx + vw / 2
  const cy = vy + vh / 2

  // The clover's own silhouette exactly fills the viewBox (its lobe tips
  // touch the edges), so vw/2 is the clover's own N/E/S/W peak distance from
  // center. cloverCornerRadius is the clover's own concave-corner distance
  // from center (the corner sits at (±cloverCornerHalfOffset,
  // ±cloverCornerHalfOffset), so its radius is that offset times sqrt(2)).
  const cloverPeakRadius = vw / 2
  const cloverCornerHalfOffset = vw * CLOVER_CORNER_RATIO
  const cloverCornerRadius = cloverCornerHalfOffset * Math.SQRT2

  // haloMargin is the absolute gap HALO_SCALE implies at the corner -
  // applied uniformly at both the inner corners and the outer peaks, so the
  // halo reads as a constant-width margin around the whole mark.
  const haloMargin = cloverCornerRadius * (HALO_SCALE - 1)
  const armHalfLength = cloverPeakRadius + haloMargin
  const armHalfThickness = (cloverCornerRadius + haloMargin) / Math.SQRT2

  // The halo bleeds past the original viewBox - and since this file is
  // loaded via <img>, anything outside the viewBox gets hard-clipped, not
  // just overflow-hidden. Expand the viewBox around the same center, tight
  // to the arm tips (the plus's farthest points from center), so the arms
  // have room without clipping or wasted margin.
  const outViewBox = [cx - armHalfLength, cy - armHalfLength, armHalfLength * 2, armHalfLength * 2].join(' ')

  // 12-point plus outline, clockwise from the top arm's top-left corner.
  // The 4 concave (inner) corners are at (cx ± armHalfThickness, cy ±
  // armHalfThickness) - haloMargin further out than the clover's own inner
  // corners, matching the margin at the outer arm tips.
  const t = armHalfThickness
  const l = armHalfLength
  const plusPoints = [
    [cx - t, cy - l],
    [cx + t, cy - l],
    [cx + t, cy - t],
    [cx + l, cy - t],
    [cx + l, cy + t],
    [cx + t, cy + t],
    [cx + t, cy + l],
    [cx - t, cy + l],
    [cx - t, cy + t],
    [cx - l, cy + t],
    [cx - l, cy - t],
    [cx - t, cy - t],
  ]
  const plusPath = `M${plusPoints.map(([x, y]) => `${x},${y}`).join('L')}Z`

  const serializer = new window.XMLSerializer()
  const rawInner = Array.from(root.children)
    .map((child) => serializer.serializeToString(child))
    .join('\n  ')

  const composite = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!-- Generated by scripts/generate-logo-halo.js from logo_color.svg - do not edit directly. -->
<svg xmlns="http://www.w3.org/2000/svg" viewBox="${outViewBox}" style="${rootStyle}">
  <path fill="#fff" d="${plusPath}"/>
  ${rawInner}
</svg>
`

  writeFileSync(OUT_PATH, composite)
  console.log(`Wrote ${OUT_PATH}`)
}

main()
