import { http, listParams } from '../../../shared/api/http'

export const contractApi = {
  list: (token, page, limit, filters = {}) =>
    http(`/api/v1/contracts?${listParams(page, limit, filters)}`, { token }),
  get: (token, id) => http(`/api/v1/contracts/${id}`, { token }),
  create: (token, body) => http('/api/v1/contracts', { token, method: 'POST', body }),
  update: (token, id, body) => http(`/api/v1/contracts/${id}`, { token, method: 'PUT', body }),
}
