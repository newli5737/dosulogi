import { useCallback, useEffect, useState } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Button } from '@/shared/ui/Button/Button'
import { shipmentApi } from '@/entities/shipment/api/shipmentApi'
import { shipmentStatusLabel } from '@/shared/lib/labels'
import type { Shipment, ShipmentEvent } from '@/entities/shipment/model/types'

interface ShipmentDetailModalProps {
  open: boolean
  shipmentId: string | null
  onClose: () => void
  onSynced?: () => void
}

export function ShipmentDetailModal({ open, shipmentId, onClose, onSynced }: ShipmentDetailModalProps) {
  const [shipment, setShipment] = useState<Shipment | null>(null)
  const [events, setEvents] = useState<ShipmentEvent[]>([])
  const [loading, setLoading] = useState(false)
  const [syncing, setSyncing] = useState(false)

  const load = useCallback(async () => {
    if (!shipmentId) return
    setLoading(true)
    try {
      const [s, ev] = await Promise.all([
        shipmentApi.get(shipmentId),
        shipmentApi.events(shipmentId),
      ])
      setShipment(s)
      setEvents(Array.isArray(ev) ? ev : [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [shipmentId])

  useEffect(() => {
    if (open && shipmentId) void load()
  }, [open, shipmentId, load])

  async function sync() {
    if (!shipmentId) return
    setSyncing(true)
    try {
      await shipmentApi.sync(shipmentId)
      await load()
      onSynced?.()
    } catch (e) {
      console.error(e)
    } finally {
      setSyncing(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={shipment ? `Vận đơn ${shipment.tracking_code}` : 'Chi tiết vận đơn'} wide>
      {loading && !shipment ? <p>Đang tải...</p> : shipment && (
        <>
          <dl className="detail-dl" style={{ marginBottom: 16 }}>
            <dt>Trạng thái</dt><dd>{shipmentStatusLabel(shipment.status)}</dd>
            <dt>Điểm đi</dt><dd>{shipment.origin || '—'}</dd>
            <dt>Điểm đến</dt><dd>{shipment.destination || '—'}</dd>
            <dt>ETA</dt><dd>{shipment.estimated_delivery?.slice(0, 10) || '—'}</dd>
            <dt>Đồng bộ lần cuối</dt><dd>{shipment.last_synced_at?.slice(0, 16).replace('T', ' ') || '—'}</dd>
          </dl>
          <div style={{ marginBottom: 16 }}>
            <Button variant="primary" onClick={() => void sync()} disabled={syncing}>
              {syncing ? 'Đang sync...' : '↻ Đồng bộ tracking'}
            </Button>
          </div>
          <h4 className="section-title">Timeline sự kiện</h4>
          <div className="timeline">
            {events.map((ev) => (
              <div key={ev.id} className="timeline__item">
                <div className="timeline__time">{ev.event_time?.slice(0, 16).replace('T', ' ') || ev.created_at.slice(0, 16).replace('T', ' ')}</div>
                <div className="timeline__body">
                  <strong>{shipmentStatusLabel(ev.status) || 'Cập nhật'}</strong>
                  {ev.location && <span> · {ev.location}</span>}
                  {ev.description && <p>{ev.description}</p>}
                </div>
              </div>
            ))}
            {!events.length && <p className="muted">Chưa có sự kiện</p>}
          </div>
        </>
      )}
    </Modal>
  )
}
