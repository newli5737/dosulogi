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

export const opportunityApi = {
  list: (token, page, limit) => http(`/api/v1/opportunities?page=${page}&limit=${limit}`, { token }),
}

export const contractApi = {
  list: (token, page, limit) => http(`/api/v1/contracts?page=${page}&limit=${limit}`, { token }),
}

export const quotationApi = {
  list: (token, page, limit) => http(`/api/v1/quotations?page=${page}&limit=${limit}`, { token }),
}

export const shipmentApi = {
  list: (token, page, limit) => http(`/api/v1/shipments?page=${page}&limit=${limit}`, { token }),
}

export const campaignApi = {
  list: (token, page, limit) => http(`/api/v1/campaigns?page=${page}&limit=${limit}`, { token }),
}
