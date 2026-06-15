import { useCallback, useState } from 'react'
import { quotationApi } from '../../../entities/quotation/api/quotationApi'
import { usePaginated } from '../../../shared/hooks/usePaginated'
import { useToken } from '../../../app/providers/AuthProvider'
import { DataTable } from '../../../shared/ui/DataTable/DataTable'
import { Pagination } from '../../../shared/ui/Pagination/Pagination'
import { Button } from '../../../shared/ui/Button/Button'
import { QuotationModal } from '../../quotation-modal/ui/QuotationModal'

export function QuotationTable() {
  const token = useToken()
  const [modal, setModal] = useState(null)

  const fetchPage = useCallback((page, limit) => quotationApi.list(token, page, limit), [token])
  const { rows, meta, page, setPage, loading, reload } = usePaginated(fetchPage)

  const columns = [
    { key: 'code', label: 'Mã' },
    { key: 'status', label: 'Trạng thái', render: (r) => <span className={`badge badge--${r.status === 'accepted' ? 'gold' : 'open'}`}>{r.status}</span> },
    { key: 'total', label: 'Tổng', render: (r) => r.total ? `${Number(r.total).toLocaleString('vi-VN')} ₫` : '—' },
    { key: 'valid_until', label: 'Hết hạn', render: (r) => r.valid_until ? r.valid_until.slice(0, 10) : '—' },
    {
      key: '_actions', label: '', render: (r) => (
        <div className="row-actions">
          <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button>
          {r.status === 'draft' && (
            <Button variant="secondary" onClick={async () => { await quotationApi.send(token, r.id); reload() }}>Gửi</Button>
          )}
          {r.status === 'sent' && (
            <Button variant="primary" onClick={async () => { await quotationApi.convert(token, r.id); reload() }}>→ HĐ</Button>
          )}
        </div>
      ),
    },
  ]

  return (
    <>
      <div className="page-header">
        <h1>Báo giá</h1>
        <Button variant="primary" onClick={() => setModal({})}>+ Tạo báo giá</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <QuotationModal open={!!modal} edit={modal?.id ? modal : null} onClose={() => setModal(null)} onSaved={reload} />
    </>
  )
}
