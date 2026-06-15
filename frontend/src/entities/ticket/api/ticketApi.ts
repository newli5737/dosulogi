import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { AddCommentInput, CreateTicketInput, Ticket, TicketComment, TicketDetail, UpdateTicketInput } from '../model/types'

export const ticketApi = {
  list: (page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Ticket>> =>
    http(`/api/v1/tickets?${listParams(page, limit, filters)}`),
  get: (id: string): Promise<{ data: TicketDetail }> =>
    http(`/api/v1/tickets/${id}`),
  create: (body: CreateTicketInput): Promise<{ data: Ticket }> =>
    http('/api/v1/tickets', { method: 'POST', body }),
  update: (id: string, body: UpdateTicketInput): Promise<{ data: Ticket }> =>
    http(`/api/v1/tickets/${id}`, { method: 'PUT', body }),
  addComment: (id: string, body: AddCommentInput): Promise<{ data: TicketComment }> =>
    http(`/api/v1/tickets/${id}/comments`, { method: 'POST', body }),
}
