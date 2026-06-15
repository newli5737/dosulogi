import { useCallback, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { customerApi } from '@/entities/customer/api/customerApi'
import type { Customer } from '@/entities/customer/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { CustomerModal } from '@/features/customer-modal/ui/CustomerModal'

type CustomerFilters = { q: string }

export function CustomerTable() {
  const [q, setQ] = useState('')
  const [modal, setModal] = useState<Customer | Record<string, never> | null>(null)

  const columns = useMemo<DataTableColumn<Customer>[]>(() => [
    { key: 'code', label: 'Mã' },
    { key: 'name', label: 'Tên KH' },
    { key: 'type', label: 'Loại' },
    { key: 'email', label: 'Email' },
    { key: 'tier', label: 'Tier', render: (r) => <span className={`badge badge--${r.tier}`}>{r.tier}</span> },
    { key: 'assigned_to', label: 'Phụ trách', render: (r) => r.assigned_to?.full_name || '—' },
    {
      key: '_actions', label: '', render: (r) => (
        <div className="row-actions">
          <Link to={`/customers/${r.id}`}><Button variant="secondary">360°</Button></Link>
          <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button>
        </div>
      ),
    },
  ], [])

  const fetchPage = useCallback(
    (page: number, limit: number, filters: CustomerFilters) => customerApi.list(page, limit, filters),
    [],
  )

  const { rows, meta, page, setPage, loading, reload } = usePaginated<Customer, CustomerFilters>(
    fetchPage,
    { filters: { q } },
  )

  return (
    <>
      <div className="page-header">
        <h1>Khách hàng</h1>
        <Button variant="primary" onClick={() => setModal({})}>+ Thêm KH</Button>
      </div>
      <div style={{ marginBottom: 16 }}>
        <input
          className="field-input"
          placeholder="Tìm theo tên, mã, email..."
          value={q}
          onChange={(e) => { setQ(e.target.value); setPage(1) }}
          style={{ maxWidth: 320 }}
        />
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <CustomerModal
        open={modal !== null}
        edit={modal && 'id' in modal && modal.id ? (modal as Customer) : null}
        onClose={() => setModal(null)}
        onSaved={reload}
      />
    </>
  )
}
