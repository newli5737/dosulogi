import type { ReactNode } from 'react'

export interface DataTableColumn<T> {
  key: string
  label: string
  render?: (row: T) => ReactNode
}

export interface DataTableProps<T extends { id?: string }> {
  columns: DataTableColumn<T>[]
  rows: T[]
  loading?: boolean
  empty?: string
}

export function formatCell(value: unknown): ReactNode {
  if (value === null || value === undefined || value === '') return '—'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}
