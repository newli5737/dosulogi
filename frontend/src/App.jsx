import { useEffect, useState } from 'react'
import { getMe, getSummary, login } from './api'

function Login({ onSuccess }) {
  const [email, setEmail] = useState('admin@dosulogi.com')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const data = await login(email, password)
      localStorage.setItem('access_token', data.access_token)
      onSuccess(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-wrap">
      <form className="card login-form" onSubmit={handleSubmit}>
        <h2>Dosu Logi ERP</h2>
        <p>Đăng nhập hệ thống quản lý logistics</p>
        <label>Email</label>
        <input value={email} onChange={(e) => setEmail(e.target.value)} type="email" required />
        <label>Mật khẩu</label>
        <input value={password} onChange={(e) => setPassword(e.target.value)} type="password" required />
        {error && <div className="error">{error}</div>}
        <button type="submit" disabled={loading}>{loading ? 'Đang đăng nhập...' : 'Đăng nhập'}</button>
      </form>
    </div>
  )
}

function Dashboard({ user, token, onLogout }) {
  const [summary, setSummary] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    getSummary(token)
      .then(setSummary)
      .catch((e) => setError(e.message))
  }, [token])

  return (
    <div className="app">
      <nav className="nav">
        <h1>Dosu Logi · {user.full_name}</h1>
        <button onClick={onLogout}>Đăng xuất</button>
      </nav>
      <div className="container">
        <div className="card">
          <h2>Dashboard</h2>
          {error && <div className="error">{error}</div>}
          {summary && (
            <div className="grid">
              <div className="kpi"><label>Doanh thu</label><strong>{summary.revenue?.toLocaleString('vi-VN')} ₫</strong></div>
              <div className="kpi"><label>Shipments</label><strong>{summary.shipment_count}</strong></div>
              <div className="kpi"><label>KH mới (tháng)</label><strong>{summary.new_customers}</strong></div>
              <div className="kpi"><label>Công nợ</label><strong>{summary.total_ar?.toLocaleString('vi-VN')} ₫</strong></div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default function App() {
  const [session, setSession] = useState(null)
  const token = localStorage.getItem('access_token')

  useEffect(() => {
    if (!token) return
    getMe(token)
      .then((user) => setSession({ user, access_token: token }))
      .catch(() => localStorage.removeItem('access_token'))
  }, [token])

  function logout() {
    localStorage.removeItem('access_token')
    setSession(null)
  }

  if (!session) {
    return <Login onSuccess={setSession} />
  }

  return <Dashboard user={session.user} token={session.access_token} onLogout={logout} />
}
