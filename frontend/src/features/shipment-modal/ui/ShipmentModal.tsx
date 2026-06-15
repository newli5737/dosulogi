import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { CustomerSelect } from '@/shared/ui/CustomerSelect/CustomerSelect'
import { shipmentApi } from '@/entities/shipment/api/shipmentApi'

interface ShipmentFormState {
  tracking_code: string
  customer_id: string
  origin: string
  destination: string
}

const empty: ShipmentFormState = { tracking_code: '', customer_id: '', origin: '', destination: '' }

interface ShipmentModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
}

export function ShipmentModal({ open, onClose, onSaved }: ShipmentModalProps) {
  const [form, setForm] = useState<ShipmentFormState>(empty)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (open) setForm(empty)
  }, [open])

  const set = <K extends keyof ShipmentFormState>(k: K, v: ShipmentFormState[K]) =>
    setForm((f) => ({ ...f, [k]: v }))

  async function submit(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await shipmentApi.create({
        tracking_code: form.tracking_code,
        customer_id: form.customer_id || null,
        origin: form.origin || null,
        destination: form.destination || null,
      })
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Thêm vận đơn">
      <form onSubmit={submit}>
        <Field label="Mã vận đơn" required><Input value={form.tracking_code} onChange={(e) => set('tracking_code', e.target.value)} required /></Field>
        <Field label="Khách hàng"><CustomerSelect value={form.customer_id} onChange={(v) => set('customer_id', v)} /></Field>
        <Field label="Điểm đi"><Input value={form.origin} onChange={(e) => set('origin', e.target.value)} /></Field>
        <Field label="Điểm đến"><Input value={form.destination} onChange={(e) => set('destination', e.target.value)} /></Field>
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang tạo...' : 'Tạo'}</Button>
        </div>
      </form>
    </Modal>
  )
}
