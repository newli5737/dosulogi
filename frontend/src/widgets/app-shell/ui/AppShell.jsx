import { NavLink, Outlet } from 'react-router-dom'
import { useAuth } from '../../../app/providers/AuthProvider'
import './app-shell.css'

const nav = [
  { to: '/', label: 'Dashboard', icon: '◉' },
  { to: '/customers', label: 'Khách hàng', icon: '◎' },
  { to: '/tickets', label: 'Tickets', icon: '⚑' },
  { to: '/opportunities', label: 'Cơ hội', icon: '◈' },
  { to: '/quotations', label: 'Báo giá', icon: '▣' },
  { to: '/contracts', label: 'Hợp đồng', icon: '▤' },
  { to: '/shipments', label: 'Vận đơn', icon: '⬡' },
  { to: '/invoices', label: 'Hóa đơn', icon: '▥' },
  { to: '/payments', label: 'Thanh toán', icon: '▦' },
  { to: '/campaigns', label: 'Marketing', icon: '◐' },
]

export function AppShell() {
  const { session, logout } = useAuth()
  return (
    <div className="shell">
      <aside className="shell-sidebar">
        <div className="shell-brand">
          <span className="shell-logo">DL</span>
          <div>
            <strong>Dosu Logi</strong>
            <small>ERP / CRM</small>
          </div>
        </div>
        <nav className="shell-nav">
          {nav.map((item) => (
            <NavLink key={item.to} to={item.to} end={item.to === '/'} className={({ isActive }) => `shell-link ${isActive ? 'active' : ''}`}>
              <span className="shell-icon">{item.icon}</span>
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <div className="shell-main">
        <header className="shell-topbar">
          <div />
          <div className="shell-user">
            <span>{session.user.full_name}</span>
            <small>{session.user.role}</small>
            <button type="button" className="shell-logout" onClick={logout}>Đăng xuất</button>
          </div>
        </header>
        <main className="shell-content"><Outlet /></main>
      </div>
    </div>
  )
}
