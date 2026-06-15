import type { UserBrief, UUID } from '@/shared/api/types'
import type { Contact } from '@/entities/contact/model/types'

export interface Customer {
  id: UUID
  code: string
  name: string
  type: 'B2B' | 'B2C'
  email?: string | null
  phone?: string | null
  address?: string | null
  province?: string | null
  tax_code?: string | null
  segment: string
  tier: string
  assigned_to?: UserBrief | null
  last_contact_at?: string | null
  created_at?: string
}

export interface CustomerDetail extends Customer {
  primary_contact?: Contact | null
  open_tickets: number
  active_contracts: number
}

export interface CreateCustomerInput {
  name: string
  type: 'B2B' | 'B2C'
  email?: string | null
  phone?: string | null
  province?: string | null
  tax_code?: string | null
  segment: string
  tier: string
}
