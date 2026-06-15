import { http, listParams } from '../../../shared/api/http'

export const shipmentApi = {
  list: (token, page, limit, filters = {}) =>
    http(`/api/v1/shipments?${listParams(page, limit, filters)}`, { token }),
  get: (token, id) => http(`/api/v1/shipments/${id}`, { token }),
  create: (token, body) => http('/api/v1/shipments', { token, method: 'POST', body }),
  map: (token) => http('/api/v1/dashboard/shipment-map', { token }),
}
