import { useState } from 'react'
import { Modal } from '../../../shared/ui/Modal/Modal'
import { Field, Input, Select } from '../../../shared/ui/Form/Form'
import { Button } from '../../../shared/ui/Button/Button'
import { userApi } from '../../../entities/user/api/userApi'
import { useToken } from '../../../app/providers/AuthProvider'

export function UserModal({ open, onClose, onSaved }) {
  const token = useToken()
  const [form, setForm] = useState({ email: '', password: '', full_name: '', role: 'sales_rep' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const set = (k, v) => setForm((f) => ({ ...f, [k]: v }))

  async function submit(e) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await userApi.create(token, form)
      onSaved?.()
      onClose()
      setForm({ email: '', password: '', full_name: '', role: 'sales_rep' })
    } catch (err) {
      setError(err.message)
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
