import { useEffect, useState } from 'react'
import { listContracts, paginatedItems } from '../api'
import { DataTable, Page, useToken } from '../components/ui'

export default function ContractsPage() {
  const token = useToken()
  const [rows, setRows] = useState([])
  useEffect(() => {
    listContracts(token).then((r) => setRows(paginatedItems(r))).catch(console.error)
  }, [token])

  return (
    <Page title="Hợp đồng">
      <DataTable columns={[
        { key: 'code', label: 'Mã' },
        { key: 'title', label: 'Tiêu đề' },
        { key: 'status', label: 'Trạng thái' },
        { key: 'service_type', label: 'Dịch vụ' },
      ]} rows={rows} />
    </Page>
  )
}
