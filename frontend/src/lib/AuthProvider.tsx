import { useCallback, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { apiRequest, clearToken, storeToken, token } from './api'
import type { User } from './api'
import { AuthContext } from './auth'

type LoginResponse = {
  token: string
  user: User
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(() => token() !== null)

  useEffect(() => {
    if (token() === null) return

    let active = true
    apiRequest<User>('/api/me')
      .then((me) => {
        if (active) setUser(me)
      })
      .catch((error: unknown) => {
        if ((error as { status?: number }).status === 401) {
          clearToken()
        }
      })
      .finally(() => {
        if (active) setLoading(false)
      })

    return () => {
      active = false
    }
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const result = await apiRequest<LoginResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
    storeToken(result.token)
    setUser(result.user)
  }, [])

  const logout = useCallback(() => {
    clearToken()
    setUser(null)
  }, [])

  const value = useMemo(
    () => ({ user, loading, login, logout }),
    [user, loading, login, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
