import { NavLink, Outlet } from 'react-router-dom'
import { useAuth } from '@/app/providers/AuthProvider'
import './app-shell.css'

interface NavItem {
  to: string
  label: string
  icon: string
  admin?: boolean
}

const nav: NavItem[] = [
  { to: '/', label: 'Dashboard', icon: '◉' },
  { to: '/customers', label: 'Khách hàng', icon: '◎' },
  { to: '/tickets', label: 'Tickets', icon: '⚑' },
  { to: '/opportunities', label: 'Cơ hội', icon: '◈' },
  { to: '/quotations', label: 'Báo giá', icon: '▣' },
  { to: '/contracts', label: 'Hợp đồng', icon: '▤' },
  { to: '/shipments', label: 'Vận đơn', icon: '⬡' },
  { to: '/shipment-map', label: 'Bản đồ', icon: '⊕' },
  { to: '/invoices', label: 'Hóa đơn', icon: '▥' },
  { to: '/payments', label: 'Thanh toán', icon: '▦' },
  { to: '/reports', label: 'Báo cáo', icon: '▧' },
  { to: '/campaigns', label: 'Marketing', icon: '◐' },
  { to: '/users', label: 'Users', icon: '◉', admin: true },
]

export function AppShell() {
  const { session, logout } = useAuth()
  const isAdmin = session?.user.role === 'admin'

  if (!session) return null

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
          {nav.filter((item) => !item.admin || isAdmin).map((item) => (
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
            <NavLink to="/profile" className="shell-profile">{session.user.full_name}</NavLink>
            <small>{session.user.role}</small>
            <button type="button" className="shell-logout" onClick={logout}>Đăng xuất</button>
          </div>
        </header>
        <main className="shell-content"><Outlet /></main>
      </div>
    </div>
  )
}
