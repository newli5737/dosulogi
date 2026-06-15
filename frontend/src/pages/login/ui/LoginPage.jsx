import { useState } from 'react'
import { useAuth } from '../../../app/providers/AuthProvider'
import './login-page.css'

export function LoginPage() {
  const { login } = useAuth()
  const [email, setEmail] = useState('admin@dosulogi.com')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function onSubmit(e) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await login(email, password)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-hero">
        <h1>Dosu Logi</h1>
        <p>Hệ thống quản lý logistics tích hợp CRM & ERP</p>
        <ul>
          <li>Quản lý khách hàng & pipeline bán hàng</li>
          <li>Tracking vận đơn realtime</li>
          <li>Kế toán & hóa đơn tự động</li>
        </ul>
      </div>
      <div className="login-panel">
        <form className="login-form" onSubmit={onSubmit}>
          <h2>Đăng nhập</h2>
          <p className="login-sub">Nhập thông tin tài khoản của bạn</p>
          <label>Email<input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required autoComplete="username" /></label>
          <label>Mật khẩu<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required autoComplete="current-password" /></label>
          {error && <div className="login-error">{error}</div>}
          <button type="submit" disabled={loading}>{loading ? 'Đang đăng nhập...' : 'Đăng nhập'}</button>
        </form>
      </div>
    </div>
  )
}
