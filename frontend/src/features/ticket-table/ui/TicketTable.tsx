import { useCallback, useState } from 'react'
import { ticketApi } from '@/entities/ticket/api/ticketApi'
import type { Ticket } from '@/entities/ticket/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { useToken } from '@/app/providers/AuthProvider'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { TicketModal } from '@/features/ticket-modal/ui/TicketModal'
import { TicketDetailModal } from '@/features/ticket-detail-modal/ui/TicketDetailModal'

const columns: DataTableColumn<Ticket>[] = [
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
  const [detailId, setDetailId] = useState<string | null>(null)

  const fetchPage = useCallback(
    (page: number, limit: number) => ticketApi.list(token!, page, limit),
    [token],
  )
  const { rows, meta, page, setPage, loading, reload } = usePaginated<Ticket>(fetchPage)

  const tableColumns: DataTableColumn<Ticket>[] = [
    ...columns,
    {
      key: '_actions', label: '', render: (r) => (
        <Button variant="secondary" onClick={() => setDetailId(r.id)}>Chi tiết</Button>
      ),
    },
  ]

  return (
    <>
      <div className="page-header">
        <h1>Tickets hỗ trợ</h1>
        <Button variant="primary" onClick={() => setOpen(true)}>+ Tạo ticket</Button>
      </div>
      <DataTable columns={tableColumns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <TicketModal open={open} onClose={() => setOpen(false)} onSaved={reload} />
      <TicketDetailModal
        open={detailId !== null}
        ticketId={detailId}
        onClose={() => setDetailId(null)}
        onSaved={reload}
      />
    </>
  )
}
