import { useEffect, useState } from 'react'
import { customerApi } from '@/entities/customer/api/customerApi'
import { Select } from '../Form/Form'

interface CustomerSelectProps {
  value: string
  onChange: (value: string) => void
  required?: boolean
}

export function CustomerSelect({ value, onChange, required }: CustomerSelectProps) {
  const [customers, setCustomers] = useState<{ id: string; code: string; name: string }[]>([])

  useEffect(() => {
    void customerApi.list(1, 200).then((res) => setCustomers(res.data)).catch(() => setCustomers([]))
  }, [])

  return (
    <Select value={value} onChange={(e) => onChange(e.target.value)} required={required}>
      <option value="">— Chọn KH —</option>
      {customers.map((c) => (
        <option key={c.id} value={c.id}>{c.code} — {c.name}</option>
      ))}
    </Select>
  )
}
