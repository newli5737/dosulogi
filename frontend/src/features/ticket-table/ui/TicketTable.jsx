import { useCallback, useState } from 'react'
import { ticketApi } from '../../../entities/ticket/api/ticketApi'
import { usePaginated } from '../../../shared/hooks/usePaginated'
import { useToken } from '../../../app/providers/AuthProvider'
import { DataTable } from '../../../shared/ui/DataTable/DataTable'
import { Pagination } from '../../../shared/ui/Pagination/Pagination'
import { Button } from '../../../shared/ui/Button/Button'
import { TicketModal } from '../../ticket-modal/ui/TicketModal'

const columns = [
  { key: 'code', label: 'Mã' },
  { key: 'title', label: 'Tiêu đề' },
  { key: 'priority', label: 'Ưu tiên', render: (r) => <span className={`badge badge--${r.priority === 'urgent' ? 'urgent' : 'open'}`}>{r.priority}</span> },
  { key: 'status', label: 'Trạng thái' },
  { key: 'customer', label: 'KH', render: (r) => r.customer?.name || '—' },
  { key: 'is_overdue', label: 'SLA', render: (r) => r.is_overdue ? '⚠ Quá hạn' : 'OK' },
]

export function TicketTable() {
  const token = useToken()
  const [open, setOpen] = useState(false)
  const fetchPage = useCallback((page, limit) => ticketApi.list(token, page, limit), [token])
  const { rows, meta, page, setPage, loading, reload } = usePaginated(fetchPage)

  return (
    <>
      <div className="page-header">
        <h1>Tickets hỗ trợ</h1>
        <Button variant="primary" onClick={() => setOpen(true)}>+ Tạo ticket</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <TicketModal open={open} onClose={() => setOpen(false)} onSaved={reload} />
    </>
  )
}
