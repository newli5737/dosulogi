import { http, listParams } from '../../../shared/api/http'

export const quotationApi = {
  list: (token, page, limit, filters = {}) =>
    http(`/api/v1/quotations?${listParams(page, limit, filters)}`, { token }),
  get: (token, id) => http(`/api/v1/quotations/${id}`, { token }),
  create: (token, body) => http('/api/v1/quotations', { token, method: 'POST', body }),
  update: (token, id, body) => http(`/api/v1/quotations/${id}`, { token, method: 'PUT', body }),
  send: (token, id) => http(`/api/v1/quotations/${id}/send`, { token, method: 'POST' }),
  convert: (token, id) => http(`/api/v1/quotations/${id}/convert`, { token, method: 'POST' }),
}
