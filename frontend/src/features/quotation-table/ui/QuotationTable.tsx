import { useCallback, useMemo, useState } from 'react'
import { quotationApi } from '@/entities/quotation/api/quotationApi'
import type { Quotation } from '@/entities/quotation/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { quotationStatusLabel } from '@/shared/lib/labels'
import { QuotationModal } from '@/features/quotation-modal/ui/QuotationModal'

export function QuotationTable() {
  const [modal, setModal] = useState<Quotation | Record<string, never> | null>(null)

  const fetchPage = useCallback(
    (page: number, limit: number) => quotationApi.list(page, limit),
    [],
  )
  const { rows, meta, page, setPage, loading, reload } = usePaginated<Quotation>(fetchPage)

  const columns = useMemo<DataTableColumn<Quotation>[]>(() => [
    { key: 'code', label: 'Mã' },
    { key: 'status', label: 'Trạng thái', render: (r) => <span className={`badge badge--${r.status === 'accepted' ? 'gold' : 'open'}`}>{quotationStatusLabel(r.status)}</span> },
    { key: 'total', label: 'Tổng', render: (r) => r.total ? `${Number(r.total).toLocaleString('vi-VN')} ₫` : '—' },
    { key: 'valid_until', label: 'Hết hạn', render: (r) => r.valid_until ? r.valid_until.slice(0, 10) : '—' },
    {
      key: '_actions', label: '', render: (r) => (
        <div className="row-actions">
          <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button>
          {r.status === 'draft' && (
            <Button variant="secondary" onClick={async () => { await quotationApi.send(r.id); reload() }}>Gửi</Button>
          )}
          {r.status === 'sent' && (
            <Button variant="primary" onClick={async () => { await quotationApi.convert(r.id); reload() }}>→ HĐ</Button>
          )}
        </div>
      ),
    },
  ], [reload])

  return (
    <>
      <div className="page-header">
        <h1>Báo giá</h1>
        <Button variant="primary" onClick={() => setModal({})}>+ Tạo báo giá</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <QuotationModal
        open={modal !== null}
        edit={modal && 'id' in modal && modal.id ? (modal as Quotation) : null}
        onClose={() => setModal(null)}
        onSaved={reload}
      />
    </>
  )
}
