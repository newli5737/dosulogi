import { useEffect, useState } from 'react'
import { dashboardApi, type DashboardSummary, type FunnelStage } from '@/entities/session/api/sessionApi'
import './dashboard-page.css'

interface KpiCard {
  label: string
  value: string | number
  tone: string
}

export function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [funnel, setFunnel] = useState<FunnelStage[]>([])

  useEffect(() => {
    dashboardApi.summary().then(setSummary).catch(console.error)
    dashboardApi.funnel().then((d) => setFunnel(Array.isArray(d) ? d : [])).catch(console.error)
  }, [])

  const cards: KpiCard[] = summary ? [
    { label: 'Doanh thu', value: `${Number(summary.revenue).toLocaleString('vi-VN')} ₫`, tone: 'blue' },
    { label: 'Vận đơn', value: summary.shipment_count, tone: 'cyan' },
    { label: 'KH mới (tháng)', value: summary.new_customers, tone: 'green' },
    { label: 'Công nợ', value: `${Number(summary.total_ar).toLocaleString('vi-VN')} ₫`, tone: 'amber' },
  ] : []

  const funnelData = funnel.length ? funnel : [{ stage: 'lead', count: 0, value: 0 }]

  return (
    <div className="page-card">
      <div className="page-header"><h1>Dashboard</h1></div>
      <div className="dash-grid">
        {cards.map((c) => (
          <div key={c.label} className={`dash-kpi dash-kpi--${c.tone}`}>
            <span>{c.label}</span>
            <strong>{c.value}</strong>
          </div>
        ))}
      </div>
      <h2 className="dash-section">Sales pipeline</h2>
      <div className="dash-funnel">
        {funnelData.map((f) => (
          <div key={f.stage} className="dash-funnel-item">
            <div className="dash-funnel-bar" style={{ width: `${Math.min(100, (f.count || 0) * 10 + 20)}%` }} />
            <span className="dash-funnel-label">{f.stage}</span>
            <span className="dash-funnel-meta">{f.count} · {Number(f.value || 0).toLocaleString('vi-VN')} ₫</span>
          </div>
        ))}
      </div>
    </div>
  )
}
