import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { MapPoint, Shipment, ShipmentEvent, ShipmentInput } from '../model/types'

export const shipmentApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Shipment>> =>
    http(`/api/v1/shipments?${listParams(page, limit, filters)}`, { token }),
  get: (token: string, id: string): Promise<Shipment> =>
    http(`/api/v1/shipments/${id}`, { token }),
  create: (token: string, body: ShipmentInput): Promise<Shipment> =>
    http('/api/v1/shipments', { token, method: 'POST', body }),
  events: (token: string, id: string): Promise<ShipmentEvent[]> =>
    http(`/api/v1/shipments/${id}/events`, { token }),
  sync: (token: string, id: string): Promise<{ message?: string }> =>
    http(`/api/v1/shipments/${id}/sync`, { token, method: 'POST' }),
  map: (token: string): Promise<MapPoint[]> =>
    http('/api/v1/dashboard/shipment-map', { token }),
}
