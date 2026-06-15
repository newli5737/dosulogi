import type { UserBrief, UUID } from '@/shared/api/types'

export interface Customer {
  id: UUID
  code: string
  name: string
  type: 'B2B' | 'B2C'
  email?: string | null
  phone?: string | null
  province?: string | null
  tax_code?: string | null
  segment: string
  tier: string
  assigned_to?: UserBrief | null
  created_at?: string
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
