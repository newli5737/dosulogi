import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuth } from './providers/AuthProvider'
import { AppShell } from '../widgets/app-shell/ui/AppShell'
import { LoginPage } from '../pages/login/ui/LoginPage'
import { DashboardPage } from '../pages/dashboard/ui/DashboardPage'
import { CustomersPage } from '../pages/customers/ui/CustomersPage'
import { TicketsPage } from '../pages/tickets/ui/TicketsPage'
import { OpportunitiesPage, ContractsPage, QuotationsPage, InvoicesPage, PaymentsPage, ShipmentsPage, CampaignsPage } from '../pages/sales/ui/SalesPages'

function Guard({ children }) {
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
        <Route path="tickets" element={<TicketsPage />} />
        <Route path="opportunities" element={<OpportunitiesPage />} />
        <Route path="contracts" element={<ContractsPage />} />
        <Route path="quotations" element={<QuotationsPage />} />
        <Route path="invoices" element={<InvoicesPage />} />
        <Route path="payments" element={<PaymentsPage />} />
        <Route path="shipments" element={<ShipmentsPage />} />
        <Route path="campaigns" element={<CampaignsPage />} />
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
