import { onUnmounted, watch, type Ref } from 'vue'
import { loadSession } from '../api/auth'
import { getAppConfig } from '../config/loadConfig'
import type { DouNiuRoom, DouNiuState } from '../types/douniu'

type SocketStatus = 'idle' | 'connecting' | 'open' | 'closed'

function buildWsUrl(path: string) {
  const { wsBaseUrl } = getAppConfig()
  const token = loadSession()?.token
  if (!token) {
    throw new Error('未登录')
  }
  const base = wsBaseUrl.replace(/\/$/, '')
  const url = new URL(`${base}${path}`)
  url.searchParams.set('token', token)
  return url.toString()
}

function shouldApplyGameState(current: DouNiuState | null, next: DouNiuState) {
  if (!current) return true
  if (current.id !== next.id) return true
  if (current.phase !== next.phase) return true
  if ((next.events?.length ?? 0) > 0) return true
  if (current.banker_index !== next.banker_index) return true
  if (current.turn_deadline_unix !== next.turn_deadline_unix) return true
  if (current.message !== next.message) return true
  for (let i = 0; i < next.players.length; i += 1) {
    const a = current.players[i]
    const b = next.players[i]
    if (!a || !b) return true
    if (a.grab_done !== b.grab_done || a.bet_done !== b.bet_done) return true
    if (a.grab_mult !== b.grab_mult || a.bet_mult !== b.bet_mult) return true
    if (a.chips !== b.chips || a.round_delta !== b.round_delta) return true
  }
  return false
}

function shouldApplyRoom(current: DouNiuRoom | null, next: DouNiuRoom) {
  if (!current) return true
  return (
    current.status !== next.status ||
    current.game_id !== next.game_id ||
    JSON.stringify(current.players) !== JSON.stringify(next.players)
  )
}

export function useDouNiuGameSocket(options: {
  gameId: Ref<string | undefined>
  enabled: Ref<boolean>
  currentState: Ref<DouNiuState | null>
  onState: (state: DouNiuState) => void | Promise<void>
  onStatus?: (status: SocketStatus) => void
}) {
  let socket: WebSocket | null = null
  let reconnectTimer: number | null = null
  let stopped = false

  function setStatus(status: SocketStatus) {
    options.onStatus?.(status)
  }

  function cleanupSocket() {
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (socket) {
      socket.onopen = null
      socket.onmessage = null
      socket.onerror = null
      socket.onclose = null
      socket.close()
      socket = null
    }
  }

  function scheduleReconnect(connect: () => void) {
    if (stopped || reconnectTimer !== null) return
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = null
      connect()
    }, 2000)
  }

  function connect() {
    cleanupSocket()
    if (stopped || !options.enabled.value) {
      setStatus('idle')
      return
    }
    const gameId = options.gameId.value
    if (!gameId) {
      setStatus('idle')
      return
    }

    setStatus('connecting')
    try {
      socket = new WebSocket(buildWsUrl(`/douniu/games/${gameId}`))
    } catch {
      setStatus('closed')
      scheduleReconnect(connect)
      return
    }

    socket.onopen = () => setStatus('open')

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(String(event.data)) as { type?: string; state?: DouNiuState }
        if (payload.type !== 'game_state' || !payload.state) return
        if (!shouldApplyGameState(options.currentState.value, payload.state)) return
        void options.onState(payload.state)
      } catch {
        // ignore malformed payloads
      }
    }

    socket.onerror = () => setStatus('closed')

    socket.onclose = () => {
      setStatus('closed')
      if (!stopped && options.enabled.value) {
        scheduleReconnect(connect)
      }
    }
  }

  const stopWatch = watch(
    [options.gameId, options.enabled],
    () => {
      stopped = false
      connect()
    },
    { immediate: true },
  )

  onUnmounted(() => {
    stopped = true
    stopWatch()
    cleanupSocket()
    setStatus('idle')
  })

  return {
    reconnect: () => {
      stopped = false
      connect()
    },
    disconnect: () => {
      stopped = true
      cleanupSocket()
      setStatus('idle')
    },
  }
}

export function useDouNiuRoomSocket(options: {
  roomId: Ref<string | undefined>
  enabled: Ref<boolean>
  currentRoom: Ref<DouNiuRoom | null>
  onRoom: (room: DouNiuRoom) => void | Promise<void>
  onStatus?: (status: SocketStatus) => void
}) {
  let socket: WebSocket | null = null
  let reconnectTimer: number | null = null
  let stopped = false

  function setStatus(status: SocketStatus) {
    options.onStatus?.(status)
  }

  function cleanupSocket() {
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (socket) {
      socket.onopen = null
      socket.onmessage = null
      socket.onerror = null
      socket.onclose = null
      socket.close()
      socket = null
    }
  }

  function scheduleReconnect(connect: () => void) {
    if (stopped || reconnectTimer !== null) return
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = null
      connect()
    }, 2000)
  }

  function connect() {
    cleanupSocket()
    if (stopped || !options.enabled.value) {
      setStatus('idle')
      return
    }
    const roomId = options.roomId.value
    if (!roomId) {
      setStatus('idle')
      return
    }

    setStatus('connecting')
    try {
      socket = new WebSocket(buildWsUrl(`/douniu/rooms/${roomId}`))
    } catch {
      setStatus('closed')
      scheduleReconnect(connect)
      return
    }

    socket.onopen = () => setStatus('open')

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(String(event.data)) as { type?: string; room?: DouNiuRoom }
        if (payload.type !== 'room' || !payload.room) return
        if (!shouldApplyRoom(options.currentRoom.value, payload.room)) return
        void options.onRoom(payload.room)
      } catch {
        // ignore malformed payloads
      }
    }

    socket.onerror = () => setStatus('closed')

    socket.onclose = () => {
      setStatus('closed')
      if (!stopped && options.enabled.value) {
        scheduleReconnect(connect)
      }
    }
  }

  const stopWatch = watch(
    [options.roomId, options.enabled],
    () => {
      stopped = false
      connect()
    },
    { immediate: true },
  )

  onUnmounted(() => {
    stopped = true
    stopWatch()
    cleanupSocket()
    setStatus('idle')
  })

  return {
    reconnect: () => {
      stopped = false
      connect()
    },
    disconnect: () => {
      stopped = true
      cleanupSocket()
      setStatus('idle')
    },
  }
}
