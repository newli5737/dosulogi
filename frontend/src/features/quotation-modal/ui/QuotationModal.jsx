import { useEffect, useState } from 'react'
import { Modal } from '../../../shared/ui/Modal/Modal'
import { Field, Input, Select } from '../../../shared/ui/Form/Form'
import { Button } from '../../../shared/ui/Button/Button'
import { CustomerSelect } from '../../../shared/ui/CustomerSelect/CustomerSelect'
import { quotationApi } from '../../../entities/quotation/api/quotationApi'
import { useToken } from '../../../app/providers/AuthProvider'

const emptyItem = { description: '', qty: 1, unit_price: 0 }
const empty = { customer_id: '', items: [{ ...emptyItem }], currency: 'VND', valid_until: '' }

function calcTotal(items) {
  return items.reduce((s, it) => s + (Number(it.qty) || 0) * (Number(it.unit_price) || 0), 0)
}

export function QuotationModal({ open, onClose, onSaved, edit }) {
  const token = useToken()
  const [form, setForm] = useState(empty)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    if (edit?.id) {
      setForm({
        customer_id: edit.customer_id || '',
        items: edit.items?.length ? edit.items : [{ ...emptyItem }],
        currency: edit.currency || 'VND',
        valid_until: edit.valid_until ? edit.valid_until.slice(0, 10) : '',
      })
    } else {
      setForm({ ...empty, items: [{ ...emptyItem }] })
    }
  }, [open, edit])

  const setItem = (i, k, v) => setForm((f) => {
    const items = [...f.items]
    items[i] = { ...items[i], [k]: v }
    return { ...f, items }
  })

  async function submit(e) {
    e.preventDefault()
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
        items,
        currency: form.currency,
        valid_until: form.valid_until || null,
        status: edit?.status || 'draft',
      }
      if (edit?.id) await quotationApi.update(token, edit.id, body)
      else await quotationApi.create(token, body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err.message)
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
          <Field label="Hết hạn"><Input type="date" value={form.valid_until} onChange={(e) => setForm((f) => ({ ...f, valid_until: e.target.value }))} /></Field>
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
          <p className="line-total">Tổng: {calcTotal(form.items).toLocaleString('vi-VN')} ₫</p>
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
