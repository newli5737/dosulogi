import { useEffect, useState } from 'react'
import { listUsers, paginatedItems } from '../api'
import { DataTable, Page, useToken } from '../components/ui'

export default function UsersPage() {
  const token = useToken()
  const [rows, setRows] = useState([])
  useEffect(() => {
    listUsers(token).then((r) => setRows(paginatedItems(r))).catch(console.error)
  }, [token])

  return (
    <Page title="Quản lý users">
      <DataTable columns={[
        { key: 'email', label: 'Email' },
        { key: 'full_name', label: 'Họ tên' },
        { key: 'role', label: 'Role' },
        { key: 'is_active', label: 'Active', render: (r) => r.is_active ? '✓' : '✗' },
      ]} rows={rows} />
    </Page>
  )
}
