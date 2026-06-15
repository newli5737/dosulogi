import { useEffect, useState, type FormEvent } from 'react'
import { MessageCircle, MessageSquare, Plus, QrCode, Trash2 } from 'lucide-react'
import { chatApi, type ChatAccount } from '@/entities/chat/api/chatApi'
import { Button } from '@/shared/ui/Button/Button'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input } from '@/shared/ui/Form/Form'
import { chatAccountStatusLabel } from '@/shared/lib/labels'
import './chat-accounts-page.css'

export function ChatAccountsPage() {
  const [accounts, setAccounts] = useState<ChatAccount[]>([])
  const [open, setOpen] = useState(false)
  const [platform, setPlatform] = useState<'facebook' | 'zalo'>('facebook')
  const [name, setName] = useState('')
  const [cookies, setCookies] = useState('')
  const [loading, setLoading] = useState(false)
  const [qrModal, setQrModal] = useState<{ id: string; image?: string; status?: string } | null>(null)

  function reload() {
    chatApi.listAccounts().then(setAccounts).catch(console.error)
  }

  useEffect(() => { reload() }, [])

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    try {
      const acc = await chatApi.createAccount({ platform, name, cookies_json: platform === 'facebook' ? cookies : undefined })
      setOpen(false)
      setName('')
      setCookies('')
      reload()
      if (acc.platform === 'zalo') {
        const qr = await chatApi.zaloQR(acc.id)
        setQrModal({ id: acc.id, image: qr.qr_image, status: qr.status })
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Lỗi tạo tài khoản')
    } finally {
      setLoading(false)
    }
  }

  async function pollZaloStatus(id: string) {
    const st = await chatApi.zaloStatus(id)
    setQrModal((prev) => prev ? { ...prev, status: st.status, image: st.qr_image || prev.image } : null)
    if (st.connected) {
      reload()
      setTimeout(() => setQrModal(null), 1500)
    }
  }

  useEffect(() => {
    if (!qrModal?.id) return
    const t = window.setInterval(() => { void pollZaloStatus(qrModal.id) }, 2500)
    return () => clearInterval(t)
  }, [qrModal?.id])

  return (
    <div className="page-card">
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1>Quản trị kênh chat</h1>
          <p className="chat-acc-sub">Kết nối nhiều tài khoản Facebook & Zalo cho team sale</p>
        </div>
        <Button variant="primary" onClick={() => setOpen(true)}><Plus size={16} /> Thêm kênh</Button>
      </div>

      <div className="chat-acc-grid">
        {accounts.map((a) => (
          <article key={a.id} className="chat-acc-card">
            <div className="chat-acc-card__icon">
              {a.platform === 'facebook' ? <MessageCircle size={20} /> : <MessageSquare size={20} />}
            </div>
            <div>
              <strong>{a.name}</strong>
              <div className="chat-acc-card__meta">
                <span className={`status status--${a.status}`}>{chatAccountStatusLabel(a.status)}</span>
                <span>{a.platform}</span>
                {a.external_id && <span>ID: {a.external_id}</span>}
              </div>
            </div>
            <div className="chat-acc-card__actions">
              {a.platform === 'zalo' && (
                <Button variant="secondary" onClick={async () => {
                  const qr = await chatApi.zaloQR(a.id)
                  setQrModal({ id: a.id, image: qr.qr_image, status: qr.status })
                }}><QrCode size={14} /> QR</Button>
              )}
              <Button variant="secondary" onClick={async () => {
                if (!confirm('Xóa kênh này?')) return
                await chatApi.deleteAccount(a.id)
                reload()
              }}><Trash2 size={14} /></Button>
            </div>
          </article>
        ))}
        {accounts.length === 0 && <p className="chat-acc-empty">Chưa có kênh chat. Thêm Facebook (cookies) hoặc Zalo (quét QR).</p>}
      </div>

      <Modal open={open} onClose={() => setOpen(false)} title="Thêm kênh chat" icon={MessageSquare} tone="cyan">
        <form onSubmit={(e) => void handleCreate(e)} className="form-stack">
          <Field label="Nền tảng">
            <select className="field-input" value={platform} onChange={(e) => setPlatform(e.target.value as 'facebook' | 'zalo')}>
              <option value="facebook">Facebook Messenger</option>
              <option value="zalo">Zalo (QR)</option>
            </select>
          </Field>
          <Field label="Tên hiển thị"><Input value={name} onChange={(e) => setName(e.target.value)} required /></Field>
          {platform === 'facebook' && (
            <Field label="Cookies JSON">
              <textarea className="field-input" rows={6} value={cookies} onChange={(e) => setCookies(e.target.value)} placeholder="Dán cookies từ Facebook Web..." required />
            </Field>
          )}
          {platform === 'zalo' && <p className="hint">Sau khi tạo, quét mã QR bằng app Zalo trên điện thoại.</p>}
          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
            <Button type="button" variant="secondary" onClick={() => setOpen(false)}>Hủy</Button>
            <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang tạo...' : 'Tạo'}</Button>
          </div>
        </form>
      </Modal>

      <Modal open={!!qrModal} onClose={() => setQrModal(null)} title="Đăng nhập Zalo QR" icon={QrCode} tone="green">
        {qrModal?.image ? (
          <div className="qr-wrap">
            <img src={qrModal.image.startsWith('data:') ? qrModal.image : `data:image/png;base64,${qrModal.image}`} alt="Zalo QR" />
            <p>Trạng thái: {qrModal.status || 'waiting'}</p>
          </div>
        ) : (
          <p>Đang tạo mã QR...</p>
        )}
      </Modal>
    </div>
  )
}
