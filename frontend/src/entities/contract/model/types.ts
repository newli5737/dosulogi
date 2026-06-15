import type { UUID } from '@/shared/api/types'

export type ServiceType = 'FCL' | 'LCL' | 'air' | 'express' | 'road' | 'domestic' | 'cold_chain' | 'warehouse' | 'last_mile' | 'sea'
export type ContractStatus = 'draft' | 'active' | 'expired' | 'terminated'

export interface Contract {
  id: UUID
  code: string
  customer_id: UUID
  title?: string | null
  service_type?: ServiceType | null
  start_date: string
  end_date?: string | null
  value?: number | null
  currency: string
  status: ContractStatus
  payment_terms?: string | null
  file_url?: string | null
}

export interface ContractInput {
  customer_id: UUID
  title?: string | null
  service_type?: ServiceType
  start_date: string
  end_date?: string | null
  value?: number | null
  currency?: string
  status?: ContractStatus
  payment_terms?: string | null
}
