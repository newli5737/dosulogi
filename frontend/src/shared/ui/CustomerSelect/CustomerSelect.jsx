import { useEffect, useState } from 'react'
import { customerApi } from '../../../entities/customer/api/customerApi'
import { useToken } from '../../../app/providers/AuthProvider'
import { Select } from '../Form/Form'

export function CustomerSelect({ value, onChange, required }) {
  const token = useToken()
  const [customers, setCustomers] = useState([])

  useEffect(() => {
    if (!token) return
    customerApi.list(token, 1, 200).then((res) => setCustomers(res.data || [])).catch(() => setCustomers([]))
  }, [token])

  return (
    <Select value={value} onChange={(e) => onChange(e.target.value)} required={required}>
      <option value="">— Chọn KH —</option>
      {customers.map((c) => (
        <option key={c.id} value={c.id}>{c.code} — {c.name}</option>
      ))}
    </Select>
  )
}
