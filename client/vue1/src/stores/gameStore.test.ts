import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useGameStore } from './gameStore'
import { MAX_SOLUTIONS } from '../constants'

describe('gameStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  describe('startNewSolution', () => {
    it('creates a new solution when under max capacity', () => {
      const store = useGameStore()
      expect(store.solutions.length).toBe(1)

      const result = store.startNewSolution()

      expect(result.newIndex).toBe(1)
      expect(result.deletedIndex).toBeNull()
      expect(store.solutions.length).toBe(2)
      expect(store.activeSolutionIndex).toBe(1)
    })

    it('auto-deletes when at max capacity', () => {
      const store = useGameStore()

      // Fill up to max solutions
      for (let i = 1; i < MAX_SOLUTIONS; i++) {
        store.startNewSolution()
      }
      expect(store.solutions.length).toBe(MAX_SOLUTIONS)

      // Add one more - should auto-delete
      const result = store.startNewSolution()

      expect(result.deletedIndex).not.toBeNull()
      expect(store.solutions.length).toBe(MAX_SOLUTIONS)
    })

    it('protects best submitted solution from deletion', () => {
      const store = useGameStore()

      // Fill up to max solutions
      for (let i = 1; i < MAX_SOLUTIONS; i++) {
        store.startNewSolution()
      }

      // Mark solution 0 as the best submitted
      const bestSubmittedIndex = 0

      // Add a new solution - should not delete index 0
      const result = store.startNewSolution(bestSubmittedIndex)

      expect(result.deletedIndex).not.toBe(bestSubmittedIndex)
      expect(store.solutions.length).toBe(MAX_SOLUTIONS)
    })
  })

  describe('findWorstSolutionIndex (via startNewSolution)', () => {
    // MAX_SOLUTIONS = 4, so we have indices 0, 1, 2, 3

    describe('when unsolved solutions exist', () => {
      it('deletes the most recent unsolved solution (highest index)', () => {
        const store = useGameStore()

        // Create max solutions: indices 0, 1, 2, 3
        for (let i = 1; i < MAX_SOLUTIONS; i++) {
          store.startNewSolution()
        }
        expect(store.solutions.length).toBe(MAX_SOLUTIONS)

        // Mark solutions 0, 1 as solved
        store.solutions[0].isSolved = true
        store.solutions[1].isSolved = true
        // Solutions 2, 3 are unsolved

        // Add new solution - should delete index 3 (most recent unsolved)
        const result = store.startNewSolution()

        expect(result.deletedIndex).toBe(3)
      })

      it('ignores move count when deleting unsolved solutions', () => {
        const store = useGameStore()

        // Create max solutions
        for (let i = 1; i < MAX_SOLUTIONS; i++) {
          store.startNewSolution()
        }

        // Solution 0 is unsolved with MANY moves
        store.solutions[0].moves = [
          { robotId: 0, direction: 'up', fromX: 0, fromY: 0, toX: 0, toY: 1 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 1, toX: 0, toY: 2 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 2, toX: 0, toY: 3 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 3, toX: 0, toY: 4 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 4, toX: 0, toY: 5 },
        ]

        // Solution 3 (last) is unsolved with NO moves (but most recent)
        store.solutions[3].moves = []

        // Add new solution - should delete index 3 (most recent unsolved), NOT index 0
        const result = store.startNewSolution()

        expect(result.deletedIndex).toBe(3)
      })

      it('skips solved solutions when looking for unsolved to delete', () => {
        const store = useGameStore()

        // Create max solutions
        for (let i = 1; i < MAX_SOLUTIONS; i++) {
          store.startNewSolution()
        }

        // Mark the last solution (3) as solved
        store.solutions[3].isSolved = true
        // Solution 2 is unsolved (most recent unsolved)

        // Add new solution - should delete index 2 (most recent unsolved)
        const result = store.startNewSolution()

        expect(result.deletedIndex).toBe(2)
      })

      it('skips best submitted when looking for unsolved to delete', () => {
        const store = useGameStore()

        // Create max solutions
        for (let i = 1; i < MAX_SOLUTIONS; i++) {
          store.startNewSolution()
        }

        // All are unsolved, but 3 is best submitted
        const bestSubmittedIndex = 3

        // Add new solution - should delete index 2 (most recent unsolved that isn't best)
        const result = store.startNewSolution(bestSubmittedIndex)

        expect(result.deletedIndex).toBe(2)
      })
    })

    describe('when all solutions are solved', () => {
      it('deletes the longest solution (most moves)', () => {
        const store = useGameStore()

        // Create max solutions, all solved
        for (let i = 1; i < MAX_SOLUTIONS; i++) {
          store.startNewSolution()
        }
        for (let i = 0; i < MAX_SOLUTIONS; i++) {
          store.solutions[i].isSolved = true
        }

        // Solution 1 has the most moves
        store.solutions[0].moves = [
          { robotId: 0, direction: 'up', fromX: 0, fromY: 0, toX: 0, toY: 1 },
        ]
        store.solutions[1].moves = [
          { robotId: 0, direction: 'up', fromX: 0, fromY: 0, toX: 0, toY: 1 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 1, toX: 0, toY: 2 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 2, toX: 0, toY: 3 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 3, toX: 0, toY: 4 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 4, toX: 0, toY: 5 },
        ]
        store.solutions[2].moves = []
        store.solutions[3].moves = []

        // Add new solution - should delete index 1 (longest)
        const result = store.startNewSolution()

        expect(result.deletedIndex).toBe(1)
      })

      it('protects best submitted even if it is the longest', () => {
        const store = useGameStore()

        // Create max solutions, all solved
        for (let i = 1; i < MAX_SOLUTIONS; i++) {
          store.startNewSolution()
        }
        for (let i = 0; i < MAX_SOLUTIONS; i++) {
          store.solutions[i].isSolved = true
        }

        // Solution 0 has the most moves AND is best submitted
        store.solutions[0].moves = [
          { robotId: 0, direction: 'up', fromX: 0, fromY: 0, toX: 0, toY: 1 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 1, toX: 0, toY: 2 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 2, toX: 0, toY: 3 },
        ]
        // Solution 1 is second longest
        store.solutions[1].moves = [
          { robotId: 0, direction: 'up', fromX: 0, fromY: 0, toX: 0, toY: 1 },
          { robotId: 0, direction: 'up', fromX: 0, fromY: 1, toX: 0, toY: 2 },
        ]

        // Add new solution with best submitted = 0
        const result = store.startNewSolution(0)

        // Should delete index 1 (second longest), not 0 (protected)
        expect(result.deletedIndex).toBe(1)
      })
    })
  })

  describe('deleteSolution', () => {
    it('removes solution at given index', () => {
      const store = useGameStore()
      store.startNewSolution()
      store.startNewSolution()
      expect(store.solutions.length).toBe(3)

      store.deleteSolution(1)

      expect(store.solutions.length).toBe(2)
    })

    it('does not delete when only one solution remains', () => {
      const store = useGameStore()
      expect(store.solutions.length).toBe(1)

      store.deleteSolution(0)

      expect(store.solutions.length).toBe(1)
    })

    it('adjusts activeSolutionIndex when deleting before active', () => {
      const store = useGameStore()
      store.startNewSolution()
      store.startNewSolution()
      store.switchSolution(2)
      expect(store.activeSolutionIndex).toBe(2)

      store.deleteSolution(0)

      expect(store.activeSolutionIndex).toBe(1)
    })
  })
})
