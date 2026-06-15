import type { DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import type { PaginatedResponse } from '@/shared/api/types'
import { useCallback } from 'react'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { useToken } from '@/app/providers/AuthProvider'
import { DataTable } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'

interface CreateListPageOptions<T extends { id?: string }> {
  title: string
  fetchList: (token: string | undefined, page: number, limit: number) => Promise<PaginatedResponse<T>>
  columns: DataTableColumn<T>[]
}

export function createListPage<T extends { id?: string }>({ title, fetchList, columns }: CreateListPageOptions<T>) {
  return function ListPage() {
    const token = useToken()
    const fetchPage = useCallback(
      (page: number, limit: number) => fetchList(token, page, limit),
      [token],
    )
    const { rows, meta, page, setPage, loading } = usePaginated<T>(fetchPage)

    return (
      <div className="page-card">
        <div className="page-header"><h1>{title}</h1></div>
        <DataTable columns={columns} rows={rows} loading={loading} />
        <Pagination page={page} limit={meta.limit || 20} total={meta.total} onChange={setPage} />
      </div>
    )
  }
}
