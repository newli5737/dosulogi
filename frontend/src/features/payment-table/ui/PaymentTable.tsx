import { useCallback, useMemo, useState } from 'react'
import { paymentApi } from '@/entities/payment/api/paymentApi'
import type { Payment } from '@/entities/payment/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { PaymentModal } from '@/features/payment-modal/ui/PaymentModal'

export function PaymentTable() {
  const [open, setOpen] = useState(false)

  const fetchPage = useCallback(
    (page: number, limit: number) => paymentApi.list(page, limit),
    [],
  )
  const { rows, meta, page, setPage, loading, reload } = usePaginated<Payment>(fetchPage)

  const columns = useMemo<DataTableColumn<Payment>[]>(() => [
    { key: 'amount', label: 'Số tiền', render: (r) => r.amount ? `${Number(r.amount).toLocaleString('vi-VN')} ₫` : '—' },
    { key: 'method', label: 'Phương thức', render: (r) => r.method || '—' },
    { key: 'reference_code', label: 'Mã CK', render: (r) => r.reference_code || '—' },
    { key: 'created_at', label: 'Ngày tạo', render: (r) => r.created_at ? r.created_at.slice(0, 10) : '—' },
  ], [])

  return (
    <>
      <div className="page-header">
        <h1>Thanh toán</h1>
        <Button variant="primary" onClick={() => setOpen(true)}>+ Ghi nhận thanh toán</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <PaymentModal open={open} onClose={() => setOpen(false)} onSaved={reload} />
    </>
  )
}
