import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Interaction, InteractionInput } from '../model/types'

export const interactionApi = {
  list: (customerId: string, page: number, limit: number): Promise<PaginatedResponse<Interaction>> =>
    http(`/api/v1/customers/${customerId}/interactions?${listParams(page, limit)}`),
  create: (customerId: string, body: InteractionInput): Promise<{ data: Interaction }> =>
    http(`/api/v1/customers/${customerId}/interactions`, { method: 'POST', body }),
}
