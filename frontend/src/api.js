const API_URL = import.meta.env.VITE_API_URL || ''

function headers(token) {
  const h = { 'Content-Type': 'application/json' }
  if (token) h.Authorization = `Bearer ${token}`
  return h
}

export async function login(email, password) {
  const res = await fetch(`${API_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: headers(),
    credentials: 'include',
    body: JSON.stringify({ email, password }),
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'Login failed')
  return data
}

export async function getSummary(token) {
  const res = await fetch(`${API_URL}/api/v1/dashboard/summary`, {
    headers: headers(token),
    credentials: 'include',
  })
  if (!res.ok) throw new Error('Failed to load dashboard')
  return res.json()
}

export async function getMe(token) {
  const res = await fetch(`${API_URL}/api/v1/auth/me`, {
    headers: headers(token),
    credentials: 'include',
  })
  if (!res.ok) throw new Error('Unauthorized')
  return res.json()
}
