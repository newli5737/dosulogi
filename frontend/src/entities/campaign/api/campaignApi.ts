import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Campaign, CampaignInput } from '../model/types'

export const campaignApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Campaign>> =>
    http(`/api/v1/campaigns?${listParams(page, limit, filters)}`, { token }),
  create: (token: string, body: CampaignInput): Promise<Campaign> =>
    http('/api/v1/campaigns', { token, method: 'POST', body }),
  update: (token: string, id: string, body: CampaignInput): Promise<Campaign> =>
    http(`/api/v1/campaigns/${id}`, { token, method: 'PUT', body }),
  send: (token: string, id: string): Promise<{ message: string }> =>
    http(`/api/v1/campaigns/${id}/send`, { token, method: 'POST' }),
}
