export type UUID = string

export interface ApiMeta {
  page: number
  limit: number
  total: number
}

export interface PaginatedResponse<T> {
  data: T[]
  meta?: ApiMeta
  pagination?: ApiMeta
}

export interface ApiErrorBody {
  code?: string
  message: string
}

export interface ApiErrorResponse {
  error: string | ApiErrorBody
}

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

export interface HttpOptions {
  method?: HttpMethod
  body?: unknown
}

export interface LineItem {
  description: string
  qty: number
  unit_price: number
  amount: number
}

export interface UserBrief {
  id: UUID
  email?: string
  full_name: string
  role?: string
}

export interface LoginResponse {
  user: UserBrief
}

export function parseMeta(res: PaginatedResponse<unknown>, page: number, limit: number): ApiMeta {
  return res.meta ?? res.pagination ?? { page, limit, total: 0 }
}

export function getErrorMessage(json: ApiErrorResponse, fallback: string): string {
  if (typeof json.error === 'string') return json.error
  return json.error?.message ?? fallback
}

export function cellValue(row: Record<string, unknown>, key: string): unknown {
  return row[key]
}
