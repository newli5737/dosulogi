import { useCallback, useMemo, useState } from 'react'
import { userApi } from '@/entities/user/api/userApi'
import type { User } from '@/entities/user/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { useAuth, useToken } from '@/app/providers/AuthProvider'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { UserModal } from '@/features/user-modal/ui/UserModal'

export function UserTable() {
  const token = useToken()
  const { session } = useAuth()
  const [open, setOpen] = useState(false)

  const isAdmin = session?.user?.role === 'admin'

  const columns = useMemo<DataTableColumn<User>[]>(() => [
    { key: 'email', label: 'Email' },
    { key: 'full_name', label: 'Họ tên' },
    { key: 'role', label: 'Role' },
    { key: 'is_active', label: 'Active', render: (r) => r.is_active ? '✓' : '✗' },
  ], [])

  const fetchPage = useCallback(
    (page: number, limit: number) => userApi.list(token!, page, limit),
    [token],
  )
  const { rows, meta, page, setPage, loading, reload } = usePaginated<User>(fetchPage, { enabled: isAdmin })

  if (!isAdmin) return <div className="page-card"><p>Chỉ admin mới truy cập được.</p></div>

  return (
    <>
      <div className="page-header">
        <h1>Quản lý users</h1>
        <Button variant="primary" onClick={() => setOpen(true)}>+ Thêm user</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <UserModal open={open} onClose={() => setOpen(false)} onSaved={reload} />
    </>
  )
}
