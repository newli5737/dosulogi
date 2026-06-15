import type { UUID } from '@/shared/api/types'

export interface Shipment {
  id: UUID
  tracking_code: string
  status?: string | null
  origin?: string | null
  destination?: string | null
  estimated_delivery?: string | null
  actual_delivery?: string | null
  last_synced_at?: string | null
  customer_id?: UUID | null
  lat?: number | null
  lng?: number | null
}

export interface ShipmentEvent {
  id: UUID
  shipment_id: UUID
  status?: string | null
  description?: string | null
  location?: string | null
  event_time?: string | null
  created_at: string
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
