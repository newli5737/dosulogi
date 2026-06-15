import { useEffect, useState } from 'react'
import { Modal } from '../../../shared/ui/Modal/Modal'
import { Field, Input, Select } from '../../../shared/ui/Form/Form'
import { Button } from '../../../shared/ui/Button/Button'
import { CustomerSelect } from '../../../shared/ui/CustomerSelect/CustomerSelect'
import { contractApi } from '../../../entities/contract/api/contractApi'
import { useToken } from '../../../app/providers/AuthProvider'

const empty = {
  customer_id: '', title: '', service_type: 'FCL', start_date: '', end_date: '',
  value: '', currency: 'VND', status: 'draft', payment_terms: '',
}

export function ContractModal({ open, onClose, onSaved, edit }) {
  const token = useToken()
  const [form, setForm] = useState(empty)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    if (edit?.id) {
      setForm({
        customer_id: edit.customer_id || '',
        title: edit.title || '',
        service_type: edit.service_type || 'FCL',
        start_date: edit.start_date ? edit.start_date.slice(0, 10) : '',
        end_date: edit.end_date ? edit.end_date.slice(0, 10) : '',
        value: edit.value ?? '',
        currency: edit.currency || 'VND',
        status: edit.status || 'draft',
        payment_terms: edit.payment_terms || '',
      })
    } else {
      setForm({ ...empty, start_date: new Date().toISOString().slice(0, 10) })
    }
  }, [open, edit])

  const set = (k, v) => setForm((f) => ({ ...f, [k]: v }))

  async function submit(e) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const body = {
        customer_id: form.customer_id,
        title: form.title || null,
        service_type: form.service_type,
        start_date: form.start_date,
        end_date: form.end_date || null,
        value: form.value ? Number(form.value) : null,
        currency: form.currency,
        status: form.status,
        payment_terms: form.payment_terms || null,
      }
      if (edit?.id) await contractApi.update(token, edit.id, body)
      else await contractApi.create(token, body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={edit ? 'Sửa hợp đồng' : 'Thêm hợp đồng'} wide>
      <form onSubmit={submit}>
        <div className="form-grid">
          <Field label="Khách hàng" required>
            <CustomerSelect value={form.customer_id} onChange={(v) => set('customer_id', v)} required />
          </Field>
          <Field label="Tiêu đề"><Input value={form.title} onChange={(e) => set('title', e.target.value)} /></Field>
          <Field label="Dịch vụ">
            <Select value={form.service_type} onChange={(e) => set('service_type', e.target.value)}>
              {['FCL', 'LCL', 'air', 'express', 'road'].map((s) => <option key={s} value={s}>{s}</option>)}
            </Select>
          </Field>
          <Field label="Trạng thái">
            <Select value={form.status} onChange={(e) => set('status', e.target.value)}>
              {['draft', 'active', 'expired', 'terminated'].map((s) => <option key={s} value={s}>{s}</option>)}
            </Select>
          </Field>
          <Field label="Ngày bắt đầu" required><Input type="date" value={form.start_date} onChange={(e) => set('start_date', e.target.value)} required /></Field>
          <Field label="Ngày kết thúc"><Input type="date" value={form.end_date} onChange={(e) => set('end_date', e.target.value)} /></Field>
          <Field label="Giá trị"><Input type="number" value={form.value} onChange={(e) => set('value', e.target.value)} /></Field>
          <Field label="Điều khoản TT"><Input value={form.payment_terms} onChange={(e) => set('payment_terms', e.target.value)} placeholder="30 ngày kể từ ngày xuất HĐ" /></Field>
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
