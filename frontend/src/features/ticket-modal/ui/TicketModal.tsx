import { useEffect, useState, type FormEvent } from 'react'
import { Ticket } from 'lucide-react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select, Textarea } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { ticketApi } from '@/entities/ticket/api/ticketApi'
import { customerApi } from '@/entities/customer/api/customerApi'
import type { Customer } from '@/entities/customer/model/types'

interface TicketFormState {
  customer_id: string
  title: string
  description: string
  priority: string
  category: string
}

const empty: TicketFormState = { customer_id: '', title: '', description: '', priority: 'medium', category: 'shipment' }

interface TicketModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
}

export function TicketModal({ open, onClose, onSaved }: TicketModalProps) {
  const [form, setForm] = useState<TicketFormState>(empty)
  const [customers, setCustomers] = useState<Customer[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    setForm(empty)
    customerApi.list(1, 100).then((res) => setCustomers(res.data || [])).catch(() => setCustomers([]))
  }, [open])

  async function submit(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await ticketApi.create({ ...form, description: form.description || null })
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Tạo ticket" icon={Ticket} tone="rose">
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
