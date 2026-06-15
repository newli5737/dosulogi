import type { LineItem, UUID } from '@/shared/api/types'

export interface Quotation {
  id: UUID
  code: string
  customer_id: UUID
  opportunity_id?: UUID | null
  opp_id?: UUID | null
  items: LineItem[]
  subtotal?: number | null
  discount?: number | null
  tax_rate?: number | null
  tax_amount?: number | null
  total?: number | null
  currency: string
  valid_until?: string | null
  status: string
  note?: string | null
}

export interface QuotationInput {
  customer_id: UUID
  opportunity_id?: UUID | null
  items: LineItem[]
  currency?: string
  valid_until?: string | null
  status?: string
  discount?: number | null
  tax_rate?: number | null
  note?: string | null
}
