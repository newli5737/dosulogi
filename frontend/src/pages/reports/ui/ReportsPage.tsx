import { useEffect, useState } from 'react'
import { reportApi, type ARReportRow, type RevenueReportRow } from '@/entities/session/api/sessionApi'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Button } from '@/shared/ui/Button/Button'

export function ReportsPage() {
  const [revenue, setRevenue] = useState<RevenueReportRow[]>([])
  const [ar, setAr] = useState<ARReportRow[]>([])
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')

  function load() {
    const fromIso = from ? new Date(from).toISOString() : undefined
    const toIso = to ? new Date(to).toISOString() : undefined
    reportApi.revenue(fromIso, toIso).then((d) => setRevenue(Array.isArray(d) ? d : [])).catch(console.error)
    reportApi.ar().then((d) => setAr(Array.isArray(d) ? d : [])).catch(console.error)
  }

  useEffect(() => { load() }, [])

  const revenueColumns: DataTableColumn<RevenueReportRow>[] = [
    { key: 'label', label: 'Kỳ' },
    { key: 'amount', label: 'Doanh thu', render: (r) => `${Number(r.amount || 0).toLocaleString('vi-VN')} ₫` },
  ]

  const arColumns: DataTableColumn<ARReportRow>[] = [
    { key: 'customer_name', label: 'Khách hàng' },
    { key: 'invoice_count', label: 'Số HĐ' },
    { key: 'total_due', label: 'Còn nợ', render: (r) => `${Number(r.total_due || 0).toLocaleString('vi-VN')} ₫` },
  ]

  return (
    <div className="page-card">
      <div className="page-header"><h1>Báo cáo kế toán</h1></div>
      <div className="toolbar" style={{ marginBottom: 16, display: 'flex', gap: 12, alignItems: 'flex-end' }}>
        <label>Từ <input className="field-input" type="date" value={from} onChange={(e) => setFrom(e.target.value)} /></label>
        <label>Đến <input className="field-input" type="date" value={to} onChange={(e) => setTo(e.target.value)} /></label>
        <Button variant="primary" onClick={load}>Lọc doanh thu</Button>
      </div>
      <h2 className="section-title">Doanh thu theo tháng</h2>
      <DataTable<RevenueReportRow & { id?: string }> columns={revenueColumns} rows={revenue} empty="Chưa có doanh thu" />
      <h2 className="section-title">Công nợ (AR)</h2>
      <DataTable<ARReportRow & { id?: string }> columns={arColumns} rows={ar} empty="Không có công nợ" />
    </div>
  )
}
