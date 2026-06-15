import type { UUID } from '@/shared/api/types'

export interface Shipment {
  id: UUID
  tracking_code: string
  status?: string | null
  origin?: string | null
  destination?: string | null
  estimated_delivery?: string | null
  customer_id?: UUID | null
}

export interface ShipmentInput {
  tracking_code: string
  customer_id?: UUID | null
  origin?: string | null
  destination?: string | null
}

export interface MapPoint {
  tracking_code: string
  status: string
  lat: number
  lng: number
  customer_name: string
  destination: string
}
