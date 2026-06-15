import type { UUID } from '@/shared/api/types'

export interface User {
  id: UUID
  email: string
  full_name: string
  role: string
  is_active: boolean
}

export interface CreateUserInput {
  email: string
  password: string
  full_name: string
  role: string
}
