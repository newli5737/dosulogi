import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Interaction, InteractionInput } from '../model/types'

export const interactionApi = {
  list: (token: string, customerId: string, page: number, limit: number): Promise<PaginatedResponse<Interaction>> =>
    http(`/api/v1/customers/${customerId}/interactions?${listParams(page, limit)}`, { token }),
  create: (token: string, customerId: string, body: InteractionInput): Promise<{ data: Interaction }> =>
    http(`/api/v1/customers/${customerId}/interactions`, { token, method: 'POST', body }),
}
