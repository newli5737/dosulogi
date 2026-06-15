import { useState, type FormEvent } from 'react'
import { authApi } from '@/entities/session/api/sessionApi'
import { useToken } from '@/app/providers/AuthProvider'
import { Field, Input } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'

export function ProfilePage() {
  const token = useToken()
  const [current, setCurrent] = useState('')
  const [next, setNext] = useState('')
  const [confirm, setConfirm] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    if (next !== confirm) {
      setError('Mật khẩu xác nhận không khớp')
      return
    }
    if (next.length < 8) {
      setError('Mật khẩu mới tối thiểu 8 ký tự')
      return
    }
    setLoading(true)
    setError('')
    setMessage('')
    try {
      await authApi.changePassword(token, { old_password: current, new_password: next })
      setMessage('Đã đổi mật khẩu thành công')
      setCurrent('')
      setNext('')
      setConfirm('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page-card" style={{ maxWidth: 480 }}>
      <div className="page-header"><h1>Tài khoản</h1></div>
      <h2 className="section-title">Đổi mật khẩu</h2>
      <form onSubmit={submit}>
        <Field label="Mật khẩu hiện tại" required>
          <Input type="password" value={current} onChange={(e) => setCurrent(e.target.value)} required />
        </Field>
        <Field label="Mật khẩu mới" required>
          <Input type="password" value={next} onChange={(e) => setNext(e.target.value)} required minLength={8} />
        </Field>
        <Field label="Xác nhận mật khẩu mới" required>
          <Input type="password" value={confirm} onChange={(e) => setConfirm(e.target.value)} required minLength={8} />
        </Field>
        {message && <p style={{ color: 'var(--success, #16a34a)' }}>{message}</p>}
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Cập nhật mật khẩu'}</Button>
        </div>
      </form>
    </div>
  )
}
