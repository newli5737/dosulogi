import { http, httpBlob, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { CreateInvoicePayload, Invoice } from '../model/types'

export const invoiceApi = {
  list: (page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Invoice>> =>
    http(`/api/v1/invoices?${listParams(page, limit, filters)}`),
  create: (body: CreateInvoicePayload): Promise<Invoice> =>
    http('/api/v1/invoices', { method: 'POST', body }),
  update: (id: string, body: CreateInvoicePayload): Promise<Invoice> =>
    http(`/api/v1/invoices/${id}`, { method: 'PUT', body }),
  send: (id: string): Promise<{ message: string }> =>
    http(`/api/v1/invoices/${id}/send`, { method: 'POST' }),
  cancel: (id: string): Promise<{ message: string }> =>
    http(`/api/v1/invoices/${id}/cancel`, { method: 'POST' }),
  download: (id: string): Promise<Blob> =>
    httpBlob(`/api/v1/invoices/${id}/download`),
}
