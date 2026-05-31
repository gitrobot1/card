import { getAppConfig } from '../config/loadConfig'
import { loadSession } from './auth'
import { parseResponse, readApiError } from './client'
import type { DouDizhuHint, DouDizhuRoom, DouDizhuState, GameMeta } from '../types/doudizhu'

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const { apiBaseUrl } = getAppConfig()
  const session = loadSession()
  const headers: Record<string, string> = {
    ...(init?.headers as Record<string, string>),
  }

  const hasBody = init?.body !== undefined
  if (hasBody) {
    headers['Content-Type'] = 'application/json'
  }

  if (session?.token) {
    headers.Authorization = `Bearer ${session.token}`
  }

  let response: Response
  try {
    response = await fetch(`${apiBaseUrl}${path}`, { ...init, headers })
  } catch {
    throw new Error('无法连接后端，请确认 backend 已启动')
  }

  const data = await parseResponse<T & { error?: string }>(response)
  if (!response.ok) {
    throw new Error(readApiError(data, `请求失败 (${response.status})`))
  }
  return data
}

export function fetchGameCatalog() {
  return apiFetch<{ games: GameMeta[] }>('/api/games/catalog')
}

export function startDouDizhuGame() {
  return apiFetch<DouDizhuState>('/api/games/doudizhu/start', { method: 'POST' })
}

export function getDouDizhuState(gameId: string) {
  return apiFetch<DouDizhuState>(`/api/games/doudizhu/${gameId}`)
}

export function callLandlord(gameId: string, call: boolean) {
  return apiFetch<DouDizhuState>(`/api/games/doudizhu/${gameId}/call`, {
    method: 'POST',
    body: JSON.stringify({ call }),
  })
}

export function playCards(gameId: string, cardIds: string[]) {
  return apiFetch<DouDizhuState>(`/api/games/doudizhu/${gameId}/play`, {
    method: 'POST',
    body: JSON.stringify({ card_ids: cardIds }),
  })
}

export function passTurn(gameId: string) {
  return apiFetch<DouDizhuState>(`/api/games/doudizhu/${gameId}/pass`, {
    method: 'POST',
  })
}

export function fetchDouDizhuHint(gameId: string) {
  return apiFetch<DouDizhuHint>(`/api/games/doudizhu/${gameId}/hint`)
}

export function tickDouDizhuGame(gameId: string) {
  return apiFetch<DouDizhuState>(`/api/games/doudizhu/${gameId}/tick`, { method: 'POST' })
}

export function joinDouDizhuRoom(roomId?: string) {
  return apiFetch<DouDizhuRoom>('/api/games/doudizhu/rooms/join', {
    method: 'POST',
    body: JSON.stringify(roomId ? { room_id: roomId } : {}),
  })
}

export function fetchDouDizhuRoom(roomId: string) {
  return apiFetch<DouDizhuRoom>(`/api/games/doudizhu/rooms/${roomId}`)
}

export function leaveDouDizhuRoom(roomId: string) {
  return apiFetch<DouDizhuRoom | { left: true }>(`/api/games/doudizhu/rooms/${roomId}/leave`, {
    method: 'POST',
  })
}

export function readyDouDizhuRoom(roomId: string, ready: boolean) {
  return apiFetch<DouDizhuRoom>(`/api/games/doudizhu/rooms/${roomId}/ready`, {
    method: 'POST',
    body: JSON.stringify({ ready }),
  })
}

export function nextDouDizhuRoom(roomId: string, gameId: string, ready: boolean) {
  return apiFetch<DouDizhuRoom>(
    `/api/games/doudizhu/rooms/${roomId}/next?game_id=${encodeURIComponent(gameId)}`,
    {
      method: 'POST',
      body: JSON.stringify({ ready }),
    },
  )
}
