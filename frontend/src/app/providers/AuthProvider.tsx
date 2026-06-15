import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { authApi } from '@/entities/session/api/sessionApi'
import type { UserBrief } from '@/shared/api/types'

interface AuthContextValue {
  session: UserBrief | null
  checking: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<UserBrief | null>(null)
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    authApi.me()
      .then((user) => setSession(user))
      .catch(() => setSession(null))
      .finally(() => setChecking(false))
  }, [])

  const value = useMemo<AuthContextValue>(() => ({
    session,
    checking,
    login: async (email: string, password: string) => {
      const res = await authApi.login({ email, password })
      setSession(res.user)
    },
    logout: async () => {
      try { await authApi.logout() } catch { /* ignore */ }
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

/** @deprecated Auth uses httpOnly cookies — use useAuth().session instead */
export function useToken(): undefined {
  return undefined
}
