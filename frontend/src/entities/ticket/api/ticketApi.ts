import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { CreateTicketInput, Ticket } from '../model/types'

export const ticketApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Ticket>> =>
    http(`/api/v1/tickets?${listParams(page, limit, filters)}`, { token }),
  create: (token: string, body: CreateTicketInput): Promise<{ data: Ticket }> =>
    http('/api/v1/tickets', { token, method: 'POST', body }),
  update: (token: string, id: string, body: Partial<CreateTicketInput>): Promise<{ data: Ticket }> =>
    http(`/api/v1/tickets/${id}`, { token, method: 'PUT', body }),
}
