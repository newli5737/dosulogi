import { useEffect, useState } from 'react'
import { createCustomer, listCustomers, paginatedItems } from '../api'
import { DataTable, Page, useToken } from '../components/ui'

export default function CustomersPage() {
  const token = useToken()
  const [rows, setRows] = useState([])
  const [form, setForm] = useState({ name: '', type: 'B2B', email: '', phone: '' })

  const load = () => listCustomers(token).then((r) => setRows(paginatedItems(r)))

  useEffect(() => { load().catch(console.error) }, [token])

  async function submit(e) {
    e.preventDefault()
    await createCustomer(token, form)
    setForm({ name: '', type: 'B2B', email: '', phone: '' })
    load()
  }

  return (
    <Page title="Khách hàng (CRM)">
      <form className="inline-form" onSubmit={submit}>
        <input placeholder="Tên KH" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
        <select value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })}>
          <option value="B2B">B2B</option>
          <option value="B2C">B2C</option>
        </select>
        <input placeholder="Email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
        <input placeholder="Phone" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} />
        <button type="submit">Thêm</button>
      </form>
      <DataTable
        columns={[
          { key: 'code', label: 'Mã' },
          { key: 'name', label: 'Tên' },
          { key: 'type', label: 'Loại' },
          { key: 'email', label: 'Email' },
          { key: 'tier', label: 'Tier' },
        ]}
        rows={rows}
      />
    </Page>
  )
}
