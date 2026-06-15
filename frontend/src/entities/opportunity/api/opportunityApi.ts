import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Opportunity, OpportunityInput, StageHistoryEntry } from '../model/types'

export const opportunityApi = {
  list: (page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Opportunity>> =>
    http(`/api/v1/opportunities?${listParams(page, limit, filters)}`),
  get: (id: string): Promise<{ data: Opportunity }> =>
    http(`/api/v1/opportunities/${id}`),
  create: (body: OpportunityInput): Promise<{ data: Opportunity }> =>
    http('/api/v1/opportunities', { method: 'POST', body }),
  update: (id: string, body: Partial<OpportunityInput>): Promise<{ data: Opportunity }> =>
    http(`/api/v1/opportunities/${id}`, { method: 'PUT', body }),
  stageHistory: (id: string): Promise<{ data: StageHistoryEntry[] }> =>
    http(`/api/v1/opportunities/${id}/stage-history`),
}
