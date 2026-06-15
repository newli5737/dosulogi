import { useEffect, useState } from 'react'
import { listPayments, paginatedItems } from '../api'
import { DataTable, Page, useToken } from '../components/ui'

export default function PaymentsPage() {
  const token = useToken()
  const [rows, setRows] = useState([])
  useEffect(() => {
    listPayments(token).then((r) => setRows(paginatedItems(r))).catch(console.error)
  }, [token])

  return (
    <Page title="Thanh toán">
      <DataTable columns={[
        { key: 'amount', label: 'Số tiền', render: (r) => r.amount ? `${Number(r.amount).toLocaleString('vi-VN')} ₫` : '—' },
        { key: 'method', label: 'Phương thức' },
        { key: 'reference_code', label: 'Mã CK' },
        { key: 'matched_auto', label: 'Auto', render: (r) => r.matched_auto ? '✓' : '—' },
      ]} rows={rows} />
    </Page>
  )
}
