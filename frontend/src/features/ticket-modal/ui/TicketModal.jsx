import { useEffect, useState } from 'react'
import { Modal } from '../../../shared/ui/Modal/Modal'
import { Field, Input, Select, Textarea } from '../../../shared/ui/Form/Form'
import { Button } from '../../../shared/ui/Button/Button'
import { ticketApi } from '../../../entities/ticket/api/ticketApi'
import { customerApi } from '../../../entities/customer/api/customerApi'
import { useToken } from '../../../app/providers/AuthProvider'

const empty = { customer_id: '', title: '', description: '', priority: 'medium', category: 'shipment' }

export function TicketModal({ open, onClose, onSaved }) {
  const token = useToken()
  const [form, setForm] = useState(empty)
  const [customers, setCustomers] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open || !token) return
    setForm(empty)
    customerApi.list(token, 1, 100).then((res) => setCustomers(res.data || [])).catch(() => setCustomers([]))
  }, [open, token])

  async function submit(e) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await ticketApi.create(token, { ...form, description: form.description || null })
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Tạo ticket">
      <form onSubmit={submit}>
        <Field label="Khách hàng" required>
          <Select value={form.customer_id} onChange={(e) => setForm({ ...form, customer_id: e.target.value })} required>
            <option value="">— Chọn KH —</option>
            {customers.map((c) => <option key={c.id} value={c.id}>{c.code} — {c.name}</option>)}
          </Select>
        </Field>
        <Field label="Tiêu đề" required><Input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required /></Field>
        <Field label="Mô tả"><Textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} /></Field>
        <Field label="Danh mục">
          <Select value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })}>
            <option value="shipment">Vận đơn</option>
            <option value="billing">Thanh toán</option>
            <option value="complaint">Khiếu nại</option>
            <option value="other">Khác</option>
          </Select>
        </Field>
        <Field label="Ưu tiên">
          <Select value={form.priority} onChange={(e) => setForm({ ...form, priority: e.target.value })}>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="urgent">Urgent</option>
          </Select>
        </Field>
        {error && <p style={{ color: 'var(--danger)', fontSize: '.875rem' }}>{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang tạo...' : 'Tạo ticket'}</Button>
        </div>
      </form>
    </Modal>
  )
}
