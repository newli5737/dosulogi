import { http, listParams } from '@/shared/api/http'
import type { PaginatedResponse } from '@/shared/api/types'
import type { MapPoint, Shipment, ShipmentInput } from '../model/types'

export const shipmentApi = {
  list: (token: string, page: number, limit: number, filters: Record<string, string | undefined> = {}): Promise<PaginatedResponse<Shipment>> =>
    http(`/api/v1/shipments?${listParams(page, limit, filters)}`, { token }),
  create: (token: string, body: ShipmentInput): Promise<Shipment> =>
    http('/api/v1/shipments', { token, method: 'POST', body }),
  map: (token: string): Promise<MapPoint[]> =>
    http('/api/v1/dashboard/shipment-map', { token }),
}
