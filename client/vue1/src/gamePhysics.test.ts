import { describe, it, expect } from 'vitest'
import { calculateDestination } from './gamePhysics'
import type { Direction } from './constants'
import type { Robot, Wall } from './stores/gameStore'
import cases from '../../../tests/physics_cases.json'

interface PhysicsCase {
  name: string
  boardSize: number
  vWalls: { x: number; y: number }[]
  hWalls: { x: number; y: number }[]
  robots: { id: number; x: number; y: number }[]
  move: { robotId: number; direction: string }
  expected: { x: number; y: number }
}

describe('physics matches server', () => {
  for (const tc of (cases as { cases: PhysicsCase[] }).cases) {
    it(tc.name, () => {
      // Build robots array
      const robots: Robot[] = tc.robots.map(r => ({
        id: r.id,
        x: r.x,
        y: r.y,
      }))

      // Build walls arrays
      const vWalls: Wall[] = tc.vWalls.map(w => ({ x: w.x, y: w.y }))
      const hWalls: Wall[] = tc.hWalls.map(w => ({ x: w.x, y: w.y }))

      // Find the robot to move
      const robot = robots.find(r => r.id === tc.move.robotId)!

      // Calculate destination
      const result = calculateDestination(
        robot,
        tc.move.direction as Direction,
        robots,
        vWalls,
        hWalls
      )

      expect(result).toEqual(tc.expected)
    })
  }
})
