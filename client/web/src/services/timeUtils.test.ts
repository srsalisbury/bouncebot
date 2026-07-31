import { describe, it, expect } from 'vitest'
import { formatDuration, getFormattedTimes, timestampToSeconds } from './timeUtils'

describe('timeUtils', () => {
  describe('timestampToSeconds', () => {
    it('returns 0 for an undefined timestamp', () => {
      expect(timestampToSeconds(undefined)).toBe(0)
    })

    it('includes sub-second precision from nanos', () => {
      // Two solves one second apart but only 100ms into their respective
      // seconds must not compare equal - dropping nanos would make them tie.
      const earlier = { seconds: 10n, nanos: 100_000_000 }
      const later = { seconds: 10n, nanos: 900_000_000 }
      expect(timestampToSeconds(earlier)).toBeLessThan(timestampToSeconds(later))
    })

    it('distinguishes two solves within the same second', () => {
      const a = { seconds: 5n, nanos: 200_000_000 }
      const b = { seconds: 5n, nanos: 800_000_000 }
      expect(timestampToSeconds(a)).not.toBe(timestampToSeconds(b))
    })
  })

  describe('formatDuration', () => {
    it('formats m:ss correctly (precision 0)', () => {
      expect(formatDuration(65.12, 0)).toBe('1:05')
      expect(formatDuration(65.89, 0)).toBe('1:05') // Floors, doesn't round
    })

    it('formats m:ss.f correctly (precision 1)', () => {
      expect(formatDuration(65.12, 1)).toBe('1:05.1')
      expect(formatDuration(65.89, 1)).toBe('1:05.8')
    })

    it('formats m:ss.ff correctly (precision 2)', () => {
      expect(formatDuration(65.12, 2)).toBe('1:05.12')
      expect(formatDuration(65.89, 2)).toBe('1:05.89')
    })

    it('handles floating point precision correctly', () => {
      expect(formatDuration(1.2, 1)).toBe('0:01.2')
    })
  })

  describe('getFormattedTimes (Smart Collision Resolution)', () => {
    it('formats unique times with base precision (m:ss)', () => {
      const items = [
        { id: '1', seconds: 65.12 }, // 1:05
        { id: '2', seconds: 72.50 }, // 1:12
      ]
      const results = getFormattedTimes(items)
      expect(results.get('1')).toBe('1:05')
      expect(results.get('2')).toBe('1:12')
    })

    it('resolves collision with tenths when sufficient', () => {
      const items = [
        { id: '1', seconds: 65.12 }, // 1:05.1
        { id: '2', seconds: 65.89 }, // 1:05.8
      ]
      // Both are '1:05' at precision 0
      const results = getFormattedTimes(items)
      expect(results.get('1')).toBe('1:05.1')
      expect(results.get('2')).toBe('1:05.8')
    })

    it('resolves collision with hundredths when tenths are equal', () => {
      const items = [
        { id: '1', seconds: 65.12 }, // 1:05.1
        { id: '2', seconds: 65.18 }, // 1:05.1
      ]
      // Both '1:05' -> Both '1:05.1' -> Need hundredths
      const results = getFormattedTimes(items)
      expect(results.get('1')).toBe('1:05.12')
      expect(results.get('2')).toBe('1:05.18')
    })

    it('resolves perfect ties identically at max precision', () => {
      const items = [
        { id: '1', seconds: 65.123 }, // 1:05.12
        { id: '2', seconds: 65.129 }, // 1:05.12
      ]
      // Both '1:05' -> Both '1:05.1' -> Both '1:05.12'
      const results = getFormattedTimes(items)
      expect(results.get('1')).toBe('1:05.12')
      expect(results.get('2')).toBe('1:05.12')
    })

    it('handles mixed groups (some unique, some colliding)', () => {
      const items = [
        { id: 'unique', seconds: 70.0 }, // 1:10
        { id: 'tie1', seconds: 65.12 },  // 1:05.1
        { id: 'tie2', seconds: 65.89 },  // 1:05.8
      ]
      const results = getFormattedTimes(items)
      expect(results.get('unique')).toBe('1:10')
      expect(results.get('tie1')).toBe('1:05.1')
      expect(results.get('tie2')).toBe('1:05.8')
    })

    it('handles complex mixed collisions', () => {
      const items = [
        { id: 'A', seconds: 65.12 }, // 1:05.12 (clashes with B at .1)
        { id: 'B', seconds: 65.18 }, // 1:05.18 (clashes with A at .1)
        { id: 'C', seconds: 65.89 }, // 1:05.8  (distinct tenth)
        { id: 'D', seconds: 80.00 }, // 1:20    (distinct minute/sec)
      ]
      
      // Group 1: A, B, C map to "1:05"
      //   -> A (1:05.1), B (1:05.1), C (1:05.8)
      //   -> C is unique at precision 1 -> "1:05.8"
      //   -> A, B clash at "1:05.1"
      //      -> A (1:05.12), B (1:05.18) -> Precision 2
      
      // Group 2: D maps to "1:20" -> unique -> "1:20"

      const results = getFormattedTimes(items)
      expect(results.get('D')).toBe('1:20')
      expect(results.get('C')).toBe('1:05.8')
      expect(results.get('A')).toBe('1:05.12')
      expect(results.get('B')).toBe('1:05.18')
    })

    it('enforces precision for near-zero times regardless of collisions', () => {
      const items = [
        { id: 'half', seconds: 0.5 },    // 0:00 -> 0:00.5
        { id: 'small', seconds: 0.05 },  // 0:00 -> 0:00.0 -> 0:00.05
        { id: 'tiny', seconds: 0.005 },  // 0:00 -> 0:00.0 -> 0:00.00 (truncates)
      ]
      const results = getFormattedTimes(items)
      expect(results.get('half')).toBe('0:00.5')
      expect(results.get('small')).toBe('0:00.05')
      expect(results.get('tiny')).toBe('0:00.00')
    })
  })
})
