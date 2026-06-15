const BASE = import.meta.env.VITE_API_URL || ''

export async function http(path, { token, method = 'GET', body } = {}) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    credentials: 'include',
    body: body ? JSON.stringify(body) : undefined,
  })
  const json = await res.json().catch(() => ({}))
  if (!res.ok) {
    const msg = typeof json.error === 'string' ? json.error : json.error?.message || res.statusText
    throw new Error(msg)
  }
  return json
}

export function listParams(page, limit, extra = {}) {
  const q = new URLSearchParams({ page: String(page), limit: String(limit) })
  Object.entries(extra).forEach(([k, v]) => { if (v) q.set(k, v) })
  return q.toString()
}
