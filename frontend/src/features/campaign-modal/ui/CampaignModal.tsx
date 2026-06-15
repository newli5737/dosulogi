import { useEffect, useState, type FormEvent } from 'react'
import { Megaphone } from 'lucide-react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select, Textarea } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { campaignApi } from '@/entities/campaign/api/campaignApi'
import type { Campaign } from '@/entities/campaign/model/types'
import { CAMPAIGN_TYPE_OPTIONS, campaignTypeLabel } from '@/shared/lib/labels'
import './campaign-modal.css'

interface CampaignFormState {
  name: string
  type: string
  subject: string
  body_html: string
  segment_note: string
}

const empty: CampaignFormState = { name: '', type: 'email', subject: '', body_html: '', segment_note: '' }

const CONTENT_HINTS: Record<string, string> = {
  email: 'Nội dung HTML email — hỗ trợ thẻ p, strong, a, img',
  sms: 'Nội dung SMS tối đa 160 ký tự / tin. Dùng {ten_khach}, {ma_van_don}',
  zalo_oa: 'Nội dung tin Zalo OA — text hoặc template ID',
  zalo_zns: 'Template ZNS đã duyệt trên Zalo Business',
  facebook: 'Nội dung tin Messenger broadcast / tag',
  push: 'Tiêu đề + body push notification',
  in_app: 'Banner / popup trong app logistics',
  webhook: 'JSON payload gửi tới endpoint tích hợp',
  multi: 'Mô tả luồng đa kênh — hệ thống gửi theo thứ tự ưu tiên',
}

interface CampaignModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
  edit: Campaign | null
}

export function CampaignModal({ open, onClose, onSaved, edit }: CampaignModalProps) {
  const [form, setForm] = useState<CampaignFormState>(empty)
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
        segment_note: '',
      })
    } else {
      setForm(empty)
    }
  }, [open, edit])

  const set = <K extends keyof CampaignFormState>(k: K, v: CampaignFormState[K]) =>
    setForm((f) => ({ ...f, [k]: v }))

  const needsSubject = ['email', 'push', 'multi', 'facebook'].includes(form.type)
  const contentLabel = form.type === 'webhook' ? 'Payload JSON' : 'Nội dung chiến dịch'

  async function submit(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const { segment_note: _seg, ...rest } = form
      const body = { ...rest, subject: form.subject || null, body_html: form.body_html || null }
      if (edit?.id) await campaignApi.update(edit.id, body)
      else await campaignApi.create(body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={edit ? 'Sửa chiến dịch' : 'Tạo chiến dịch'} wide icon={Megaphone} tone="amber">
      <form onSubmit={submit}>
        <div className="campaign-modal__layout">
          <section>
            <h4 className="campaign-modal__section">Thông tin chiến dịch</h4>
            <div className="form-grid">
              <Field label="Tên chiến dịch" required>
                <Input value={form.name} onChange={(e) => set('name', e.target.value)} required />
              </Field>
              <Field label="Kênh marketing" required>
                <Select value={form.type} onChange={(e) => set('type', e.target.value)}>
                  {CAMPAIGN_TYPE_OPTIONS.map((t) => (
                    <option key={t.value} value={t.value}>{t.label}</option>
                  ))}
                </Select>
              </Field>
              {needsSubject && (
                <Field label="Tiêu đề / Subject">
                  <Input value={form.subject} onChange={(e) => set('subject', e.target.value)} />
                </Field>
              )}
              <div className="campaign-modal__full">
                <Field label="Phân khúc khách hàng">
                  <Input value={form.segment_note} onChange={(e) => set('segment_note', e.target.value)} placeholder="VD: tier gold, tỉnh HCM, khách cold chain..." />
                </Field>
              </div>
            </div>
          </section>

          <aside className="campaign-modal__channel-info">
            <strong>{campaignTypeLabel(form.type)}</strong>
            <p>{CONTENT_HINTS[form.type] || 'Nội dung chiến dịch'}</p>
          </aside>
        </div>

        <Field label={contentLabel}>
          <Textarea
            value={form.body_html}
            onChange={(e) => set('body_html', e.target.value)}
            rows={8}
            placeholder={CONTENT_HINTS[form.type]}
          />
        </Field>

        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <Button variant="secondary" type="button" onClick={onClose}>Hủy</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu chiến dịch'}</Button>
        </div>
      </form>
    </Modal>
  )
}
