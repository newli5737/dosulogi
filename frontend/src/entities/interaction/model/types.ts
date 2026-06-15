import type { UserBrief, UUID } from '@/shared/api/types'

export interface Interaction {
  id: UUID
  customer_id: UUID
  channel: string
  direction?: string | null
  summary: string
  occurred_at: string
  created_by?: UserBrief | null
  created_at: string
}

export interface InteractionInput {
  channel: string
  direction?: string | null
  summary: string
  occurred_at?: string | null
}
