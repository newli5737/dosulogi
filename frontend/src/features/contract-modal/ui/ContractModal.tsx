import { useEffect, useState, type ChangeEvent, type FormEvent } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { CustomerSelect } from '@/shared/ui/CustomerSelect/CustomerSelect'
import { contractApi } from '@/entities/contract/api/contractApi'
import type { Contract, ContractStatus, ServiceType } from '@/entities/contract/model/types'
import { useToken } from '@/app/providers/AuthProvider'

interface ContractFormState {
  customer_id: string
  title: string
  service_type: ServiceType
  start_date: string
  end_date: string
  value: string | number
  currency: string
  status: ContractStatus
  payment_terms: string
}

const empty: ContractFormState = {
  customer_id: '', title: '', service_type: 'FCL', start_date: '', end_date: '',
  value: '', currency: 'VND', status: 'draft', payment_terms: '',
}

const SERVICE_TYPES: ServiceType[] = ['FCL', 'LCL', 'air', 'express', 'road']
const STATUSES: ContractStatus[] = ['draft', 'active', 'expired', 'terminated']

interface ContractModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
  edit: Contract | null
}

export function ContractModal({ open, onClose, onSaved, edit }: ContractModalProps) {
  const token = useToken()
  const [form, setForm] = useState<ContractFormState>(empty)
  const [fileUrl, setFileUrl] = useState<string | null>(null)
  const [uploading, setUploading] = useState(false)
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
      setFileUrl(edit.file_url ?? null)
    } else {
      setForm({ ...empty, start_date: new Date().toISOString().slice(0, 10) })
      setFileUrl(null)
    }
  }, [open, edit])

  const set = <K extends keyof ContractFormState>(k: K, v: ContractFormState[K]) =>
    setForm((f) => ({ ...f, [k]: v }))

  async function handleFileUpload(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file || !token || !edit?.id) return
    setUploading(true)
    setError('')
    try {
      const updated = await contractApi.upload(token, edit.id, file)
      setFileUrl(updated.file_url ?? null)
      onSaved?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed')
    } finally {
      setUploading(false)
      e.target.value = ''
    }
  }

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!token) return
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
      setError(err instanceof Error ? err.message : 'Unknown error')
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
            <Select value={form.service_type} onChange={(e) => set('service_type', e.target.value as ServiceType)}>
              {SERVICE_TYPES.map((s) => <option key={s} value={s}>{s}</option>)}
            </Select>
          </Field>
          <Field label="Trạng thái">
            <Select value={form.status} onChange={(e) => set('status', e.target.value as ContractStatus)}>
              {STATUSES.map((s) => <option key={s} value={s}>{s}</option>)}
            </Select>
          </Field>
          <Field label="Ngày bắt đầu" required><Input type="date" value={form.start_date} onChange={(e) => set('start_date', e.target.value)} required /></Field>
          <Field label="Ngày kết thúc"><Input type="date" value={form.end_date} onChange={(e) => set('end_date', e.target.value)} /></Field>
          <Field label="Giá trị"><Input type="number" value={form.value} onChange={(e) => set('value', e.target.value)} /></Field>
          <Field label="Điều khoản TT"><Input value={form.payment_terms} onChange={(e) => set('payment_terms', e.target.value)} placeholder="30 ngày kể từ ngày xuất HĐ" /></Field>
        </div>
        {edit?.id && (
          <Field label="File hợp đồng">
            {fileUrl && (
              <p style={{ marginBottom: 8 }}>
                <a href={fileUrl} target="_blank" rel="noopener noreferrer">Tải file hiện tại</a>
              </p>
            )}
            <input type="file" className="field-input" onChange={handleFileUpload} disabled={uploading} />
            {uploading && <small>Đang upload...</small>}
          </Field>
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
