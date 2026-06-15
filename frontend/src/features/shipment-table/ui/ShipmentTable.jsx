import { useCallback, useMemo, useState } from 'react'
import { shipmentApi } from '../../../entities/shipment/api/shipmentApi'
import { usePaginated } from '../../../shared/hooks/usePaginated'
import { useToken } from '../../../app/providers/AuthProvider'
import { DataTable } from '../../../shared/ui/DataTable/DataTable'
import { Pagination } from '../../../shared/ui/Pagination/Pagination'
import { Button } from '../../../shared/ui/Button/Button'
import { ShipmentModal } from '../../shipment-modal/ui/ShipmentModal'

export function ShipmentTable() {
  const token = useToken()
  const [open, setOpen] = useState(false)

  const columns = useMemo(() => [
    { key: 'tracking_code', label: 'Mã vận đơn' },
    { key: 'status', label: 'Trạng thái', render: (r) => r.status || '—' },
    { key: 'origin', label: 'Điểm đi', render: (r) => r.origin || '—' },
    { key: 'destination', label: 'Điểm đến', render: (r) => r.destination || '—' },
    { key: 'estimated_delivery', label: 'ETA', render: (r) => r.estimated_delivery ? r.estimated_delivery.slice(0, 10) : '—' },
  ], [])

  const fetchPage = useCallback((page, limit) => shipmentApi.list(token, page, limit), [token])
  const { rows, meta, page, setPage, loading, reload } = usePaginated(fetchPage)

  return (
    <>
      <div className="page-header">
        <h1>Vận đơn</h1>
        <Button variant="primary" onClick={() => setOpen(true)}>+ Thêm vận đơn</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <ShipmentModal open={open} onClose={() => setOpen(false)} onSaved={reload} />
    </>
  )
}
