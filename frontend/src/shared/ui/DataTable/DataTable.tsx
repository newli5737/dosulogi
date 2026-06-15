import './table.css'
import type { DataTableProps } from './types'
import { formatCell } from './types'

export function DataTable<T extends { id?: string }>({ columns, rows, loading, empty = 'Không có dữ liệu' }: DataTableProps<T>) {
  if (loading) return <div className="table-loading">Đang tải...</div>
  if (!rows.length) return <div className="table-empty">{empty}</div>
  return (
    <div className="table-wrap">
      <table className="data-table">
        <thead>
          <tr>{columns.map((c) => <th key={c.key}>{c.label}</th>)}</tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={row.id ?? i}>
              {columns.map((c) => (
                <td key={c.key}>{c.render ? c.render(row) : formatCell((row as Record<string, unknown>)[c.key])}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export type { DataTableColumn, DataTableProps } from './types'
