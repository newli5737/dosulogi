import './pagination.css'

interface PaginationProps {
  page: number
  limit: number
  total: number
  onChange: (page: number) => void
}

export function Pagination({ page, limit, total, onChange }: PaginationProps) {
  const pages = Math.max(1, Math.ceil(total / limit))
  return (
    <div className="pagination">
      <button type="button" disabled={page <= 1} onClick={() => onChange(page - 1)}>←</button>
      <span>Trang {page} / {pages} ({total} bản ghi)</span>
      <button type="button" disabled={page >= pages} onClick={() => onChange(page + 1)}>→</button>
    </div>
  )
}
