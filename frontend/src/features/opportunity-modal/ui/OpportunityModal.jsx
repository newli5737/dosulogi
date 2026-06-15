import { useEffect, useState } from 'react'
import { Modal } from '../../../shared/ui/Modal/Modal'
import { Field, Input, Select, Textarea } from '../../../shared/ui/Form/Form'
import { Button } from '../../../shared/ui/Button/Button'
import { CustomerSelect } from '../../../shared/ui/CustomerSelect/CustomerSelect'
import { opportunityApi } from '../../../entities/opportunity/api/opportunityApi'
import { useToken } from '../../../app/providers/AuthProvider'

const empty = { customer_id: '', title: '', stage: 'lead', value: '', currency: 'VND', expected_close: '', lost_reason: '' }

export function OpportunityModal({ open, onClose, onSaved, edit }) {
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
        stage: edit.stage || 'lead',
        value: edit.value ?? '',
        currency: edit.currency || 'VND',
        expected_close: edit.expected_close ? edit.expected_close.slice(0, 10) : '',
        lost_reason: edit.lost_reason || '',
      })
    } else {
      setForm(empty)
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
        title: form.title,
        stage: form.stage,
        currency: form.currency,
        value: form.value ? Number(form.value) : null,
        expected_close: form.expected_close || null,
        lost_reason: form.lost_reason || null,
      }
      if (edit?.id) await opportunityApi.update(token, edit.id, body)
      else await opportunityApi.create(token, body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={edit ? 'Sửa cơ hội' : 'Thêm cơ hội'} wide>
      <form onSubmit={submit}>
        <div className="form-grid">
          <Field label="Khách hàng" required>
            <CustomerSelect value={form.customer_id} onChange={(v) => set('customer_id', v)} required />
          </Field>
          <Field label="Tiêu đề" required><Input value={form.title} onChange={(e) => set('title', e.target.value)} required /></Field>
          <Field label="Stage">
            <Select value={form.stage} onChange={(e) => set('stage', e.target.value)}>
              {['lead', 'qualified', 'proposal', 'negotiation', 'won', 'lost'].map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </Select>
          </Field>
          <Field label="Giá trị (VND)"><Input type="number" value={form.value} onChange={(e) => set('value', e.target.value)} /></Field>
          <Field label="Dự kiến đóng"><Input type="date" value={form.expected_close} onChange={(e) => set('expected_close', e.target.value)} /></Field>
          {form.stage === 'lost' && (
            <Field label="Lý do thua"><Textarea value={form.lost_reason} onChange={(e) => set('lost_reason', e.target.value)} /></Field>
          )}
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
