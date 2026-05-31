import { getAppConfig } from '../config/loadConfig'
import { parseResponse, readApiError } from './client'
import type { AuthSession, LoginResponse, User } from '../types/auth'

const STORAGE_KEY = 'card.auth.session'
const REMEMBER_KEY = 'card.auth.remember'

export interface RememberUsername {
  username: string
  enabled: boolean
}

export async function loginByUsername(username: string): Promise<AuthSession> {
  const { apiBaseUrl } = getAppConfig()
  let response: Response
  try {
    response = await fetch(`${apiBaseUrl}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username }),
    })
  } catch {
    throw new Error('无法连接后端，请确认 backend 已启动')
  }

  const data = await parseResponse<LoginResponse & { error?: string }>(response)
  if (!response.ok) {
    throw new Error(readApiError(data, 'login failed'))
  }

  const payload = data
  const session: AuthSession = {
    token: payload.token,
    expiresAt: payload.expires_at,
    user: payload.user,
  }
  saveSession(session)
  return session
}

export async function fetchCurrentUser(token: string): Promise<User> {
  const { apiBaseUrl } = getAppConfig()
  let response: Response
  try {
    response = await fetch(`${apiBaseUrl}/api/auth/me`, {
      headers: { Authorization: `Bearer ${token}` },
    })
  } catch {
    throw new Error('无法连接后端，请确认 backend 已启动')
  }

  const data = await parseResponse<{ user: User; error?: string }>(response)
  if (!response.ok) {
    throw new Error(readApiError(data, 'session expired'))
  }

  return data.user
}

export function loadSession(): AuthSession | null {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as AuthSession
  } catch {
    localStorage.removeItem(STORAGE_KEY)
    return null
  }
}

export function saveSession(session: AuthSession) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
}

export function clearSession() {
  localStorage.removeItem(STORAGE_KEY)
}

export function loadRememberUsername(): RememberUsername | null {
  const raw = localStorage.getItem(REMEMBER_KEY)
  if (!raw) {
    return null
  }

  try {
    const data = JSON.parse(raw) as RememberUsername
    if (!data.enabled || !data.username) {
      return null
    }
    return data
  } catch {
    localStorage.removeItem(REMEMBER_KEY)
    return null
  }
}

export function saveRememberUsername(username: string) {
  const payload: RememberUsername = {
    username,
    enabled: true,
  }
  localStorage.setItem(REMEMBER_KEY, JSON.stringify(payload))
}

export function clearRememberUsername() {
  localStorage.removeItem(REMEMBER_KEY)
}

export function getStoredUsername(): string {
  return loadRememberUsername()?.username ?? loadSession()?.user.username ?? ''
}
