import { useCallback, useMemo, useState } from 'react'
import { invoiceApi } from '@/entities/invoice/api/invoiceApi'
import type { Invoice } from '@/entities/invoice/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { InvoiceModal } from '@/features/invoice-modal/ui/InvoiceModal'

async function downloadInvoicePdf(invoice: Invoice) {
  const blob = await invoiceApi.download(invoice.id)
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `invoice-${invoice.code}.pdf`
  a.click()
  URL.revokeObjectURL(url)
}

export function InvoiceTable() {
  const [modal, setModal] = useState<Invoice | Record<string, never> | null>(null)

  const fetchPage = useCallback(
    (page: number, limit: number) => invoiceApi.list(page, limit),
    [],
  )
  const { rows, meta, page, setPage, loading, reload } = usePaginated<Invoice>(fetchPage)

  const columns = useMemo<DataTableColumn<Invoice>[]>(() => [
    { key: 'code', label: 'Mã' },
    { key: 'status', label: 'Trạng thái', render: (r) => <span className={`badge badge--${r.status === 'paid' ? 'gold' : 'open'}`}>{r.status}</span> },
    { key: 'total', label: 'Tổng', render: (r) => r.total ? `${Number(r.total).toLocaleString('vi-VN')} ₫` : '—' },
    { key: 'due_date', label: 'Hạn TT', render: (r) => r.due_date ? r.due_date.slice(0, 10) : '—' },
    {
      key: '_actions', label: '', render: (r) => (
        <div className="row-actions">
          <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button>
          {r.status === 'draft' && (
            <Button variant="secondary" onClick={async () => { await invoiceApi.send(r.id); reload() }}>Gửi</Button>
          )}
          {r.status !== 'cancelled' && (
            <Button variant="secondary" onClick={() => downloadInvoicePdf(r)}>PDF</Button>
          )}
          {r.status !== 'cancelled' && r.status !== 'paid' && (
            <Button variant="secondary" onClick={async () => { await invoiceApi.cancel(r.id); reload() }}>Hủy</Button>
          )}
        </div>
      ),
    },
  ], [reload])

  return (
    <>
      <div className="page-header">
        <h1>Hóa đơn</h1>
        <Button variant="primary" onClick={() => setModal({})}>+ Tạo hóa đơn</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <InvoiceModal
        open={modal !== null}
        edit={modal && 'id' in modal && modal.id ? (modal as Invoice) : null}
        onClose={() => setModal(null)}
        onSaved={reload}
      />
    </>
  )
}
