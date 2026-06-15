import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select, Textarea } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { interactionApi } from '@/entities/interaction/api/interactionApi'
import type { InteractionInput } from '@/entities/interaction/model/types'
import { useToken } from '@/app/providers/AuthProvider'

const empty: InteractionInput = { channel: 'call', direction: 'outbound', summary: '', occurred_at: '' }

interface InteractionModalProps {
  open: boolean
  customerId: string
  onClose: () => void
  onSaved?: () => void
}

export function InteractionModal({ open, customerId, onClose, onSaved }: InteractionModalProps) {
  const token = useToken()
  const [form, setForm] = useState<InteractionInput>(empty)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (open) setForm(empty)
  }, [open])

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    setLoading(true)
    setError('')
    try {
      await interactionApi.create(token, customerId, {
        ...form,
        occurred_at: form.occurred_at || null,
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
    <Modal open={open} onClose={onClose} title="Ghi nhận tương tác">
      <form onSubmit={submit}>
        <Field label="Kênh" required>
          <Select value={form.channel} onChange={(e) => setForm((f) => ({ ...f, channel: e.target.value }))}>
            {['call', 'email', 'meeting', 'chat', 'visit', 'other'].map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </Select>
        </Field>
        <Field label="Hướng">
          <Select value={form.direction || 'outbound'} onChange={(e) => setForm((f) => ({ ...f, direction: e.target.value }))}>
            <option value="inbound">Inbound</option>
            <option value="outbound">Outbound</option>
          </Select>
        </Field>
        <Field label="Thời gian"><Input type="datetime-local" value={form.occurred_at || ''} onChange={(e) => setForm((f) => ({ ...f, occurred_at: e.target.value }))} /></Field>
        <Field label="Tóm tắt" required><Textarea value={form.summary} onChange={(e) => setForm((f) => ({ ...f, summary: e.target.value }))} required /></Field>
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu'}</Button>
        </div>
      </form>
    </Modal>
  )
}
