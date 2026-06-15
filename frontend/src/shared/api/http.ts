import type { ApiErrorResponse, HttpMethod, HttpOptions } from './types'
import { getErrorMessage } from './types'

const BASE = import.meta.env.VITE_API_URL ?? ''

export async function http<T>(path: string, options: HttpOptions = {}): Promise<T> {
  const { token, method = 'GET', body } = options
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, {
    method: method as HttpMethod,
    headers,
    credentials: 'include',
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 204) return undefined as T

  const json: unknown = await res.json().catch(() => ({}))
  if (!res.ok) {
    const msg = getErrorMessage(json as ApiErrorResponse, res.statusText)
    throw new Error(msg)
  }
  return json as T
}

export async function httpBlob(path: string, token: string): Promise<Blob> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { Authorization: `Bearer ${token}` },
    credentials: 'include',
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.blob()
}

export async function httpForm<T>(path: string, token: string, formData: FormData, method: HttpMethod = 'POST'): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers: { Authorization: `Bearer ${token}` },
    credentials: 'include',
    body: formData,
  })
  const json: unknown = await res.json().catch(() => ({}))
  if (!res.ok) {
    const msg = getErrorMessage(json as ApiErrorResponse, res.statusText)
    throw new Error(msg)
  }
  return json as T
}

export function listParams(page: number, limit: number, extra: Record<string, string | undefined> = {}): string {
  const q = new URLSearchParams({ page: String(page), limit: String(limit) })
  for (const [k, v] of Object.entries(extra)) {
    if (v) q.set(k, v)
  }
  return q.toString()
}
