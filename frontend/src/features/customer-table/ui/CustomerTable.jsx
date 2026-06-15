import { useCallback, useMemo, useState } from 'react'
import { customerApi } from '../../../entities/customer/api/customerApi'
import { usePaginated } from '../../../shared/hooks/usePaginated'
import { useToken } from '../../../app/providers/AuthProvider'
import { DataTable } from '../../../shared/ui/DataTable/DataTable'
import { Pagination } from '../../../shared/ui/Pagination/Pagination'
import { Button } from '../../../shared/ui/Button/Button'
import { CustomerModal } from '../../customer-modal/ui/CustomerModal'

export function CustomerTable() {
  const token = useToken()
  const [q, setQ] = useState('')
  const [modal, setModal] = useState(null)

  const columns = useMemo(() => [
    { key: 'code', label: 'Mã' },
    { key: 'name', label: 'Tên KH' },
    { key: 'type', label: 'Loại' },
    { key: 'email', label: 'Email' },
    { key: 'tier', label: 'Tier', render: (r) => <span className={`badge badge--${r.tier}`}>{r.tier}</span> },
    { key: 'assigned_to', label: 'Phụ trách', render: (r) => r.assigned_to?.full_name || '—' },
    { key: '_actions', label: '', render: (r) => <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button> },
  ], [])

  const fetchPage = useCallback(
    (page, limit, filters) => customerApi.list(token, page, limit, filters),
    [token],
  )

  const { rows, meta, page, setPage, loading, reload } = usePaginated(fetchPage, { filters: { q } })

  return (
    <>
      <div className="page-header">
        <h1>Khách hàng</h1>
        <Button variant="primary" onClick={() => setModal({})}>+ Thêm KH</Button>
      </div>
      <div style={{ marginBottom: 16 }}>
        <input className="field-input" placeholder="Tìm theo tên, mã, email..." value={q} onChange={(e) => { setQ(e.target.value); setPage(1) }} style={{ maxWidth: 320 }} />
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <CustomerModal open={!!modal} edit={modal?.id ? modal : null} onClose={() => setModal(null)} onSaved={reload} />
    </>
  )
}
