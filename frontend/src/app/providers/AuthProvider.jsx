import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { authApi } from '../../entities/session/api/sessionApi'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [session, setSession] = useState(null)
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (!token) { setChecking(false); return }
    authApi.me(token)
      .then((user) => setSession({ user, token }))
      .catch(() => localStorage.removeItem('access_token'))
      .finally(() => setChecking(false))
  }, [])

  const value = useMemo(() => ({
    session,
    checking,
    login: async (email, password) => {
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

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth outside provider')
  return ctx
}

export function useToken() {
  return useAuth().session?.token
}
