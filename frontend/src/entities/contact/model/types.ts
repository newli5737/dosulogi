import type { UUID } from '@/shared/api/types'

export interface Contact {
  id: UUID
  customer_id: UUID
  name: string
  role?: string | null
  phone?: string | null
  email?: string | null
  is_primary: boolean
  note?: string | null
}

export interface ContactInput {
  name: string
  role?: string | null
  phone?: string | null
  email?: string | null
  is_primary?: boolean
  note?: string | null
}
