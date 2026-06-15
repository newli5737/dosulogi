import { useEffect, useState } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { getMe } from './api'
import Layout from './components/Layout'
import LoginPage from './pages/Login'
import DashboardPage from './pages/Dashboard'
import CustomersPage from './pages/Customers'
import OpportunitiesPage from './pages/Opportunities'
import ContractsPage from './pages/Contracts'
import QuotationsPage from './pages/Quotations'
import ShipmentsPage from './pages/Shipments'
import ShipmentMapPage from './pages/ShipmentMap'
import InvoicesPage from './pages/Invoices'
import PaymentsPage from './pages/Payments'
import ReportsPage from './pages/Reports'
import CampaignsPage from './pages/Campaigns'
import UsersPage from './pages/Users'

function AppRoutes() {
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
    return (
      <Routes>
        <Route path="/login" element={<LoginPage onSuccess={setSession} />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    )
  }

  return (
    <Routes>
      <Route element={<Layout user={session.user} onLogout={logout} />}>
        <Route index element={<DashboardPage />} />
        <Route path="customers" element={<CustomersPage />} />
        <Route path="opportunities" element={<OpportunitiesPage />} />
        <Route path="contracts" element={<ContractsPage />} />
        <Route path="quotations" element={<QuotationsPage />} />
        <Route path="shipments" element={<ShipmentsPage />} />
        <Route path="map" element={<ShipmentMapPage />} />
        <Route path="invoices" element={<InvoicesPage />} />
        <Route path="payments" element={<PaymentsPage />} />
        <Route path="reports" element={<ReportsPage />} />
        <Route path="campaigns" element={<CampaignsPage />} />
        <Route path="users" element={<UsersPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  )
}
