import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { userApi } from '@/entities/user/api/userApi'
import type { User } from '@/entities/user/model/types'
import { useToken } from '@/app/providers/AuthProvider'

interface UserFormState {
  email: string
  password: string
  full_name: string
  role: string
  is_active: boolean
}

const empty: UserFormState = { email: '', password: '', full_name: '', role: 'sales_rep', is_active: true }

interface UserModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
  edit?: User | null
}

export function UserModal({ open, onClose, onSaved, edit }: UserModalProps) {
  const token = useToken()
  const [form, setForm] = useState<UserFormState>(empty)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    if (edit) {
      setForm({
        email: edit.email,
        password: '',
        full_name: edit.full_name,
        role: edit.role,
        is_active: edit.is_active,
      })
    } else {
      setForm(empty)
    }
  }, [open, edit])

  const set = <K extends keyof UserFormState>(k: K, v: UserFormState[K]) =>
    setForm((f) => ({ ...f, [k]: v }))

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    setLoading(true)
    setError('')
    try {
      if (edit?.id) {
        await userApi.update(token, edit.id, {
          full_name: form.full_name,
          role: form.role,
          is_active: form.is_active,
        })
      } else {
        await userApi.create(token, {
          email: form.email,
          password: form.password,
          full_name: form.full_name,
          role: form.role,
        })
      }
      onSaved?.()
      onClose()
      if (!edit) setForm(empty)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={edit ? 'Sửa user' : 'Thêm user'}>
      <form onSubmit={submit}>
        {!edit && (
          <>
            <Field label="Email" required><Input type="email" value={form.email} onChange={(e) => set('email', e.target.value)} required /></Field>
            <Field label="Mật khẩu" required><Input type="password" value={form.password} onChange={(e) => set('password', e.target.value)} required minLength={8} /></Field>
          </>
        )}
        <Field label="Họ tên" required><Input value={form.full_name} onChange={(e) => set('full_name', e.target.value)} required /></Field>
        <Field label="Role">
          <Select value={form.role} onChange={(e) => set('role', e.target.value)}>
            <option value="admin">Admin</option>
            <option value="sales_rep">Sales</option>
            <option value="accountant">Accountant</option>
            <option value="director">Director</option>
          </Select>
        </Field>
        {edit && (
          <Field label="Trạng thái">
            <Select value={form.is_active ? '1' : '0'} onChange={(e) => set('is_active', e.target.value === '1')}>
              <option value="1">Active</option>
              <option value="0">Inactive</option>
            </Select>
          </Field>
        )}
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : edit ? 'Lưu' : 'Tạo'}</Button>
        </div>
      </form>
    </Modal>
  )
}
