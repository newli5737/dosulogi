import { useCallback, useMemo, useState } from 'react'
import { shipmentApi } from '@/entities/shipment/api/shipmentApi'
import type { Shipment } from '@/entities/shipment/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { ShipmentModal } from '@/features/shipment-modal/ui/ShipmentModal'
import { ShipmentDetailModal } from '@/features/shipment-detail-modal/ui/ShipmentDetailModal'

export function ShipmentTable() {
  const [open, setOpen] = useState(false)
  const [detailId, setDetailId] = useState<string | null>(null)

  const columns = useMemo<DataTableColumn<Shipment>[]>(() => [
    { key: 'tracking_code', label: 'Mã vận đơn' },
    { key: 'status', label: 'Trạng thái', render: (r) => r.status || '—' },
    { key: 'origin', label: 'Điểm đi', render: (r) => r.origin || '—' },
    { key: 'destination', label: 'Điểm đến', render: (r) => r.destination || '—' },
    { key: 'estimated_delivery', label: 'ETA', render: (r) => r.estimated_delivery ? r.estimated_delivery.slice(0, 10) : '—' },
    {
      key: '_actions', label: '', render: (r) => (
        <Button variant="secondary" onClick={() => setDetailId(r.id)}>Chi tiết</Button>
      ),
    },
  ], [])

  const fetchPage = useCallback(
    (page: number, limit: number) => shipmentApi.list(page, limit),
    [],
  )
  const { rows, meta, page, setPage, loading, reload } = usePaginated<Shipment>(fetchPage)

  return (
    <>
      <div className="page-header">
        <h1>Vận đơn</h1>
        <Button variant="primary" onClick={() => setOpen(true)}>+ Thêm vận đơn</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <ShipmentModal open={open} onClose={() => setOpen(false)} onSaved={reload} />
      <ShipmentDetailModal
        open={detailId !== null}
        shipmentId={detailId}
        onClose={() => setDetailId(null)}
        onSynced={reload}
      />
    </>
  )
}
