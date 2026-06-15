import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Opportunity, OpportunityInput } from '../model/types'

export const opportunityApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Opportunity>> =>
    http(`/api/v1/opportunities?${listParams(page, limit, filters)}`, { token }),
  create: (token: string, body: OpportunityInput): Promise<Opportunity> =>
    http('/api/v1/opportunities', { token, method: 'POST', body }),
  update: (token: string, id: string, body: Partial<OpportunityInput>): Promise<Opportunity> =>
    http(`/api/v1/opportunities/${id}`, { token, method: 'PUT', body }),
}
