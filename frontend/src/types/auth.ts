export interface User {
  id: number
  username: string
  nickname: string
  last_login: string
  created_at: string
  updated_at: string
}

export interface LoginResponse {
  token: string
  expires_at: string
  user: User
  is_new_user: boolean
}

export interface AuthSession {
  token: string
  expiresAt: string
  user: User
}
