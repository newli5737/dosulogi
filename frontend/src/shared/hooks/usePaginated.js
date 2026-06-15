import { useCallback, useEffect, useState } from 'react'

export function usePaginated(fetchPage, { limit = 20, filters = {}, enabled = true } = {}) {
  const [page, setPage] = useState(1)
  const [rows, setRows] = useState([])
  const [meta, setMeta] = useState({ page: 1, limit, total: 0 })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const reload = useCallback(async () => {
    if (!enabled) return
    setLoading(true)
    setError('')
    try {
      const res = await fetchPage(page, limit, filters)
      setRows(Array.isArray(res.data) ? res.data : [])
      setMeta(res.meta || res.pagination || { page, limit, total: 0 })
    } catch (e) {
      setError(e.message)
      setRows([])
    } finally {
      setLoading(false)
    }
  }, [fetchPage, page, limit, filters, enabled])

  useEffect(() => { reload() }, [reload])

  return { rows, meta, page, setPage, loading, error, reload }
}
