import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Payment, PaymentInput } from '../model/types'

export const paymentApi = {
  list: (page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Payment>> =>
    http(`/api/v1/payments?${listParams(page, limit, filters)}`),
  create: (body: PaymentInput): Promise<Payment> =>
    http('/api/v1/payments', { method: 'POST', body }),
}
