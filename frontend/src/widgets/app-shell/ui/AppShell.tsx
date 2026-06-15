import { useCallback, useEffect, useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import type { LucideIcon } from 'lucide-react'
import {
  LayoutDashboard,
  Users,
  Ticket,
  Target,
  FileText,
  FileSignature,
  Package,
  MapPin,
  Receipt,
  CreditCard,
  BarChart3,
  Megaphone,
  MessageSquare,
  Settings2,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { useAuth } from '@/app/providers/AuthProvider'
import { COMPANY_NAME, COMPANY_TAGLINE, LOGO_ALT, LOGO_SRC } from '@/shared/config/brand'
import './app-shell.css'

interface NavItem {
  to: string
  label: string
  icon: LucideIcon
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
    items: [{ to: '/', label: 'Dashboard', icon: LayoutDashboard }],
  },
  {
    id: 'crm',
    label: 'CRM',
    items: [
      { to: '/customers', label: 'Khách hàng', icon: Users },
      { to: '/tickets', label: 'Tickets', icon: Ticket },
      { to: '/inbox', label: 'Hộp thư chat', icon: MessageSquare },
    ],
  },
  {
    id: 'sales',
    label: 'Bán hàng',
    items: [
      { to: '/opportunities', label: 'Cơ hội', icon: Target },
      { to: '/quotations', label: 'Báo giá', icon: FileText },
      { to: '/contracts', label: 'Hợp đồng', icon: FileSignature },
    ],
  },
  {
    id: 'logistics',
    label: 'Vận hành',
    items: [
      { to: '/shipments', label: 'Vận đơn', icon: Package },
      { to: '/shipment-map', label: 'Bản đồ', icon: MapPin },
    ],
  },
  {
    id: 'accounting',
    label: 'Kế toán',
    items: [
      { to: '/invoices', label: 'Hóa đơn', icon: Receipt },
      { to: '/payments', label: 'Thanh toán', icon: CreditCard },
      { to: '/reports', label: 'Báo cáo', icon: BarChart3 },
    ],
  },
  {
    id: 'marketing',
    label: 'Marketing',
    items: [{ to: '/campaigns', label: 'Chiến dịch', icon: Megaphone }],
  },
  {
    id: 'system',
    label: 'Hệ thống',
    items: [
      { to: '/chat-accounts', label: 'Kênh chat', icon: Settings2, admin: true },
      { to: '/users', label: 'Users', icon: Users, admin: true },
    ],
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
  const isAdmin = session?.role === 'admin'
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

  const toggleCollapsed = useCallback(() => {
    setCollapsed((v) => !v)
  }, [])

  const toggleGroup = useCallback((id: string) => {
    if (collapsed) setCollapsed(false)
    setOpenGroups((prev) => ({ ...prev, [id]: !prev[id] }))
  }, [collapsed])

  if (!session) return null

  const visibleGroups = navGroups.map((g) => ({
    ...g,
    items: g.items.filter((item) => !item.admin || isAdmin),
  })).filter((g) => g.items.length > 0)

  return (
    <div className={`shell ${collapsed ? 'shell--collapsed' : ''} ${mobileOpen ? 'shell--mobile-open' : ''}`}>
      {mobileOpen && (
        <button type="button" className="shell-backdrop" aria-label="Đóng menu" onClick={() => setMobileOpen(false)} />
      )}
      <aside className={`shell-sidebar ${collapsed ? 'shell-sidebar--collapsed' : ''}`}>
        <button
          type="button"
          className="shell-sidebar-toggle"
          onClick={() => {
            if (window.innerWidth <= 900) setMobileOpen((v) => !v)
            else toggleCollapsed()
          }}
          aria-label={collapsed ? 'Mở sidebar' : 'Thu gọn sidebar'}
          title={collapsed ? 'Mở rộng' : 'Thu gọn'}
        >
          {collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
        </button>
        <div className="shell-brand">
          <img src={LOGO_SRC} alt={LOGO_ALT} className="shell-logo-img" />
          {!collapsed && (
            <div>
              <strong>{COMPANY_NAME}</strong>
              <small>{COMPANY_TAGLINE}</small>
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
                <ChevronDown size={14} className={`shell-nav-group__chevron ${openGroups[group.id] ? 'open' : ''}`} />
              </button>
              {(openGroups[group.id] || collapsed) && (
                <div className="shell-nav-group__items">
                  {group.items.map((item) => {
                    const Icon = item.icon
                    return (
                      <NavLink
                        key={item.to}
                        to={item.to}
                        end={item.to === '/'}
                        title={item.label}
                        className={({ isActive }) => `shell-link ${isActive ? 'active' : ''}`}
                        onClick={() => setMobileOpen(false)}
                      >
                        <span className="shell-icon"><Icon size={18} strokeWidth={2} /></span>
                        {!collapsed && item.label}
                      </NavLink>
                    )
                  })}
                </div>
              )}
            </div>
          ))}
        </nav>
      </aside>
      <div className="shell-main">
        <header className="shell-topbar">
          <div />
          <div className="shell-user">
            <NavLink to="/profile" className="shell-profile">{session.full_name}</NavLink>
            <small>{session.role}</small>
            <button type="button" className="shell-logout" onClick={() => void logout()}>Đăng xuất</button>
          </div>
        </header>
        <main className="shell-content"><Outlet /></main>
      </div>
    </div>
  )
}
