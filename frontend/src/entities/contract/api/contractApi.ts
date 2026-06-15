import { http, httpForm, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { Contract, ContractInput } from '../model/types'

export const contractApi = {
  list: (page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Contract>> =>
    http(`/api/v1/contracts?${listParams(page, limit, filters)}`),
  create: (body: ContractInput): Promise<Contract> =>
    http('/api/v1/contracts', { method: 'POST', body }),
  update: (id: string, body: Partial<ContractInput>): Promise<Contract> =>
    http(`/api/v1/contracts/${id}`, { method: 'PUT', body }),
  upload: (id: string, file: File): Promise<Contract> => {
    const fd = new FormData()
    fd.append('file', file)
    return httpForm(`/api/v1/contracts/${id}/upload`, fd)
  },
}
