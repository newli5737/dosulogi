import type { LineItem, UUID } from '@/shared/api/types'

export interface Quotation {
  id: UUID
  code: string
  customer_id: UUID
  items: LineItem[]
  total?: number | null
  currency: string
  valid_until?: string | null
  status: string
}

export interface QuotationInput {
  customer_id: UUID
  items: LineItem[]
  currency?: string
  valid_until?: string | null
  status?: string
}
