import { createListPage } from '../../_shared/createListPage'
import { opportunityApi, contractApi, quotationApi, invoiceApi, paymentApi, shipmentApi, campaignApi } from '../../../entities/session/api/sessionApi'

export const OpportunitiesPage = createListPage({
  title: 'Cơ hội bán hàng',
  fetchList: opportunityApi.list,
  columns: [
    { key: 'title', label: 'Tiêu đề' },
    { key: 'stage', label: 'Stage' },
    { key: 'value', label: 'Giá trị', render: (r) => r.value ? `${Number(r.value).toLocaleString('vi-VN')} ₫` : '—' },
  ],
})

export const ContractsPage = createListPage({
  title: 'Hợp đồng',
  fetchList: contractApi.list,
  columns: [
    { key: 'code', label: 'Mã' },
    { key: 'title', label: 'Tiêu đề' },
    { key: 'status', label: 'Trạng thái' },
    { key: 'service_type', label: 'Dịch vụ' },
  ],
})

export const QuotationsPage = createListPage({
  title: 'Báo giá',
  fetchList: quotationApi.list,
  columns: [
    { key: 'code', label: 'Mã' },
    { key: 'status', label: 'Trạng thái' },
    { key: 'total', label: 'Tổng', render: (r) => r.total ? `${Number(r.total).toLocaleString('vi-VN')} ₫` : '—' },
  ],
})

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

export const ShipmentsPage = createListPage({
  title: 'Vận đơn',
  fetchList: shipmentApi.list,
  columns: [
    { key: 'tracking_number', label: 'Mã vận đơn' },
    { key: 'status', label: 'Trạng thái' },
    { key: 'origin', label: 'Điểm đi' },
    { key: 'destination', label: 'Điểm đến' },
  ],
})

export const CampaignsPage = createListPage({
  title: 'Marketing',
  fetchList: campaignApi.list,
  columns: [
    { key: 'name', label: 'Tên chiến dịch' },
    { key: 'status', label: 'Trạng thái' },
    { key: 'channel', label: 'Kênh' },
  ],
})
