import type { UUID } from '@/shared/api/types'

export interface Opportunity {
  id: UUID
  customer_id: UUID
  title: string
  stage: string
  value?: number | null
  currency: string
  expected_close?: string | null
  lost_reason?: string | null
}

export interface OpportunityInput {
  customer_id: UUID
  title: string
  stage: string
  currency?: string
  value?: number | null
  expected_close?: string | null
  lost_reason?: string | null
}
