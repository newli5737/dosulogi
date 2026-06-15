import type { LineItem, UUID } from '@/shared/api/types'

export type InvoiceStatus = 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled'

export interface Invoice {
  id: UUID
  code: string
  customer_id: UUID
  items?: LineItem[]
  subtotal?: number | null
  tax_rate: number
  tax_amount?: number | null
  total?: number | null
  currency: string
  status: InvoiceStatus
  due_date?: string | null
  file_url?: string | null
}

export interface InvoiceInput {
  customer_id: UUID
  contract_id?: UUID | null
  tax_rate?: number
  currency?: string
  status?: InvoiceStatus
  due_date?: string | null
}

export interface CreateInvoicePayload {
  invoice: InvoiceInput
  items: LineItem[]
}
