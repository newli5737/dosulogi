const API_URL = import.meta.env.VITE_API_URL || ''

function headers(token) {
  const h = { 'Content-Type': 'application/json' }
  if (token) h.Authorization = `Bearer ${token}`
  return h
}

async function request(path, token, options = {}) {
  const res = await fetch(`${API_URL}${path}`, {
    credentials: 'include',
    ...options,
    headers: { ...headers(token), ...options.headers },
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || res.statusText)
  return data
}

export const login = (email, password) =>
  request('/api/v1/auth/login', null, { method: 'POST', body: JSON.stringify({ email, password }) })

export const getMe = (token) => request('/api/v1/auth/me', token)
export const getSummary = (token) => request('/api/v1/dashboard/summary', token)
export const getSalesFunnel = (token) => request('/api/v1/dashboard/sales-funnel', token)
export const getShipmentMap = (token) => request('/api/v1/dashboard/shipment-map', token)

export const listCustomers = (token, q = '') => request(`/api/v1/customers?limit=50${q ? `&q=${q}` : ''}`, token)
export const createCustomer = (token, body) => request('/api/v1/customers', token, { method: 'POST', body: JSON.stringify(body) })

export const listOpportunities = (token) => request('/api/v1/opportunities?limit=50', token)
export const createOpportunity = (token, body) => request('/api/v1/opportunities', token, { method: 'POST', body: JSON.stringify(body) })

export const listContracts = (token) => request('/api/v1/contracts?limit=50', token)
export const listQuotations = (token) => request('/api/v1/quotations?limit=50', token)

export const listShipments = (token) => request('/api/v1/shipments?limit=50', token)
export const syncShipment = (token, id) => request(`/api/v1/shipments/${id}/sync`, token, { method: 'POST' })

export const listInvoices = (token) => request('/api/v1/invoices?limit=50', token)
export const listPayments = (token) => request('/api/v1/payments?limit=50', token)
export const getRevenueReport = (token) => request('/api/v1/reports/revenue', token)
export const getARReport = (token) => request('/api/v1/reports/ar', token)

export const listCampaigns = (token) => request('/api/v1/campaigns?limit=50', token)
export const createCampaign = (token, body) => request('/api/v1/campaigns', token, { method: 'POST', body: JSON.stringify(body) })
export const sendCampaign = (token, id) => request(`/api/v1/campaigns/${id}/send`, token, { method: 'POST' })

export const listUsers = (token) => request('/api/v1/users?limit=50', token)

export function asArray(value) {
  if (Array.isArray(value)) return value
  if (value && Array.isArray(value.data)) return value.data
  return []
}

export function paginatedItems(res) {
  return asArray(res?.data ?? res)
}
