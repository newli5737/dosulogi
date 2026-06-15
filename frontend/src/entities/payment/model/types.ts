import type { UUID } from '@/shared/api/types'

export interface Payment {
  id: UUID
  invoice_id: UUID
  amount?: number | null
  method?: string | null
  reference_code?: string | null
  note?: string | null
  created_at?: string
}

export interface PaymentInput {
  invoice_id: UUID
  amount: number
  method?: string | null
  reference_code?: string | null
  note?: string | null
}
