import { useEffect, useState } from 'react'
import { getSalesFunnel, getSummary, asArray } from '../api'
import { Page, useToken } from '../components/ui'

export default function DashboardPage() {
  const token = useToken()
  const [summary, setSummary] = useState(null)
  const [funnel, setFunnel] = useState([])

  useEffect(() => {
    getSummary(token).then(setSummary).catch(console.error)
    getSalesFunnel(token).then((d) => setFunnel(asArray(d))).catch(console.error)
  }, [token])

  return (
    <Page title="Dashboard">
      {summary && (
        <div className="grid">
          <div className="kpi"><label>Doanh thu</label><strong>{Number(summary.revenue).toLocaleString('vi-VN')} ₫</strong></div>
          <div className="kpi"><label>Shipments</label><strong>{summary.shipment_count}</strong></div>
          <div className="kpi"><label>KH mới (tháng)</label><strong>{summary.new_customers}</strong></div>
          <div className="kpi"><label>Công nợ</label><strong>{Number(summary.total_ar).toLocaleString('vi-VN')} ₫</strong></div>
        </div>
      )}
      <h3 className="section-title">Sales funnel</h3>
      <div className="grid">
        {funnel.map((f) => (
          <div className="kpi" key={f.stage}>
            <label>{f.stage}</label>
            <strong>{f.count}</strong>
            <small>{Number(f.value).toLocaleString('vi-VN')} ₫</small>
          </div>
        ))}
      </div>
    </Page>
  )
}
