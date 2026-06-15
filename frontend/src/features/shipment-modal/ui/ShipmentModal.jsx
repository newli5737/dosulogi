import { useEffect, useState } from 'react'
import { Modal } from '../../../shared/ui/Modal/Modal'
import { Field, Input } from '../../../shared/ui/Form/Form'
import { Button } from '../../../shared/ui/Button/Button'
import { CustomerSelect } from '../../../shared/ui/CustomerSelect/CustomerSelect'
import { shipmentApi } from '../../../entities/shipment/api/shipmentApi'
import { useToken } from '../../../app/providers/AuthProvider'

const empty = { tracking_code: '', customer_id: '', origin: '', destination: '' }

export function ShipmentModal({ open, onClose, onSaved }) {
  const token = useToken()
  const [form, setForm] = useState(empty)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (open) setForm(empty)
  }, [open])

  const set = (k, v) => setForm((f) => ({ ...f, [k]: v }))

  async function submit(e) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await shipmentApi.create(token, {
        tracking_code: form.tracking_code,
        customer_id: form.customer_id || null,
        origin: form.origin || null,
        destination: form.destination || null,
      })
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err.message)
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
