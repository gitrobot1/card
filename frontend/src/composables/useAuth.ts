import { computed, ref } from 'vue'
import {
  clearSession,
  fetchCurrentUser,
  loadSession,
  loginByUsername,
  saveSession,
} from '../api/auth'
import type { AuthSession, User } from '../types/auth'

const session = ref<AuthSession | null>(loadSession())
const bootstrapping = ref(true)

export function useAuth() {
  const isLoggedIn = computed(() => session.value !== null)
  const user = computed(() => session.value?.user ?? null)

  async function bootstrap() {
    bootstrapping.value = true
    const stored = loadSession()
    if (!stored) {
      session.value = null
      bootstrapping.value = false
      return
    }

    try {
      const currentUser = await fetchCurrentUser(stored.token)
      session.value = {
        token: stored.token,
        expiresAt: stored.expiresAt,
        user: currentUser,
      }
      saveSession(session.value)
    } catch {
      clearSession()
      session.value = null
    } finally {
      bootstrapping.value = false
    }
  }

  async function login(username: string) {
    const nextSession = await loginByUsername(username)
    session.value = nextSession
    return nextSession
  }

  function logout() {
    clearSession()
    session.value = null
  }

  return {
    session,
    user,
    isLoggedIn,
    bootstrapping,
    bootstrap,
    login,
    logout,
  }
}

export type AuthUser = User
