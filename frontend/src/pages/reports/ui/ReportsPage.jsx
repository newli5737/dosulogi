import { useEffect, useState } from 'react'
import { reportApi } from '../../../entities/session/api/sessionApi'
import { useToken } from '../../../app/providers/AuthProvider'
import { DataTable } from '../../../shared/ui/DataTable/DataTable'

export function ReportsPage() {
  const token = useToken()
  const [revenue, setRevenue] = useState([])
  const [ar, setAr] = useState([])

  useEffect(() => {
    reportApi.revenue(token).then((d) => setRevenue(Array.isArray(d) ? d : d?.data || [])).catch(console.error)
    reportApi.ar(token).then((d) => setAr(Array.isArray(d) ? d : d?.data || [])).catch(console.error)
  }, [token])

  return (
    <div className="page-card">
      <div className="page-header"><h1>Báo cáo kế toán</h1></div>
      <h2 className="section-title">Doanh thu theo tháng</h2>
      <DataTable columns={[
        { key: 'label', label: 'Kỳ' },
        { key: 'amount', label: 'Doanh thu', render: (r) => `${Number(r.amount || 0).toLocaleString('vi-VN')} ₫` },
      ]} rows={revenue} empty="Chưa có doanh thu" />
      <h2 className="section-title">Công nợ (AR)</h2>
      <DataTable columns={[
        { key: 'customer_name', label: 'Khách hàng' },
        { key: 'invoice_count', label: 'Số HĐ' },
        { key: 'total_due', label: 'Còn nợ', render: (r) => `${Number(r.total_due || 0).toLocaleString('vi-VN')} ₫` },
      ]} rows={ar} empty="Không có công nợ" />
    </div>
  )
}
