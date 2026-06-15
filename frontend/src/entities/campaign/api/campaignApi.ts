import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Campaign, CampaignInput, CampaignLog } from '../model/types'

export const campaignApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Campaign>> =>
    http(`/api/v1/campaigns?${listParams(page, limit, filters)}`, { token }),
  create: (token: string, body: CampaignInput): Promise<Campaign> =>
    http('/api/v1/campaigns', { token, method: 'POST', body }),
  update: (token: string, id: string, body: CampaignInput): Promise<Campaign> =>
    http(`/api/v1/campaigns/${id}`, { token, method: 'PUT', body }),
  send: (token: string, id: string): Promise<{ message: string }> =>
    http(`/api/v1/campaigns/${id}/send`, { token, method: 'POST' }),
  schedule: (token: string, id: string, scheduled_at: string): Promise<Campaign> =>
    http(`/api/v1/campaigns/${id}/schedule`, { token, method: 'POST', body: { scheduled_at } }),
  logs: (token: string, id: string, page: number, limit: number): Promise<PaginatedResponse<CampaignLog>> =>
    http(`/api/v1/campaigns/${id}/logs?${listParams(page, limit)}`, { token }),
}
