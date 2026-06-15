import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Quotation, QuotationInput } from '../model/types'
import type { Contract } from '@/entities/contract/model/types'

export const quotationApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Quotation>> =>
    http(`/api/v1/quotations?${listParams(page, limit, filters)}`, { token }),
  create: (token: string, body: QuotationInput): Promise<Quotation> =>
    http('/api/v1/quotations', { token, method: 'POST', body }),
  update: (token: string, id: string, body: QuotationInput): Promise<Quotation> =>
    http(`/api/v1/quotations/${id}`, { token, method: 'PUT', body }),
  send: (token: string, id: string): Promise<{ message: string }> =>
    http(`/api/v1/quotations/${id}/send`, { token, method: 'POST' }),
  convert: (token: string, id: string): Promise<Contract> =>
    http(`/api/v1/quotations/${id}/convert`, { token, method: 'POST' }),
}
