import type { ApiErrorResponse, HttpMethod, HttpOptions } from './types'
import { getErrorMessage } from './types'

const BASE = import.meta.env.VITE_API_URL ?? ''

let refreshPromise: Promise<boolean> | null = null

async function tryRefresh(): Promise<boolean> {
  try {
    const res = await fetch(`${BASE}/api/v1/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    })
    return res.ok
  } catch {
    return false
  }
}

async function request(path: string, options: HttpOptions = {}, retried = false): Promise<Response> {
  const { method = 'GET', body } = options
  const headers: Record<string, string> = {}
  if (body !== undefined) headers['Content-Type'] = 'application/json'

  const res = await fetch(`${BASE}${path}`, {
    method: method as HttpMethod,
    headers,
    credentials: 'include',
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 401 && !retried && !path.includes('/auth/login') && !path.includes('/auth/refresh')) {
    if (!refreshPromise) {
      refreshPromise = tryRefresh().finally(() => { refreshPromise = null })
    }
    if (await refreshPromise) {
      return request(path, options, true)
    }
  }
  // Do not retry on 429 — avoids amplifying rate-limit storms.
  return res
}

export async function http<T>(path: string, options: HttpOptions = {}): Promise<T> {
  const res = await request(path, options)

  if (res.status === 204) return undefined as T

  const json: unknown = await res.json().catch(() => ({}))
  if (!res.ok) {
    const msg = getErrorMessage(json as ApiErrorResponse, res.statusText)
    throw new Error(msg)
  }
  return json as T
}

export async function httpBlob(path: string): Promise<Blob> {
  const res = await request(path)
  if (!res.ok) throw new Error(res.statusText)
  return res.blob()
}

export async function httpForm<T>(path: string, formData: FormData, method: HttpMethod = 'POST'): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method,
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
