import { useCallback, useEffect, useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { useAuth } from '@/app/providers/AuthProvider'
import { LOGO_ALT, LOGO_SRC } from '@/shared/config/brand'
import './app-shell.css'

interface NavItem {
  to: string
  label: string
  icon: string
  admin?: boolean
}

interface NavGroup {
  id: string
  label: string
  items: NavItem[]
}

const navGroups: NavGroup[] = [
  {
    id: 'overview',
    label: 'Tổng quan',
    items: [{ to: '/', label: 'Dashboard', icon: '◉' }],
  },
  {
    id: 'crm',
    label: 'CRM',
    items: [
      { to: '/customers', label: 'Khách hàng', icon: '◎' },
      { to: '/tickets', label: 'Tickets', icon: '⚑' },
    ],
  },
  {
    id: 'sales',
    label: 'Bán hàng',
    items: [
      { to: '/opportunities', label: 'Cơ hội', icon: '◈' },
      { to: '/quotations', label: 'Báo giá', icon: '▣' },
      { to: '/contracts', label: 'Hợp đồng', icon: '▤' },
    ],
  },
  {
    id: 'logistics',
    label: 'Vận hành',
    items: [
      { to: '/shipments', label: 'Vận đơn', icon: '⬡' },
      { to: '/shipment-map', label: 'Bản đồ', icon: '⊕' },
    ],
  },
  {
    id: 'accounting',
    label: 'Kế toán',
    items: [
      { to: '/invoices', label: 'Hóa đơn', icon: '▥' },
      { to: '/payments', label: 'Thanh toán', icon: '▦' },
      { to: '/reports', label: 'Báo cáo', icon: '▧' },
    ],
  },
  {
    id: 'marketing',
    label: 'Marketing',
    items: [{ to: '/campaigns', label: 'Chiến dịch', icon: '◐' }],
  },
  {
    id: 'system',
    label: 'Hệ thống',
    items: [{ to: '/users', label: 'Users', icon: '◉', admin: true }],
  },
]

const STORAGE_KEY = 'dosulogi.sidebar.collapsed'
const OPEN_GROUPS_KEY = 'dosulogi.sidebar.openGroups'

function loadCollapsed(): boolean {
  try { return localStorage.getItem(STORAGE_KEY) === '1' } catch { return false }
}

function loadOpenGroups(): Record<string, boolean> {
  try {
    const raw = localStorage.getItem(OPEN_GROUPS_KEY)
    return raw ? JSON.parse(raw) as Record<string, boolean> : {}
  } catch { return {} }
}

export function AppShell() {
  const { session, logout } = useAuth()
  const isAdmin = session?.user.role === 'admin'
  const [collapsed, setCollapsed] = useState(loadCollapsed)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>(() => {
    const saved = loadOpenGroups()
    const defaults: Record<string, boolean> = {}
    for (const g of navGroups) defaults[g.id] = saved[g.id] ?? true
    return defaults
  })

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, collapsed ? '1' : '0')
  }, [collapsed])

  useEffect(() => {
    localStorage.setItem(OPEN_GROUPS_KEY, JSON.stringify(openGroups))
  }, [openGroups])

  const toggleGroup = useCallback((id: string) => {
    if (collapsed) setCollapsed(false)
    setOpenGroups((prev) => ({ ...prev, [id]: !prev[id] }))
  }, [collapsed])

  if (!session) return null

  const visibleGroups = navGroups.map((g) => ({
    ...g,
    items: g.items.filter((item) => !item.admin || isAdmin),
  })).filter((g) => g.items.length > 0)

  const sidebar = (
    <aside className={`shell-sidebar ${collapsed ? 'shell-sidebar--collapsed' : ''}`}>
      <div className="shell-brand">
        <img src={LOGO_SRC} alt={LOGO_ALT} className="shell-logo-img" />
        {!collapsed && (
          <div>
            <strong>Dosu Logi</strong>
            <small>ERP / CRM</small>
          </div>
        )}
      </div>
      <nav className="shell-nav">
        {visibleGroups.map((group) => (
          <div key={group.id} className="shell-nav-group">
            <button
              type="button"
              className="shell-nav-group__toggle"
              onClick={() => toggleGroup(group.id)}
              title={group.label}
            >
              {!collapsed && <span className="shell-nav-group__label">{group.label}</span>}
              <span className={`shell-nav-group__chevron ${openGroups[group.id] ? 'open' : ''}`}>▾</span>
            </button>
            {(openGroups[group.id] || collapsed) && (
              <div className="shell-nav-group__items">
                {group.items.map((item) => (
                  <NavLink
                    key={item.to}
                    to={item.to}
                    end={item.to === '/'}
                    title={item.label}
                    className={({ isActive }) => `shell-link ${isActive ? 'active' : ''}`}
                    onClick={() => setMobileOpen(false)}
                  >
                    <span className="shell-icon">{item.icon}</span>
                    {!collapsed && item.label}
                  </NavLink>
                ))}
              </div>
            )}
          </div>
        ))}
      </nav>
      {!collapsed && (
        <button type="button" className="shell-collapse-btn" onClick={() => setCollapsed(true)} aria-label="Thu gọn sidebar">
          « Thu gọn
        </button>
      )}
    </aside>
  )

  return (
    <div className={`shell ${collapsed ? 'shell--collapsed' : ''} ${mobileOpen ? 'shell--mobile-open' : ''}`}>
      {mobileOpen && <button type="button" className="shell-backdrop" aria-label="Đóng menu" onClick={() => setMobileOpen(false)} />}
      {sidebar}
      <div className="shell-main">
        <header className="shell-topbar">
          <div className="shell-topbar__left">
            <button
              type="button"
              className="shell-menu-btn"
              onClick={() => {
                if (window.innerWidth <= 900) setMobileOpen((v) => !v)
                else setCollapsed((v) => !v)
              }}
              aria-label="Toggle menu"
            >
              ☰
            </button>
            <img src={LOGO_SRC} alt={LOGO_ALT} className="shell-topbar-logo" />
            <span className="shell-topbar-title">Dosu Logi ERP</span>
          </div>
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
