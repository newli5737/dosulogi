import { useCallback } from 'react'
import { usePaginated } from '../../shared/hooks/usePaginated'
import { useToken } from '../../app/providers/AuthProvider'
import { DataTable } from '../../shared/ui/DataTable/DataTable'
import { Pagination } from '../../shared/ui/Pagination/Pagination'

export function createListPage({ title, fetchList, columns }) {
  return function ListPage() {
    const token = useToken()
    const fetchPage = useCallback((page, limit) => fetchList(token, page, limit), [token])
    const { rows, meta, page, setPage, loading } = usePaginated(fetchPage)

    return (
      <div className="page-card">
        <div className="page-header"><h1>{title}</h1></div>
        <DataTable columns={columns} rows={rows} loading={loading} />
        <Pagination page={page} limit={meta.limit || 20} total={meta.total} onChange={setPage} />
      </div>
    )
  }
}
