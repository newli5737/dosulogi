import type { UUID } from '@/shared/api/types'

export interface Campaign {
  id: UUID
  name: string
  type: string
  status: string
  subject?: string | null
  body_html?: string | null
  sent_count: number
}

export interface CampaignInput {
  name: string
  type: string
  subject?: string | null
  body_html?: string | null
}
