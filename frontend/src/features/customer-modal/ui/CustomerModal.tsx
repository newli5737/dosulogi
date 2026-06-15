import { useEffect, useState, type FormEvent } from 'react'
import { Users } from 'lucide-react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { customerApi } from '@/entities/customer/api/customerApi'
import type { Customer } from '@/entities/customer/model/types'

interface CustomerFormState {
  name: string
  type: 'B2B' | 'B2C'
  email: string
  phone: string
  province: string
  segment: string
  tier: string
  tax_code: string
}

const empty: CustomerFormState = {
  name: '', type: 'B2B', email: '', phone: '', province: '', segment: 'sme', tier: 'standard', tax_code: '',
}

interface CustomerModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
  edit: Customer | null
}

export function CustomerModal({ open, onClose, onSaved, edit }: CustomerModalProps) {
  const [form, setForm] = useState<CustomerFormState>(empty)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (open) {
      setForm(edit ? {
        name: edit.name,
        type: edit.type,
        email: edit.email ?? '',
        phone: edit.phone ?? '',
        province: edit.province ?? '',
        segment: edit.segment,
        tier: edit.tier,
        tax_code: edit.tax_code ?? '',
      } : empty)
    }
  }, [open, edit])

  const set = <K extends keyof CustomerFormState>(k: K, v: CustomerFormState[K]) =>
    setForm((f) => ({ ...f, [k]: v }))

  async function submit(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const body = { ...form, email: form.email || null, phone: form.phone || null }
      if (edit?.id) await customerApi.update(edit.id, body)
      else await customerApi.create(body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={edit ? 'Sửa khách hàng' : 'Thêm khách hàng'} wide icon={Users} tone="blue">
      <form onSubmit={submit}>
        <div className="form-grid">
          <Field label="Tên" required><Input value={form.name} onChange={(e) => set('name', e.target.value)} required /></Field>
          <Field label="Loại" required>
            <Select value={form.type} onChange={(e) => set('type', e.target.value as 'B2B' | 'B2C')}>
              <option value="B2B">B2B</option>
              <option value="B2C">B2C</option>
            </Select>
          </Field>
          <Field label="Email"><Input type="email" value={form.email} onChange={(e) => set('email', e.target.value)} /></Field>
          <Field label="Điện thoại"><Input value={form.phone} onChange={(e) => set('phone', e.target.value)} /></Field>
          <Field label="Tỉnh/TP"><Input value={form.province} onChange={(e) => set('province', e.target.value)} /></Field>
          <Field label="MST"><Input value={form.tax_code} onChange={(e) => set('tax_code', e.target.value)} /></Field>
          <Field label="Segment">
            <Select value={form.segment} onChange={(e) => set('segment', e.target.value)}>
              <option value="enterprise">Enterprise</option>
              <option value="sme">SME</option>
              <option value="individual">Individual</option>
            </Select>
          </Field>
          <Field label="Tier">
            <Select value={form.tier} onChange={(e) => set('tier', e.target.value)}>
              <option value="gold">Gold</option>
              <option value="silver">Silver</option>
              <option value="standard">Standard</option>
            </Select>
          </Field>
        </div>
        {error && <p style={{ color: 'var(--danger)', fontSize: '.875rem' }}>{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu'}</Button>
        </div>
      </form>
    </Modal>
  )
}
