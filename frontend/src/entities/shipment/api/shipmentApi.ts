import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { MapPoint, Shipment, ShipmentEvent, ShipmentInput } from '../model/types'

export const shipmentApi = {
  list: (page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Shipment>> =>
    http(`/api/v1/shipments?${listParams(page, limit, filters)}`),
  get: (id: string): Promise<Shipment> =>
    http(`/api/v1/shipments/${id}`),
  create: (body: ShipmentInput): Promise<Shipment> =>
    http('/api/v1/shipments', { method: 'POST', body }),
  events: (id: string): Promise<ShipmentEvent[]> =>
    http(`/api/v1/shipments/${id}/events`),
  sync: (id: string): Promise<{ message?: string }> =>
    http(`/api/v1/shipments/${id}/sync`, { method: 'POST' }),
  map: (): Promise<MapPoint[]> =>
    http('/api/v1/dashboard/shipment-map'),
}
