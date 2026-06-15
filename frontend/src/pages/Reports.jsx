import { useEffect, useState } from 'react'
import { getARReport, getRevenueReport, asArray } from '../api'
import { DataTable, Page, useToken } from '../components/ui'

export default function ReportsPage() {
  const token = useToken()
  const [revenue, setRevenue] = useState([])
  const [ar, setAr] = useState([])

  useEffect(() => {
    getRevenueReport(token).then((d) => setRevenue(asArray(d))).catch(console.error)
    getARReport(token).then((d) => setAr(asArray(d))).catch(console.error)
  }, [token])

  return (
    <Page title="Báo cáo kế toán">
      <h3>Doanh thu theo tháng</h3>
      <DataTable columns={[
        { key: 'label', label: 'Kỳ' },
        { key: 'amount', label: 'Doanh thu', render: (r) => `${Number(r.amount).toLocaleString('vi-VN')} ₫` },
      ]} rows={revenue} empty="Chưa có doanh thu" />

      <h3 className="section-title">Công nợ (AR)</h3>
      <DataTable columns={[
        { key: 'customer_name', label: 'Khách hàng' },
        { key: 'invoice_count', label: 'Số HĐ' },
        { key: 'total_due', label: 'Còn nợ', render: (r) => `${Number(r.total_due).toLocaleString('vi-VN')} ₫` },
      ]} rows={ar} empty="Không có công nợ" />
    </Page>
  )
}
