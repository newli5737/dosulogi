import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { authApi } from '@/entities/session/api/sessionApi'
import type { UserBrief } from '@/shared/api/types'

export interface Session {
  user: UserBrief
  token: string
}

interface AuthContextValue {
  session: Session | null
  checking: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(null)
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (!token) { setChecking(false); return }
    authApi.me(token)
      .then((user) => setSession({ user, token }))
      .catch(() => localStorage.removeItem('access_token'))
      .finally(() => setChecking(false))
  }, [])

  const value = useMemo<AuthContextValue>(() => ({
    session,
    checking,
    login: async (email: string, password: string) => {
      const res = await authApi.login({ email, password })
      localStorage.setItem('access_token', res.access_token)
      setSession({ user: res.user, token: res.access_token })
    },
    logout: () => {
      localStorage.removeItem('access_token')
      setSession(null)
    },
  }), [session, checking])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth outside provider')
  return ctx
}

export function useToken(): string | undefined {
  return useAuth().session?.token
}
