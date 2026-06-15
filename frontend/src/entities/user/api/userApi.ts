import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { CreateUserInput, UpdateUserInput, User } from '../model/types'

export const userApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<User>> =>
    http(`/api/v1/users?${listParams(page, limit, filters)}`, { token }),
  create: (token: string, body: CreateUserInput): Promise<User> =>
    http('/api/v1/users', { token, method: 'POST', body }),
  update: (token: string, id: string, body: UpdateUserInput): Promise<User> =>
    http(`/api/v1/users/${id}`, { token, method: 'PUT', body }),
  deactivate: (token: string, id: string): Promise<void> =>
    http(`/api/v1/users/${id}`, { token, method: 'DELETE' }),
}
