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
  customer?: { name: string }
  is_overdue?: boolean
}

export interface CreateTicketInput {
  customer_id: UUID
  title: string
  description?: string | null
  priority: string
  category?: string | null
}
