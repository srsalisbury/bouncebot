// WebSocket service for real-time session updates

export type EventType = 'player_joined' | 'player_left' | 'game_started' | 'player_solved' | 'solution_retracted' | 'player_finished_solving' | 'player_ready_for_next' | 'game_ended'

export interface PlayerJoinedPayload {
  playerId: string
  playerName: string
}

export interface PlayerLeftPayload {
  playerId: string
}

export interface GameStartedPayload {
  // Empty - client should refresh session data
}

export interface PlayerSolvedPayload {
  playerId: string
  moveCount: number
}

export interface SolutionRetractedPayload {
  playerId: string
}

export interface PlayerFinishedSolvingPayload {
  playerId: string
}

export interface PlayerReadyForNextPayload {
  playerId: string
}

export interface MovePayload {
  robotId: number
  x: number
  y: number
}

export interface GameEndedPayload {
  winnerId: string
  winnerName: string
  moves: MovePayload[]
}

export interface WebSocketEvent {
  type: EventType
  payload: PlayerJoinedPayload | PlayerLeftPayload | GameStartedPayload | PlayerSolvedPayload | SolutionRetractedPayload | PlayerFinishedSolvingPayload | PlayerReadyForNextPayload | GameEndedPayload
}

type EventHandler = (event: WebSocketEvent) => void

const WS_URL = 'ws://localhost:8080/ws'
const RECONNECT_DELAY = 3000

class WebSocketService {
  private ws: WebSocket | null = null
  private sessionId: string | null = null
  private eventHandler: EventHandler | null = null
  private reconnectTimeout: number | null = null
  private shouldReconnect = false

  connect(sessionId: string, onEvent: EventHandler): void {
    this.sessionId = sessionId
    this.eventHandler = onEvent
    this.shouldReconnect = true
    this.doConnect()
  }

  private doConnect(): void {
    if (!this.sessionId) return

    const url = `${WS_URL}?sessionId=${this.sessionId}`
    console.log('WebSocket: connecting to', url)

    this.ws = new WebSocket(url)

    this.ws.onopen = () => {
      console.log('WebSocket: connected')
    }

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as WebSocketEvent
        console.log('WebSocket: received event', data.type)
        this.eventHandler?.(data)
      } catch (e) {
        console.error('WebSocket: failed to parse message', e)
      }
    }

    this.ws.onclose = () => {
      console.log('WebSocket: disconnected')
      this.ws = null
      if (this.shouldReconnect) {
        this.scheduleReconnect()
      }
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket: error', error)
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimeout) return

    console.log(`WebSocket: reconnecting in ${RECONNECT_DELAY}ms`)
    this.reconnectTimeout = window.setTimeout(() => {
      this.reconnectTimeout = null
      this.doConnect()
    }, RECONNECT_DELAY)
  }

  disconnect(): void {
    this.shouldReconnect = false
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.sessionId = null
    this.eventHandler = null
  }
}

export const websocketService = new WebSocketService()
