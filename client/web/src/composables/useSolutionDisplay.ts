import { computed } from 'vue'
import type { PlayerSolution } from '../gen/bouncebot_pb'
import type { Timestamp } from '@bufbuild/protobuf/wkt'
import { getFormattedTimes, calculateDurationSeconds } from '../services/timeUtils'

export interface SolutionDisplayProps {
  playerSolutions: PlayerSolution[]
  currentPlayerSolution?: PlayerSolution | null
  topThreeSolutions: PlayerSolution[]
  solverSolutions: PlayerSolution[]
  singlePlayer?: boolean
  gameStartedAt?: Timestamp
}

export function useSolutionDisplay(props: SolutionDisplayProps) {
  // Check if current player is in top 3
  const isCurrentPlayerInTopThree = computed(() => {
    if (!props.currentPlayerSolution) return false
    return props.topThreeSolutions.some(s => s.playerId === props.currentPlayerSolution?.playerId)
  })

  // Calculate the starting index for solver solutions in the combined array
  // New order: [top 3, solver, current player (if not in top 3)]
  const solverStartIndex = computed(() => {
    if (props.singlePlayer) {
      return props.playerSolutions.length
    }
    // Multiplayer: after top solutions
    return props.topThreeSolutions.length
  })

  // Calculate the starting index for current player's solution (after solver)
  const currentPlayerStartIndex = computed(() => {
    return solverStartIndex.value + props.solverSolutions.length
  })

  // Check if any player solutions are displayed before solver (for divider logic)
  const hasTopThreeSolutions = computed(() => {
    if (props.singlePlayer) {
      return props.playerSolutions.length > 0
    }
    return props.topThreeSolutions.length > 0
  })

  // Check if current player solution should be shown (not already in top 3)
  const showCurrentPlayerSolution = computed(() => {
    return !!props.currentPlayerSolution && !isCurrentPlayerInTopThree.value
  })

  // Collision-aware time formatting map
  const formattedTimesMap = computed(() => {
    if (!props.gameStartedAt) return new Map<string, string>()

    // Gather all solutions we know about to ensure consistent collision resolution
    const allSols = [
      ...(props.playerSolutions ?? []),
      ...(props.topThreeSolutions ?? []),
      props.currentPlayerSolution
    ].filter((s): s is PlayerSolution => !!s && !!s.solvedAt)

    // Deduplicate by ID
    const uniqueSols = new Map<string, PlayerSolution>()
    allSols.forEach(s => uniqueSols.set(s.playerId, s))

    const timeData = Array.from(uniqueSols.values()).map(s => ({
      id: s.playerId,
      seconds: calculateDurationSeconds(props.gameStartedAt!, s.solvedAt!)
    }))

    return getFormattedTimes(timeData)
  })

  function getSolveTime(playerId: string): string {
    return formattedTimesMap.value.get(playerId) ?? ''
  }

  return {
    isCurrentPlayerInTopThree,
    solverStartIndex,
    currentPlayerStartIndex,
    hasTopThreeSolutions,
    showCurrentPlayerSolution,
    formattedTimesMap,
    getSolveTime
  }
}
