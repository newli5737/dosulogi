import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import type { ReactNode } from 'react'
import { AuthProvider, useAuth } from '@/app/providers/AuthProvider'
import { AppShell } from '@/widgets/app-shell/ui/AppShell'
import { LoginPage } from '@/pages/login/ui/LoginPage'
import { DashboardPage } from '@/pages/dashboard/ui/DashboardPage'
import { CustomersPage } from '@/pages/customers/ui/CustomersPage'
import { CustomerDetailPage } from '@/pages/customer-detail/ui/CustomerDetailPage'
import { TicketsPage } from '@/pages/tickets/ui/TicketsPage'
import { OpportunitiesPage } from '@/pages/opportunities/ui/OpportunitiesPage'
import { ContractsPage } from '@/pages/contracts/ui/ContractsPage'
import { QuotationsPage } from '@/pages/quotations/ui/QuotationsPage'
import { ShipmentsPage } from '@/pages/shipments/ui/ShipmentsPage'
import { CampaignsPage } from '@/pages/campaigns/ui/CampaignsPage'
import { ReportsPage } from '@/pages/reports/ui/ReportsPage'
import { UsersPage } from '@/pages/users/ui/UsersPage'
import { ShipmentMapPage } from '@/pages/shipment-map/ui/ShipmentMapPage'
import { InvoicesPage } from '@/pages/invoices/ui/InvoicesPage'
import { PaymentsPage } from '@/pages/payments/ui/PaymentsPage'
import { ProfilePage } from '@/pages/profile/ui/ProfilePage'
import { InboxPage } from '@/pages/inbox/ui/InboxPage'
import { ChatAccountsPage } from '@/pages/chat-accounts/ui/ChatAccountsPage'

function Guard({ children }: { children: ReactNode }) {
  const { session, checking } = useAuth()
  if (checking) return <div className="login-page" style={{ placeItems: 'center', display: 'grid' }}>Đang tải...</div>
  if (!session) return <Navigate to="/login" replace />
  return children
}

function LoginRoute() {
  const { session, checking } = useAuth()
  if (checking) return <div className="login-page" style={{ placeItems: 'center', display: 'grid' }}>Đang tải...</div>
  if (session) return <Navigate to="/" replace />
  return <LoginPage />
}

function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<LoginRoute />} />
      <Route element={<Guard><AppShell /></Guard>}>
        <Route index element={<DashboardPage />} />
        <Route path="customers" element={<CustomersPage />} />
        <Route path="customers/:id" element={<CustomerDetailPage />} />
        <Route path="tickets" element={<TicketsPage />} />
        <Route path="inbox" element={<InboxPage />} />
        <Route path="opportunities" element={<OpportunitiesPage />} />
        <Route path="contracts" element={<ContractsPage />} />
        <Route path="quotations" element={<QuotationsPage />} />
        <Route path="invoices" element={<InvoicesPage />} />
        <Route path="payments" element={<PaymentsPage />} />
        <Route path="shipments" element={<ShipmentsPage />} />
        <Route path="shipment-map" element={<ShipmentMapPage />} />
        <Route path="campaigns" element={<CampaignsPage />} />
        <Route path="reports" element={<ReportsPage />} />
        <Route path="users" element={<UsersPage />} />
        <Route path="chat-accounts" element={<ChatAccountsPage />} />
        <Route path="profile" element={<ProfilePage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppRouter />
      </AuthProvider>
    </BrowserRouter>
  )
}
