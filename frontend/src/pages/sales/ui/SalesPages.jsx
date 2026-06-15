import { createListPage } from '../../_shared/createListPage'
import { invoiceApi, paymentApi } from '../../../entities/session/api/sessionApi'

export const InvoicesPage = createListPage({
  title: 'Hóa đơn',
  fetchList: invoiceApi.list,
  columns: [
    { key: 'code', label: 'Mã' },
    { key: 'status', label: 'Trạng thái' },
    { key: 'total', label: 'Tổng', render: (r) => r.total ? `${Number(r.total).toLocaleString('vi-VN')} ₫` : '—' },
  ],
})

export const PaymentsPage = createListPage({
  title: 'Thanh toán',
  fetchList: paymentApi.list,
  columns: [
    { key: 'amount', label: 'Số tiền', render: (r) => r.amount ? `${Number(r.amount).toLocaleString('vi-VN')} ₫` : '—' },
    { key: 'method', label: 'Phương thức' },
    { key: 'reference_code', label: 'Mã CK' },
  ],
})
