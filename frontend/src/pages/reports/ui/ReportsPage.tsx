import { useEffect, useState } from 'react'
import {
  Bar, BarChart, CartesianGrid, Cell, Legend, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis, YAxis,
} from 'recharts'
import { reportApi, type ARReportRow, type RevenueReportRow } from '@/entities/session/api/sessionApi'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Button } from '@/shared/ui/Button/Button'
import '@/pages/dashboard/ui/dashboard-page.css'

const PIE_COLORS = ['#2563eb', '#0891b2', '#16a34a', '#d97706', '#e11d48', '#7c3aed']

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

  const totalRevenue = revenue.reduce((s, r) => s + Number(r.amount || 0), 0)
  const totalAR = ar.reduce((s, r) => s + Number(r.total_due || 0), 0)

  return (
    <div className="page-card">
      <div className="page-header"><h1>Báo cáo kế toán</h1></div>

      <div className="dash-grid" style={{ marginBottom: 24 }}>
        <div className="dash-kpi dash-kpi--blue">
          <span>Tổng doanh thu (kỳ lọc)</span>
          <strong>{totalRevenue.toLocaleString('vi-VN')} ₫</strong>
        </div>
        <div className="dash-kpi dash-kpi--amber">
          <span>Tổng công nợ</span>
          <strong>{totalAR.toLocaleString('vi-VN')} ₫</strong>
        </div>
        <div className="dash-kpi dash-kpi--cyan">
          <span>Số khách nợ</span>
          <strong>{ar.length}</strong>
        </div>
      </div>

      <div className="toolbar" style={{ marginBottom: 16, display: 'flex', gap: 12, alignItems: 'flex-end' }}>
        <label>Từ <input className="field-input" type="date" value={from} onChange={(e) => setFrom(e.target.value)} /></label>
        <label>Đến <input className="field-input" type="date" value={to} onChange={(e) => setTo(e.target.value)} /></label>
        <Button variant="primary" onClick={load}>Lọc doanh thu</Button>
      </div>

      <div className="dash-charts" style={{ marginBottom: 24 }}>
        <section className="dash-chart-card dash-chart-card--wide">
          <h2>Doanh thu theo tháng</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={revenue}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis dataKey="label" />
              <YAxis tickFormatter={(v) => `${(Number(v) / 1e6).toFixed(0)}M`} />
              <Tooltip formatter={(v) => [`${Number(v).toLocaleString('vi-VN')} ₫`, 'Doanh thu']} />
              <Bar dataKey="amount" fill="#2563eb" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </section>

        <section className="dash-chart-card">
          <h2>Cơ cấu công nợ (top KH)</h2>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={ar.slice(0, 6)}
                dataKey="total_due"
                nameKey="customer_name"
                cx="50%"
                cy="50%"
                outerRadius={95}
                label={({ name }) => String(name || '').slice(0, 12)}
              >
                {ar.slice(0, 6).map((_, i) => (
                  <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />
                ))}
              </Pie>
              <Tooltip formatter={(v) => `${Number(v).toLocaleString('vi-VN')} ₫`} />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </section>

        <section className="dash-chart-card">
          <h2>Top công nợ</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={ar.slice(0, 8)} layout="vertical" margin={{ left: 8 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis type="number" tickFormatter={(v) => `${(Number(v) / 1e6).toFixed(0)}M`} />
              <YAxis type="category" dataKey="customer_name" width={100} tick={{ fontSize: 11 }} />
              <Tooltip formatter={(v) => [`${Number(v).toLocaleString('vi-VN')} ₫`, 'Còn nợ']} />
              <Bar dataKey="total_due" fill="#d97706" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </section>
      </div>

      <h2 className="section-title">Doanh thu theo tháng</h2>
      <DataTable<RevenueReportRow & { id?: string }> columns={revenueColumns} rows={revenue} empty="Chưa có doanh thu" />
      <h2 className="section-title">Công nợ (AR)</h2>
      <DataTable<ARReportRow & { id?: string }> columns={arColumns} rows={ar} empty="Không có công nợ" />
    </div>
  )
}
