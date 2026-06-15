import { useEffect, useState } from 'react'
import { listOpportunities, paginatedItems } from '../api'
import { DataTable, Page, useToken } from '../components/ui'

export default function OpportunitiesPage() {
  const token = useToken()
  const [rows, setRows] = useState([])
  useEffect(() => {
    listOpportunities(token).then((r) => setRows(paginatedItems(r))).catch(console.error)
  }, [token])

  return (
    <Page title="Cơ hội bán hàng">
      <DataTable columns={[
        { key: 'title', label: 'Tiêu đề' },
        { key: 'stage', label: 'Stage' },
        { key: 'value', label: 'Giá trị', render: (r) => r.value ? `${Number(r.value).toLocaleString('vi-VN')} ₫` : '—' },
        { key: 'currency', label: 'Tiền tệ' },
      ]} rows={rows} />
    </Page>
  )
}
