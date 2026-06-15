import { http } from '@/shared/api/http'
import type { UserBrief } from '@/shared/api/types'

export interface LoginResponse {
  user: UserBrief
}

export interface DashboardSummary {
  revenue: number
  shipment_count: number
  new_customers: number
  total_ar: number
  open_tickets: number
  active_opportunities: number
  paid_invoices: number
}

export interface TrendPoint {
  label: string
  amount: number
}

export interface StatusCount {
  status: string
  count: number
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
  refresh: (): Promise<LoginResponse> =>
    http('/api/v1/auth/refresh', { method: 'POST' }),
  logout: (): Promise<{ message: string }> =>
    http('/api/v1/auth/logout', { method: 'POST' }),
  me: (): Promise<UserBrief> =>
    http('/api/v1/auth/me'),
  changePassword: (body: { old_password: string; new_password: string }): Promise<{ message: string }> =>
    http('/api/v1/auth/me/password', { method: 'PUT', body }),
}

export const dashboardApi = {
  summary: (): Promise<DashboardSummary> =>
    http('/api/v1/dashboard/summary'),
  funnel: (): Promise<FunnelStage[]> =>
    http('/api/v1/dashboard/sales-funnel'),
  revenueTrend: (): Promise<TrendPoint[]> =>
    http('/api/v1/dashboard/revenue-trend'),
  ticketStats: (): Promise<StatusCount[]> =>
    http('/api/v1/dashboard/ticket-stats'),
  shipmentStats: (): Promise<StatusCount[]> =>
    http('/api/v1/dashboard/shipment-stats'),
}

export const reportApi = {
  revenue: (from?: string, to?: string): Promise<RevenueReportRow[]> => {
    const q = new URLSearchParams()
    if (from) q.set('from', from)
    if (to) q.set('to', to)
    const qs = q.toString()
    return http(`/api/v1/reports/revenue${qs ? `?${qs}` : ''}`)
  },
  ar: (): Promise<ARReportRow[]> =>
    http('/api/v1/reports/ar'),
}
