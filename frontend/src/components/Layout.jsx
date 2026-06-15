import { NavLink, Outlet, useNavigate } from 'react-router-dom'

const links = [
  { to: '/', label: 'Dashboard' },
  { to: '/customers', label: 'CRM · Khách hàng' },
  { to: '/opportunities', label: 'Sales · Cơ hội' },
  { to: '/contracts', label: 'Sales · Hợp đồng' },
  { to: '/quotations', label: 'Sales · Báo giá' },
  { to: '/shipments', label: 'Tracking · Vận đơn' },
  { to: '/map', label: 'Tracking · Bản đồ' },
  { to: '/invoices', label: 'Kế toán · Hóa đơn' },
  { to: '/payments', label: 'Kế toán · Thanh toán' },
  { to: '/reports', label: 'Kế toán · Báo cáo' },
  { to: '/campaigns', label: 'Marketing' },
  { to: '/users', label: 'Users (Admin)', admin: true },
]

export default function Layout({ user, onLogout }) {
  const nav = useNavigate()
  const visible = links.filter((l) => !l.admin || user.role === 'admin')

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">Dosu Logi</div>
        <nav>
          {visible.map((l) => (
            <NavLink key={l.to} to={l.to} end={l.to === '/'} className={({ isActive }) => (isActive ? 'active' : '')}>
              {l.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <div className="main">
        <header className="topbar">
          <span>{user.full_name} · {user.role}</span>
          <button type="button" onClick={() => { onLogout(); nav('/login') }}>Đăng xuất</button>
        </header>
        <main className="content"><Outlet context={{ user }} /></main>
      </div>
    </div>
  )
}
