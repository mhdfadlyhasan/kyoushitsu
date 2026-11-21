import type { Role } from './roles'

export type { Role }

const TOKEN_KEY = 'study-companion.token'

export type User = {
  id: number
  name: string
  email: string
  role: Role
  created_at: string
}

export function token(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function storeToken(value: string) {
  localStorage.setItem(TOKEN_KEY, value)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')
  if (options.body !== undefined) {
    headers.set('Content-Type', 'application/json')
  }

  const current = token()
  if (current) {
    headers.set('Authorization', `Bearer ${current}`)
  }

  const response = await fetch(path, { ...options, headers })
  const payload: unknown = await response.json().catch(() => null)

  if (!response.ok) {
    const message =
      payload !== null &&
      typeof payload === 'object' &&
      'error' in payload
        ? String(payload.error)
        : `Request failed (${response.status})`
    throw Object.assign(new Error(message), { status: response.status })
  }

  return payload as T
}
