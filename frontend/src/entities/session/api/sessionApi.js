import { http } from '../../../shared/api/http'

export const authApi = {
  login: (body) => http('/api/v1/auth/login', { method: 'POST', body }),
  me: (token) => http('/api/v1/auth/me', { token }),
}

export const dashboardApi = {
  summary: (token) => http('/api/v1/dashboard/summary', { token }),
  funnel: (token) => http('/api/v1/dashboard/sales-funnel', { token }),
}

export const paymentApi = {
  list: (token, page, limit) => http(`/api/v1/payments?page=${page}&limit=${limit}`, { token }),
}

export const invoiceApi = {
  list: (token, page, limit) => http(`/api/v1/invoices?page=${page}&limit=${limit}`, { token }),
}

export const reportApi = {
  revenue: (token) => http('/api/v1/reports/revenue', { token }),
  ar: (token) => http('/api/v1/reports/ar', { token }),
}
