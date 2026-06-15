import { http, listParams } from '../../../shared/api/http'

export const opportunityApi = {
  list: (token, page, limit, filters = {}) =>
    http(`/api/v1/opportunities?${listParams(page, limit, filters)}`, { token }),
  get: (token, id) => http(`/api/v1/opportunities/${id}`, { token }),
  create: (token, body) => http('/api/v1/opportunities', { token, method: 'POST', body }),
  update: (token, id, body) => http(`/api/v1/opportunities/${id}`, { token, method: 'PUT', body }),
  remove: (token, id) => http(`/api/v1/opportunities/${id}`, { token, method: 'DELETE' }),
}
