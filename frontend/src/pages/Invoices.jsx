import { useEffect, useState } from 'react'
import { listInvoices, paginatedItems } from '../api'
import { DataTable, Page, useToken } from '../components/ui'

export default function InvoicesPage() {
  const token = useToken()
  const [rows, setRows] = useState([])
  useEffect(() => {
    listInvoices(token).then((r) => setRows(paginatedItems(r))).catch(console.error)
  }, [token])

  return (
    <Page title="Hóa đơn">
      <DataTable columns={[
        { key: 'code', label: 'Mã' },
        { key: 'status', label: 'Trạng thái' },
        { key: 'total', label: 'Tổng', render: (r) => r.total ? `${Number(r.total).toLocaleString('vi-VN')} ₫` : '—' },
        { key: 'due_date', label: 'Hạn TT' },
      ]} rows={rows} />
    </Page>
  )
}
