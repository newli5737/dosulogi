import { http, listParams } from '../../../shared/api/http'

export const campaignApi = {
  list: (token, page, limit, filters = {}) =>
    http(`/api/v1/campaigns?${listParams(page, limit, filters)}`, { token }),
  get: (token, id) => http(`/api/v1/campaigns/${id}`, { token }),
  create: (token, body) => http('/api/v1/campaigns', { token, method: 'POST', body }),
  update: (token, id, body) => http(`/api/v1/campaigns/${id}`, { token, method: 'PUT', body }),
  send: (token, id) => http(`/api/v1/campaigns/${id}/send`, { token, method: 'POST' }),
}
