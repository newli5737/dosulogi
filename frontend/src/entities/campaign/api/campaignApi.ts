import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Campaign, CampaignInput, CampaignLog } from '../model/types'

export const campaignApi = {
  list: (page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Campaign>> =>
    http(`/api/v1/campaigns?${listParams(page, limit, filters)}`),
  create: (body: CampaignInput): Promise<Campaign> =>
    http('/api/v1/campaigns', { method: 'POST', body }),
  update: (id: string, body: CampaignInput): Promise<Campaign> =>
    http(`/api/v1/campaigns/${id}`, { method: 'PUT', body }),
  send: (id: string): Promise<{ message: string }> =>
    http(`/api/v1/campaigns/${id}/send`, { method: 'POST' }),
  schedule: (id: string, scheduled_at: string): Promise<Campaign> =>
    http(`/api/v1/campaigns/${id}/schedule`, { method: 'POST', body: { scheduled_at } }),
  logs: (id: string, page: number, limit: number): Promise<PaginatedResponse<CampaignLog>> =>
    http(`/api/v1/campaigns/${id}/logs?${listParams(page, limit)}`),
}
