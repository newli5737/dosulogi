import './pagination.css'

export function Pagination({ page, limit, total, onChange }) {
  const pages = Math.max(1, Math.ceil(total / limit))
  if (total <= limit) return null
  return (
    <div className="pagination">
      <span className="pagination-info">
        Trang {page}/{pages} · {total} bản ghi
      </span>
      <div className="pagination-actions">
        <button type="button" disabled={page <= 1} onClick={() => onChange(page - 1)}>Trước</button>
        <button type="button" disabled={page >= pages} onClick={() => onChange(page + 1)}>Sau</button>
      </div>
    </div>
  )
}
