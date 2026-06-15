import { http, listParams } from '../../../shared/api/http'

export const userApi = {
  list: (token, page, limit, filters = {}) =>
    http(`/api/v1/users?${listParams(page, limit, filters)}`, { token }),
  create: (token, body) => http('/api/v1/users', { token, method: 'POST', body }),
  update: (token, id, body) => http(`/api/v1/users/${id}`, { token, method: 'PUT', body }),
}
