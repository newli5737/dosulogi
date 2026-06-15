import { useEffect, useState } from 'react'
import { listQuotations, paginatedItems } from '../api'
import { DataTable, Page, useToken } from '../components/ui'

export default function QuotationsPage() {
  const token = useToken()
  const [rows, setRows] = useState([])
  useEffect(() => {
    listQuotations(token).then((r) => setRows(paginatedItems(r))).catch(console.error)
  }, [token])

  return (
    <Page title="Báo giá">
      <DataTable columns={[
        { key: 'code', label: 'Mã' },
        { key: 'status', label: 'Trạng thái' },
        { key: 'total', label: 'Tổng', render: (r) => r.total ? `${Number(r.total).toLocaleString('vi-VN')} ₫` : '—' },
      ]} rows={rows} />
    </Page>
  )
}
