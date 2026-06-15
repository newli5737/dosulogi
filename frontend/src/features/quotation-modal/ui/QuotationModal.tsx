import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { CustomerSelect } from '@/shared/ui/CustomerSelect/CustomerSelect'
import { quotationApi } from '@/entities/quotation/api/quotationApi'
import type { Quotation } from '@/entities/quotation/model/types'
import { useToken } from '@/app/providers/AuthProvider'

interface LineItemForm {
  description: string
  qty: number | string
  unit_price: number | string
}

interface QuotationFormState {
  customer_id: string
  opportunity_id: string
  items: LineItemForm[]
  currency: string
  valid_until: string
  discount: string | number
  tax_rate: string | number
  note: string
}

const emptyItem: LineItemForm = { description: '', qty: 1, unit_price: 0 }

function calcSubtotal(items: LineItemForm[]): number {
  return items.reduce((s, it) => s + (Number(it.qty) || 0) * (Number(it.unit_price) || 0), 0)
}

interface QuotationModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
  edit: Quotation | null
}

export function QuotationModal({ open, onClose, onSaved, edit }: QuotationModalProps) {
  const token = useToken()
  const [form, setForm] = useState<QuotationFormState>({
    customer_id: '', opportunity_id: '', items: [{ ...emptyItem }], currency: 'VND', valid_until: '', discount: 0, tax_rate: 10, note: '',
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    if (edit?.id) {
      setForm({
        customer_id: edit.customer_id || '',
        opportunity_id: edit.opportunity_id || edit.opp_id || '',
        items: edit.items?.length
          ? edit.items.map((it) => ({ description: it.description, qty: it.qty, unit_price: it.unit_price }))
          : [{ ...emptyItem }],
        currency: edit.currency || 'VND',
        valid_until: edit.valid_until ? edit.valid_until.slice(0, 10) : '',
        discount: edit.discount ?? 0,
        tax_rate: edit.tax_rate ?? 10,
        note: edit.note || '',
      })
    } else {
      setForm({ customer_id: '', opportunity_id: '', items: [{ ...emptyItem }], currency: 'VND', valid_until: '', discount: 0, tax_rate: 10, note: '' })
    }
  }, [open, edit])

  const setItem = (i: number, k: keyof LineItemForm, v: string | number) => setForm((f) => {
    const items = [...f.items]
    items[i] = { ...items[i], [k]: v } as LineItemForm
    return { ...f, items }
  })

  const subtotal = calcSubtotal(form.items)
  const discount = Number(form.discount) || 0
  const taxRate = Number(form.tax_rate) || 0
  const taxable = Math.max(subtotal - discount, 0)
  const taxAmount = taxable * taxRate / 100
  const total = taxable + taxAmount

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    setLoading(true)
    setError('')
    try {
      const items = form.items.map((it) => ({
        description: it.description,
        qty: Number(it.qty) || 1,
        unit_price: Number(it.unit_price) || 0,
        amount: (Number(it.qty) || 1) * (Number(it.unit_price) || 0),
      }))
      const body = {
        customer_id: form.customer_id,
        opportunity_id: form.opportunity_id || null,
        items,
        currency: form.currency,
        valid_until: form.valid_until || null,
        status: edit?.status || 'draft',
        discount: discount || null,
        tax_rate: taxRate || null,
        note: form.note || null,
      }
      if (edit?.id) await quotationApi.update(token, edit.id, body)
      else await quotationApi.create(token, body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={edit ? 'Sửa báo giá' : 'Tạo báo giá'} wide>
      <form onSubmit={submit}>
        <div className="form-grid">
          <Field label="Khách hàng" required>
            <CustomerSelect value={form.customer_id} onChange={(v) => setForm((f) => ({ ...f, customer_id: v }))} required />
          </Field>
          <Field label="Opportunity ID"><Input value={form.opportunity_id} onChange={(e) => setForm((f) => ({ ...f, opportunity_id: e.target.value }))} placeholder="UUID cơ hội (tuỳ chọn)" /></Field>
          <Field label="Hết hạn"><Input type="date" value={form.valid_until} onChange={(e) => setForm((f) => ({ ...f, valid_until: e.target.value }))} /></Field>
          <Field label="Giảm giá"><Input type="number" value={form.discount} onChange={(e) => setForm((f) => ({ ...f, discount: e.target.value }))} /></Field>
          <Field label="Thuế (%)"><Input type="number" value={form.tax_rate} onChange={(e) => setForm((f) => ({ ...f, tax_rate: e.target.value }))} /></Field>
        </div>
        <div className="line-items">
          <strong>Dòng hàng</strong>
          {form.items.map((it, i) => (
            <div key={i} className="line-item-row">
              <Input placeholder="Mô tả" value={it.description} onChange={(e) => setItem(i, 'description', e.target.value)} />
              <Input type="number" placeholder="SL" value={it.qty} onChange={(e) => setItem(i, 'qty', e.target.value)} style={{ width: 70 }} />
              <Input type="number" placeholder="Đơn giá" value={it.unit_price} onChange={(e) => setItem(i, 'unit_price', e.target.value)} style={{ width: 120 }} />
              {form.items.length > 1 && (
                <Button variant="secondary" onClick={() => setForm((f) => ({ ...f, items: f.items.filter((_, j) => j !== i) }))}>×</Button>
              )}
            </div>
          ))}
          <Button variant="secondary" onClick={() => setForm((f) => ({ ...f, items: [...f.items, { ...emptyItem }] }))}>+ Dòng</Button>
          <p className="line-total">
            Tạm tính: {subtotal.toLocaleString('vi-VN')} ₫ · Thuế: {taxAmount.toLocaleString('vi-VN')} ₫ ·
            <strong> Tổng: {total.toLocaleString('vi-VN')} ₫</strong>
          </p>
        </div>
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu'}</Button>
        </div>
      </form>
    </Modal>
  )
}
