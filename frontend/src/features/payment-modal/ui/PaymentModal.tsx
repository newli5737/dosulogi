import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select, Textarea } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { paymentApi } from '@/entities/payment/api/paymentApi'
import { invoiceApi } from '@/entities/invoice/api/invoiceApi'
import type { Invoice } from '@/entities/invoice/model/types'
import { useToken } from '@/app/providers/AuthProvider'

interface PaymentFormState {
  invoice_id: string
  amount: string | number
  method: string
  reference_code: string
  note: string
}

const empty: PaymentFormState = { invoice_id: '', amount: '', method: 'bank_transfer', reference_code: '', note: '' }

interface PaymentModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
}

export function PaymentModal({ open, onClose, onSaved }: PaymentModalProps) {
  const token = useToken()
  const [form, setForm] = useState<PaymentFormState>(empty)
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open || !token) return
    setForm(empty)
    invoiceApi.list(token, 1, 100).then((res) => setInvoices(res.data || [])).catch(() => setInvoices([]))
  }, [open, token])

  const set = <K extends keyof PaymentFormState>(k: K, v: PaymentFormState[K]) =>
    setForm((f) => ({ ...f, [k]: v }))

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    setLoading(true)
    setError('')
    try {
      await paymentApi.create(token, {
        invoice_id: form.invoice_id,
        amount: Number(form.amount) || 0,
        method: form.method || null,
        reference_code: form.reference_code || null,
        note: form.note || null,
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
    <Modal open={open} onClose={onClose} title="Ghi nhận thanh toán">
      <form onSubmit={submit}>
        <Field label="Hóa đơn" required>
          <Select value={form.invoice_id} onChange={(e) => set('invoice_id', e.target.value)} required>
            <option value="">— Chọn hóa đơn —</option>
            {invoices.map((inv) => (
              <option key={inv.id} value={inv.id}>
                {inv.code} — {inv.total ? `${Number(inv.total).toLocaleString('vi-VN')} ₫` : '—'}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Số tiền" required>
          <Input type="number" value={form.amount} onChange={(e) => set('amount', e.target.value)} required />
        </Field>
        <Field label="Phương thức">
          <Select value={form.method} onChange={(e) => set('method', e.target.value)}>
            <option value="bank_transfer">Chuyển khoản</option>
            <option value="cash">Tiền mặt</option>
            <option value="card">Thẻ</option>
          </Select>
        </Field>
        <Field label="Mã tham chiếu"><Input value={form.reference_code} onChange={(e) => set('reference_code', e.target.value)} /></Field>
        <Field label="Ghi chú"><Textarea value={form.note} onChange={(e) => set('note', e.target.value)} /></Field>
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu'}</Button>
        </div>
      </form>
    </Modal>
  )
}
