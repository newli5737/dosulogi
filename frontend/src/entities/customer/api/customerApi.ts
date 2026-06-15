import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { CreateCustomerInput, Customer, CustomerDetail } from '../model/types'

export const customerApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Customer>> =>
    http(`/api/v1/customers?${listParams(page, limit, filters)}`, { token }),
  get: (token: string, id: string): Promise<{ data: CustomerDetail }> =>
    http(`/api/v1/customers/${id}`, { token }),
  create: (token: string, body: CreateCustomerInput): Promise<{ data: Customer }> =>
    http('/api/v1/customers', { token, method: 'POST', body }),
  update: (token: string, id: string, body: Partial<CreateCustomerInput>): Promise<{ data: Customer }> =>
    http(`/api/v1/customers/${id}`, { token, method: 'PUT', body }),
  remove: (token: string, id: string): Promise<void> =>
    http(`/api/v1/customers/${id}`, { token, method: 'DELETE' }),
}
