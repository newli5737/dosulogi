import type { UUID } from '@/shared/api/types'

export interface OpportunityCustomer {
  id: UUID
  name: string
  code: string
}

export interface Opportunity {
  id: UUID
  code?: string
  customer_id: UUID
  title: string
  stage: string
  value?: number | null
  currency: string
  expected_close?: string | null
  lost_reason?: string | null
  note?: string | null
  customer?: OpportunityCustomer
  shipment_ids?: UUID[]
}

export interface StageHistoryEntry {
  id: UUID
  from_stage?: string | null
  to_stage: string
  note?: string | null
  changed_at: string
  changer_name?: string | null
}

export interface OpportunityInput {
  customer_id: UUID
  title: string
  stage: string
  currency?: string
  value?: number | null
  expected_close?: string | null
  lost_reason?: string | null
  note?: string | null
  shipment_ids?: UUID[]
}
