import { useState, type FormEvent } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { userApi } from '@/entities/user/api/userApi'
import { useToken } from '@/app/providers/AuthProvider'

interface UserFormState {
  email: string
  password: string
  full_name: string
  role: string
}

const empty: UserFormState = { email: '', password: '', full_name: '', role: 'sales_rep' }

interface UserModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
}

export function UserModal({ open, onClose, onSaved }: UserModalProps) {
  const token = useToken()
  const [form, setForm] = useState<UserFormState>(empty)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const set = <K extends keyof UserFormState>(k: K, v: UserFormState[K]) =>
    setForm((f) => ({ ...f, [k]: v }))

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    setLoading(true)
    setError('')
    try {
      await userApi.create(token, form)
      onSaved?.()
      onClose()
      setForm(empty)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Thêm user">
      <form onSubmit={submit}>
        <Field label="Email" required><Input type="email" value={form.email} onChange={(e) => set('email', e.target.value)} required /></Field>
        <Field label="Mật khẩu" required><Input type="password" value={form.password} onChange={(e) => set('password', e.target.value)} required minLength={8} /></Field>
        <Field label="Họ tên" required><Input value={form.full_name} onChange={(e) => set('full_name', e.target.value)} required /></Field>
        <Field label="Role">
          <Select value={form.role} onChange={(e) => set('role', e.target.value)}>
            <option value="admin">Admin</option>
            <option value="sales_rep">Sales</option>
            <option value="accountant">Accountant</option>
            <option value="director">Director</option>
          </Select>
        </Field>
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang tạo...' : 'Tạo'}</Button>
        </div>
      </form>
    </Modal>
  )
}
