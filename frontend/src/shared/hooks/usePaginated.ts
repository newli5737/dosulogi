import { useCallback, useEffect, useState } from 'react'
import type { ApiMeta, PaginatedResponse } from '../api/types'
import { parseMeta } from '../api/types'

interface UsePaginatedOptions<F extends Record<string, string | undefined>> {
  limit?: number
  filters?: F
  enabled?: boolean
}

export function usePaginated<T, F extends Record<string, string | undefined> = Record<string, never>>(
  fetchPage: (page: number, limit: number, filters: F) => Promise<PaginatedResponse<T>>,
  { limit = 20, filters = {} as F, enabled = true }: UsePaginatedOptions<F> = {},
) {
  const [page, setPage] = useState(1)
  const [rows, setRows] = useState<T[]>([])
  const [meta, setMeta] = useState<ApiMeta>({ page: 1, limit, total: 0 })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const reload = useCallback(async () => {
    if (!enabled) return
    setLoading(true)
    setError('')
    try {
      const res = await fetchPage(page, limit, filters)
      setRows(Array.isArray(res.data) ? res.data : [])
      setMeta(parseMeta(res, page, limit))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Unknown error')
      setRows([])
    } finally {
      setLoading(false)
    }
  }, [fetchPage, page, limit, filters, enabled])

  useEffect(() => { void reload() }, [reload])

  return { rows, meta, page, setPage, loading, error, reload }
}
