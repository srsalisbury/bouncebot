import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { websocketService } from './websocket'

// Minimal fake WebSocket: captures the most recently constructed instance so
// tests can drive its lifecycle callbacks directly, without a real socket.
class FakeWebSocket {
  static instances: FakeWebSocket[] = []
  onopen: (() => void) | null = null
  onclose: ((event: { code: number }) => void) | null = null
  onmessage: ((event: { data: string }) => void) | null = null
  onerror: ((error: unknown) => void) | null = null
  closed = false

  constructor(public url: string) {
    FakeWebSocket.instances.push(this)
  }

  close() {
    this.closed = true
  }
}

describe('WebSocketService resync handling', () => {
  beforeEach(() => {
    FakeWebSocket.instances = []
    vi.stubGlobal('WebSocket', FakeWebSocket)
    vi.useFakeTimers()
  })

  afterEach(() => {
    websocketService.disconnect()
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('fires onResync on the very first successful connection', () => {
    // A fresh page load rejoining a room this player was already in looks
    // like a "first connect" from this service instance's point of view, but
    // the server may still be catching up on marking them reconnected - the
    // initial loadRoom() on mount can race ahead of that. Firing here too
    // (cheap, idempotent) closes that gap.
    const onResync = vi.fn()
    websocketService.connect('ROOM', 'token', vi.fn(), undefined, onResync)

    FakeWebSocket.instances[0].onopen?.()

    expect(onResync).toHaveBeenCalledTimes(1)
  })

  it('fires onResync again when the socket reopens after dropping', () => {
    const onResync = vi.fn()
    websocketService.connect('ROOM', 'token', vi.fn(), undefined, onResync)

    FakeWebSocket.instances[0].onopen?.()
    expect(onResync).toHaveBeenCalledTimes(1)

    // Socket drops (e.g. network blip) and the service schedules a reconnect.
    FakeWebSocket.instances[0].onclose?.({ code: 1006 })
    vi.advanceTimersByTime(3000)

    expect(FakeWebSocket.instances).toHaveLength(2)
    FakeWebSocket.instances[1].onopen?.()

    expect(onResync).toHaveBeenCalledTimes(2)
  })

  it('does not fire onResync for a connection attempt that never opens', () => {
    const onResync = vi.fn()
    websocketService.connect('ROOM', 'token', vi.fn(), undefined, onResync)

    // First attempt fails before ever opening successfully.
    FakeWebSocket.instances[0].onclose?.({ code: 1006 })
    vi.advanceTimersByTime(3000)

    expect(FakeWebSocket.instances).toHaveLength(2)
    expect(onResync).not.toHaveBeenCalled()
  })
})
