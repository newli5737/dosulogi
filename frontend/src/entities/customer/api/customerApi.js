import { http, listParams } from '../../../shared/api/http'

export const customerApi = {
  list: (token, page, limit, filters = {}) =>
    http(`/api/v1/customers?${listParams(page, limit, filters)}`, { token }),
  get: (token, id) => http(`/api/v1/customers/${id}`, { token }),
  create: (token, body) => http('/api/v1/customers', { token, method: 'POST', body }),
  update: (token, id, body) => http(`/api/v1/customers/${id}`, { token, method: 'PUT', body }),
  remove: (token, id) => http(`/api/v1/customers/${id}`, { token, method: 'DELETE' }),
}
