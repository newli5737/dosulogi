import { useEffect, useState } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select, Textarea } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { CustomerSelect } from '@/shared/ui/CustomerSelect/CustomerSelect'
import { opportunityApi } from '@/entities/opportunity/api/opportunityApi'
import { shipmentApi } from '@/entities/shipment/api/shipmentApi'
import type { Opportunity, StageHistoryEntry } from '@/entities/opportunity/model/types'
import type { Shipment } from '@/entities/shipment/model/types'

interface OpportunityFormState {
  customer_id: string
  title: string
  stage: string
  value: string | number
  currency: string
  expected_close: string
  lost_reason: string
  note: string
  shipment_ids: string[]
}

const empty: OpportunityFormState = {
  customer_id: '', title: '', stage: 'lead', value: '', currency: 'VND', expected_close: '', lost_reason: '', note: '', shipment_ids: [],
}

interface OpportunityModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
  edit: Opportunity | null
}

export function OpportunityModal({ open, onClose, onSaved, edit }: OpportunityModalProps) {
  const [form, setForm] = useState<OpportunityFormState>(empty)
  const [shipments, setShipments] = useState<Shipment[]>([])
  const [history, setHistory] = useState<StageHistoryEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    shipmentApi.list(1, 100).then((r) => setShipments(Array.isArray(r.data) ? r.data : [])).catch(console.error)
  }, [open])

  useEffect(() => {
    if (!open) return
    if (edit?.id) {
      opportunityApi.get(edit.id).then((res) => {
        const o = res.data
        setForm({
          customer_id: o.customer_id || '',
          title: o.title || '',
          stage: o.stage || 'lead',
          value: o.value ?? '',
          currency: o.currency || 'VND',
          expected_close: o.expected_close ? o.expected_close.slice(0, 10) : '',
          lost_reason: o.lost_reason || '',
          note: o.note || '',
          shipment_ids: o.shipment_ids || [],
        })
      }).catch(console.error)
      opportunityApi.stageHistory(edit.id).then((r) => setHistory(Array.isArray(r.data) ? r.data : [])).catch(console.error)
    } else {
      setForm(empty)
      setHistory([])
    }
  }, [open, edit])

  const set = <K extends keyof OpportunityFormState>(k: K, v: OpportunityFormState[K]) =>
    setForm((f) => ({ ...f, [k]: v }))

  function toggleShipment(id: string) {
    setForm((f) => ({
      ...f,
      shipment_ids: f.shipment_ids.includes(id)
        ? f.shipment_ids.filter((x) => x !== id)
        : [...f.shipment_ids, id],
    }))
  }

  async function submit(e: React.FormEvent) {
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
        note: form.note || null,
        shipment_ids: form.shipment_ids,
      }
      if (edit?.id) await opportunityApi.update(edit.id, body)
      else await opportunityApi.create(body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
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
          <Field label="Ghi chú"><Textarea value={form.note} onChange={(e) => set('note', e.target.value)} /></Field>
        </div>

        <div style={{ margin: '16px 0' }}>
          <strong>Liên kết vận đơn</strong>
          <div className="checkbox-list">
            {shipments.map((s) => (
              <label key={s.id}>
                <input type="checkbox" checked={form.shipment_ids.includes(s.id)} onChange={() => toggleShipment(s.id)} />
                {s.tracking_code} ({s.status || '—'})
              </label>
            ))}
            {!shipments.length && <p className="muted">Chưa có vận đơn</p>}
          </div>
        </div>

        {edit && history.length > 0 && (
          <div style={{ marginBottom: 16 }}>
            <strong>Lịch sử stage</strong>
            <ul className="stage-history">
              {history.map((h) => (
                <li key={h.id}>
                  {h.from_stage || '—'} → <strong>{h.to_stage}</strong>
                  {' · '}{h.changed_at.slice(0, 16).replace('T', ' ')}
                  {h.changer_name && ` · ${h.changer_name}`}
                </li>
              ))}
            </ul>
          </div>
        )}

        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu'}</Button>
        </div>
      </form>
    </Modal>
  )
}
