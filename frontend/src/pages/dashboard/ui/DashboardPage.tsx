import { useEffect, useState } from 'react'
import {
  Bar, BarChart, CartesianGrid, Cell, Legend, Line, LineChart, Pie, PieChart,
  ResponsiveContainer, Tooltip, XAxis, YAxis,
} from 'recharts'
import { dashboardApi, type DashboardSummary, type FunnelStage, type StatusCount, type TrendPoint } from '@/entities/session/api/sessionApi'
import './dashboard-page.css'

import { opportunityStageLabel, statusCountLabel } from '@/shared/lib/labels'

const CHART_COLORS = ['#2563eb', '#0891b2', '#16a34a', '#d97706', '#e11d48', '#7c3aed']

export function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [funnel, setFunnel] = useState<FunnelStage[]>([])
  const [revenueTrend, setRevenueTrend] = useState<TrendPoint[]>([])
  const [ticketStats, setTicketStats] = useState<StatusCount[]>([])
  const [shipmentStats, setShipmentStats] = useState<StatusCount[]>([])

  useEffect(() => {
    dashboardApi.summary().then(setSummary).catch(console.error)
    dashboardApi.funnel().then((d) => setFunnel(Array.isArray(d) ? d : [])).catch(console.error)
    dashboardApi.revenueTrend().then((d) => setRevenueTrend(Array.isArray(d) ? d : [])).catch(console.error)
    dashboardApi.ticketStats().then((d) => setTicketStats(Array.isArray(d) ? d : [])).catch(console.error)
    dashboardApi.shipmentStats().then((d) => setShipmentStats(Array.isArray(d) ? d : [])).catch(console.error)
  }, [])

  const cards = summary ? [
    { label: 'Doanh thu', value: `${Number(summary.revenue).toLocaleString('vi-VN')} ₫`, tone: 'blue' },
    { label: 'Vận đơn', value: summary.shipment_count, tone: 'cyan' },
    { label: 'KH mới (tháng)', value: summary.new_customers, tone: 'green' },
    { label: 'Công nợ', value: `${Number(summary.total_ar).toLocaleString('vi-VN')} ₫`, tone: 'amber' },
    { label: 'Ticket mở', value: summary.open_tickets, tone: 'rose' },
    { label: 'Cơ hội active', value: summary.active_opportunities, tone: 'violet' },
  ] : []

  const ticketChart = ticketStats.map((s) => ({ ...s, label: statusCountLabel(s.status, 'ticket') }))
  const shipmentChart = shipmentStats.map((s) => ({ ...s, label: statusCountLabel(s.status, 'shipment') }))
  const funnelChart = funnel.map((f) => ({
    name: opportunityStageLabel(f.stage),
    count: f.count,
    value: f.value,
  }))

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

      <div className="dash-charts">
        <section className="dash-chart-card dash-chart-card--wide">
          <h2>Doanh thu 6 tháng</h2>
          <ResponsiveContainer width="100%" height={280}>
            <LineChart data={revenueTrend}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis dataKey="label" tick={{ fontSize: 12 }} />
              <YAxis tickFormatter={(v) => `${(Number(v) / 1e6).toFixed(0)}M`} tick={{ fontSize: 12 }} />
              <Tooltip formatter={(v) => [`${Number(v).toLocaleString('vi-VN')} ₫`, 'Doanh thu']} />
              <Line type="monotone" dataKey="amount" stroke="#2563eb" strokeWidth={2.5} dot={{ r: 4 }} />
            </LineChart>
          </ResponsiveContainer>
        </section>

        <section className="dash-chart-card">
          <h2>Pipeline bán hàng</h2>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={funnelChart} layout="vertical" margin={{ left: 8, right: 16 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis type="number" tick={{ fontSize: 12 }} />
              <YAxis type="category" dataKey="name" width={90} tick={{ fontSize: 12 }} />
              <Tooltip />
              <Bar dataKey="count" fill="#0891b2" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </section>

        <section className="dash-chart-card">
          <h2>Tickets theo trạng thái</h2>
          <ResponsiveContainer width="100%" height={280}>
            <PieChart>
              <Pie data={ticketChart} dataKey="count" nameKey="label" cx="50%" cy="50%" outerRadius={90} label>
                {ticketChart.map((_, i) => (
                  <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </section>

        <section className="dash-chart-card">
          <h2>Vận đơn theo trạng thái</h2>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={shipmentChart}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis dataKey="label" tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 12 }} />
              <Tooltip />
              <Bar dataKey="count" fill="#16a34a" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </section>
      </div>
    </div>
  )
}
