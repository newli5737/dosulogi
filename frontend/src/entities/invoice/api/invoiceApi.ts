import { http, httpBlob, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { CreateInvoicePayload, Invoice } from '../model/types'

export const invoiceApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Invoice>> =>
    http(`/api/v1/invoices?${listParams(page, limit, filters)}`, { token }),
  create: (token: string, body: CreateInvoicePayload): Promise<Invoice> =>
    http('/api/v1/invoices', { token, method: 'POST', body }),
  update: (token: string, id: string, body: CreateInvoicePayload): Promise<Invoice> =>
    http(`/api/v1/invoices/${id}`, { token, method: 'PUT', body }),
  send: (token: string, id: string): Promise<{ message: string }> =>
    http(`/api/v1/invoices/${id}/send`, { token, method: 'POST' }),
  cancel: (token: string, id: string): Promise<{ message: string }> =>
    http(`/api/v1/invoices/${id}/cancel`, { token, method: 'POST' }),
  download: (token: string, id: string): Promise<Blob> =>
    httpBlob(`/api/v1/invoices/${id}/download`, token),
}
