import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { CreateUserInput, UpdateUserInput, User } from '../model/types'

export const userApi = {
  list: (page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<User>> =>
    http(`/api/v1/users?${listParams(page, limit, filters)}`),
  create: (body: CreateUserInput): Promise<User> =>
    http('/api/v1/users', { method: 'POST', body }),
  update: (id: string, body: UpdateUserInput): Promise<User> =>
    http(`/api/v1/users/${id}`, { method: 'PUT', body }),
  deactivate: (id: string): Promise<void> =>
    http(`/api/v1/users/${id}`, { method: 'DELETE' }),
}
