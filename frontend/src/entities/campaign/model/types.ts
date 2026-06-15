import type { UUID } from '@/shared/api/types'

export interface Campaign {
  id: UUID
  name: string
  type: string
  status: string
  subject?: string | null
  body_html?: string | null
  scheduled_at?: string | null
  sent_count: number
}

export interface CampaignInput {
  name: string
  type: string
  subject?: string | null
  body_html?: string | null
}

export interface CampaignLog {
  id: UUID
  campaign_id: UUID
  email?: string | null
  status?: string | null
  created_at: string
}
