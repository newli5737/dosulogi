import { useEffect, useState } from 'react'
import { Modal } from '../../../shared/ui/Modal/Modal'
import { Field, Input, Select, Textarea } from '../../../shared/ui/Form/Form'
import { Button } from '../../../shared/ui/Button/Button'
import { campaignApi } from '../../../entities/campaign/api/campaignApi'
import { useToken } from '../../../app/providers/AuthProvider'

const empty = { name: '', type: 'email', subject: '', body_html: '' }

export function CampaignModal({ open, onClose, onSaved, edit }) {
  const token = useToken()
  const [form, setForm] = useState(empty)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    if (edit?.id) {
      setForm({
        name: edit.name || '',
        type: edit.type || 'email',
        subject: edit.subject || '',
        body_html: edit.body_html || '',
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
      const body = { ...form, subject: form.subject || null, body_html: form.body_html || null }
      if (edit?.id) await campaignApi.update(token, edit.id, body)
      else await campaignApi.create(token, body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={edit ? 'Sửa chiến dịch' : 'Tạo chiến dịch'} wide>
      <form onSubmit={submit}>
        <Field label="Tên" required><Input value={form.name} onChange={(e) => set('name', e.target.value)} required /></Field>
        <Field label="Loại">
          <Select value={form.type} onChange={(e) => set('type', e.target.value)}>
            <option value="email">Email</option>
            <option value="sms">SMS</option>
          </Select>
        </Field>
        <Field label="Tiêu đề"><Input value={form.subject} onChange={(e) => set('subject', e.target.value)} /></Field>
        <Field label="Nội dung HTML"><Textarea value={form.body_html} onChange={(e) => set('body_html', e.target.value)} rows={5} /></Field>
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu'}</Button>
        </div>
      </form>
    </Modal>
  )
}
