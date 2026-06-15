import { http, listParams } from '../../../shared/api/http'

export const ticketApi = {
  list: (token, page, limit, filters = {}) =>
    http(`/api/v1/tickets?${listParams(page, limit, filters)}`, { token }),
  create: (token, body) => http('/api/v1/tickets', { token, method: 'POST', body }),
  update: (token, id, body) => http(`/api/v1/tickets/${id}`, { token, method: 'PUT', body }),
}
