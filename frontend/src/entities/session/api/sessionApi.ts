import { http } from '@/shared/api/http'
import type { LoginResponse, UserBrief } from '@/shared/api/types'

export interface DashboardSummary {
  revenue: number
  shipment_count: number
  new_customers: number
  total_ar: number
}

export interface FunnelStage {
  stage: string
  count: number
  value: number
}

export interface RevenueReportRow {
  label: string
  amount: number
}

export interface ARReportRow {
  customer_id: string
  customer_name: string
  total_due: number
  invoice_count: number
}

export const authApi = {
  login: (body: { email: string; password: string }): Promise<LoginResponse> =>
    http('/api/v1/auth/login', { method: 'POST', body }),
  me: (token: string): Promise<UserBrief> =>
    http('/api/v1/auth/me', { token }),
  changePassword: (token: string, body: { old_password: string; new_password: string }): Promise<{ message: string }> =>
    http('/api/v1/auth/me/password', { token, method: 'PUT', body }),
}

export const dashboardApi = {
  summary: (token: string): Promise<DashboardSummary> =>
    http('/api/v1/dashboard/summary', { token }),
  funnel: (token: string): Promise<FunnelStage[]> =>
    http('/api/v1/dashboard/sales-funnel', { token }),
}

export const reportApi = {
  revenue: (token: string, from?: string, to?: string): Promise<RevenueReportRow[]> => {
    const q = new URLSearchParams()
    if (from) q.set('from', from)
    if (to) q.set('to', to)
    const qs = q.toString()
    return http(`/api/v1/reports/revenue${qs ? `?${qs}` : ''}`, { token })
  },
  ar: (token: string): Promise<ARReportRow[]> =>
    http('/api/v1/reports/ar', { token }),
}
