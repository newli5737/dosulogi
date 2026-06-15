import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select, Textarea } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { contactApi } from '@/entities/contact/api/contactApi'
import type { Contact, ContactInput } from '@/entities/contact/model/types'
import { useToken } from '@/app/providers/AuthProvider'

const empty: ContactInput = { name: '', role: '', phone: '', email: '', is_primary: false, note: '' }

interface ContactModalProps {
  open: boolean
  customerId: string
  edit: Contact | null
  onClose: () => void
  onSaved?: () => void
}

export function ContactModal({ open, customerId, edit, onClose, onSaved }: ContactModalProps) {
  const token = useToken()
  const [form, setForm] = useState<ContactInput>(empty)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    if (edit) {
      setForm({
        name: edit.name,
        role: edit.role || '',
        phone: edit.phone || '',
        email: edit.email || '',
        is_primary: edit.is_primary,
        note: edit.note || '',
      })
    } else {
      setForm(empty)
    }
  }, [open, edit])

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    setLoading(true)
    setError('')
    try {
      if (edit?.id) await contactApi.update(token, customerId, edit.id, form)
      else await contactApi.create(token, customerId, form)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={edit ? 'Sửa liên hệ' : 'Thêm liên hệ'}>
      <form onSubmit={submit}>
        <Field label="Họ tên" required><Input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} required /></Field>
        <Field label="Chức vụ"><Input value={form.role || ''} onChange={(e) => setForm((f) => ({ ...f, role: e.target.value }))} /></Field>
        <Field label="Email"><Input type="email" value={form.email || ''} onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))} /></Field>
        <Field label="Điện thoại"><Input value={form.phone || ''} onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))} /></Field>
        <Field label="Liên hệ chính">
          <Select value={form.is_primary ? '1' : '0'} onChange={(e) => setForm((f) => ({ ...f, is_primary: e.target.value === '1' }))}>
            <option value="0">Không</option>
            <option value="1">Có</option>
          </Select>
        </Field>
        <Field label="Ghi chú"><Textarea value={form.note || ''} onChange={(e) => setForm((f) => ({ ...f, note: e.target.value }))} /></Field>
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu'}</Button>
        </div>
      </form>
    </Modal>
  )
}
