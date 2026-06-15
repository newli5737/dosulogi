import type { UserBrief, UUID } from '@/shared/api/types'

export interface Ticket {
  id: UUID
  code: string
  customer_id: UUID
  title: string
  description?: string | null
  priority: string
  status: string
  category?: string | null
  assigned_to?: UUID | null
  assigned_user?: UserBrief | null
  customer?: { id?: UUID; name: string; code?: string }
  is_overdue?: boolean
  sla_deadline?: string | null
  created_at?: string
}

export interface TicketComment {
  id: UUID
  ticket_id: UUID
  body: string
  is_internal: boolean
  created_by?: UserBrief | null
  created_at: string
}

export interface TicketDetail {
  ticket: Ticket
  comments: TicketComment[]
}

export interface CreateTicketInput {
  customer_id: UUID
  title: string
  description?: string | null
  priority: string
  category?: string | null
}

export interface UpdateTicketInput {
  status?: string
  priority?: string
  assigned_to?: UUID | null
}

export interface AddCommentInput {
  body: string
  is_internal?: boolean
}
