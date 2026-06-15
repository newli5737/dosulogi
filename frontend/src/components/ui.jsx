export function DataTable({ columns, rows, empty = 'Không có dữ liệu' }) {
  if (!rows?.length) return <p className="muted">{empty}</p>
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>{columns.map((c) => <th key={c.key}>{c.label}</th>)}</tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={row.id || i}>
              {columns.map((c) => <td key={c.key}>{c.render ? c.render(row) : row[c.key] ?? '—'}</td>)}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function Page({ title, children, actions }) {
  return (
    <div className="card page">
      <div className="page-head">
        <h2>{title}</h2>
        {actions}
      </div>
      {children}
    </div>
  )
}

export function useToken() {
  return localStorage.getItem('access_token')
}
