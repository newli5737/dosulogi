import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { CreateUserInput, User } from '../model/types'

export const userApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<User>> =>
    http(`/api/v1/users?${listParams(page, limit, filters)}`, { token }),
  create: (token: string, body: CreateUserInput): Promise<User> =>
    http('/api/v1/users', { token, method: 'POST', body }),
}
