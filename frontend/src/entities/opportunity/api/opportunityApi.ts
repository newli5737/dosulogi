import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Opportunity, OpportunityInput, StageHistoryEntry } from '../model/types'

export const opportunityApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Opportunity>> =>
    http(`/api/v1/opportunities?${listParams(page, limit, filters)}`, { token }),
  get: (token: string, id: string): Promise<{ data: Opportunity }> =>
    http(`/api/v1/opportunities/${id}`, { token }),
  create: (token: string, body: OpportunityInput): Promise<{ data: Opportunity }> =>
    http('/api/v1/opportunities', { token, method: 'POST', body }),
  update: (token: string, id: string, body: Partial<OpportunityInput>): Promise<{ data: Opportunity }> =>
    http(`/api/v1/opportunities/${id}`, { token, method: 'PUT', body }),
  stageHistory: (token: string, id: string): Promise<{ data: StageHistoryEntry[] }> =>
    http(`/api/v1/opportunities/${id}/stage-history`, { token }),
}
