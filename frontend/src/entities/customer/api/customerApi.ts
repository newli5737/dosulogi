import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { CreateCustomerInput, Customer, CustomerDetail } from '../model/types'

export const customerApi = {
  list: (page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Customer>> =>
    http(`/api/v1/customers?${listParams(page, limit, filters)}`),
  get: (id: string): Promise<{ data: CustomerDetail }> =>
    http(`/api/v1/customers/${id}`),
  create: (body: CreateCustomerInput): Promise<{ data: Customer }> =>
    http('/api/v1/customers', { method: 'POST', body }),
  update: (id: string, body: Partial<CreateCustomerInput>): Promise<{ data: Customer }> =>
    http(`/api/v1/customers/${id}`, { method: 'PUT', body }),
  remove: (id: string): Promise<void> =>
    http(`/api/v1/customers/${id}`, { method: 'DELETE' }),
}
