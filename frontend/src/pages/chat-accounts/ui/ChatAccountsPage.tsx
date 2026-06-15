import { useEffect, useRef, useState, type ChangeEvent, type FormEvent } from 'react'
import { MessageCircle, MessageSquare, Pencil, Plus, QrCode, Trash2, Upload } from 'lucide-react'
import { chatApi, type ChatAccount } from '@/entities/chat/api/chatApi'
import { Button } from '@/shared/ui/Button/Button'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input } from '@/shared/ui/Form/Form'
import { chatAccountStatusLabel } from '@/shared/lib/labels'
import './chat-accounts-page.css'

async function parseCookiesFile(file: File): Promise<string> {
  const text = (await file.text()).trim()
  let parsed: unknown
  try {
    parsed = JSON.parse(text)
  } catch {
    throw new Error('File không phải JSON hợp lệ')
  }
  if (Array.isArray(parsed)) {
    if (parsed.length === 0) throw new Error('Mảng cookies trống')
    return JSON.stringify(parsed)
  }
  if (parsed && typeof parsed === 'object' && Array.isArray((parsed as { cookies?: unknown }).cookies)) {
    const cookies = (parsed as { cookies: unknown[] }).cookies
    if (cookies.length === 0) throw new Error('Mảng cookies trống')
    return JSON.stringify(cookies)
  }
  throw new Error('JSON phải là mảng cookies hoặc object { cookies: [...] }')
}

export function ChatAccountsPage() {
  const [accounts, setAccounts] = useState<ChatAccount[]>([])
  const [open, setOpen] = useState(false)
  const [editAccount, setEditAccount] = useState<ChatAccount | null>(null)
  const [platform, setPlatform] = useState<'facebook' | 'zalo'>('facebook')
  const [name, setName] = useState('')
  const [cookiesFile, setCookiesFile] = useState<File | null>(null)
  const [cookiesFileName, setCookiesFileName] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [qrModal, setQrModal] = useState<{ id: string; image?: string; status?: string } | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)
  const editFileRef = useRef<HTMLInputElement>(null)

  function reload() {
    chatApi.listAccounts().then(setAccounts).catch(console.error)
  }

  useEffect(() => { reload() }, [])

  function resetCreateForm() {
    setName('')
    setCookiesFile(null)
    setCookiesFileName('')
    setError('')
    if (fileRef.current) fileRef.current.value = ''
  }

  function onCookiesPick(e: ChangeEvent<HTMLInputElement>, forEdit = false) {
    const file = e.target.files?.[0]
    if (!file) return
    if (!file.name.endsWith('.json') && file.type !== 'application/json') {
      setError('Chọn file .json export từ extension cookies')
      return
    }
    if (forEdit) {
      setCookiesFile(file)
      setCookiesFileName(file.name)
    } else {
      setCookiesFile(file)
      setCookiesFileName(file.name)
    }
    setError('')
  }

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      let cookiesJson: string | undefined
      if (platform === 'facebook') {
        if (!cookiesFile) {
          setError('Vui lòng upload file cookies.json')
          setLoading(false)
          return
        }
        cookiesJson = await parseCookiesFile(cookiesFile)
      }
      const acc = await chatApi.createAccount({ platform, name, cookies_json: cookiesJson })
      setOpen(false)
      resetCreateForm()
      reload()
      if (acc.platform === 'zalo') {
        const qr = await chatApi.zaloQR(acc.id)
        setQrModal({ id: acc.id, image: qr.qr_image, status: qr.status })
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Lỗi tạo tài khoản')
    } finally {
      setLoading(false)
    }
  }

  async function handleUpdate(e: FormEvent) {
    e.preventDefault()
    if (!editAccount) return
    setLoading(true)
    setError('')
    try {
      const body: { name?: string; cookies_json?: string } = { name }
      if (editAccount.platform === 'facebook' && cookiesFile) {
        body.cookies_json = await parseCookiesFile(cookiesFile)
      }
      await chatApi.updateAccount(editAccount.id, body)
      setEditAccount(null)
      resetCreateForm()
      reload()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Lỗi cập nhật')
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

  function openEdit(acc: ChatAccount) {
    setEditAccount(acc)
    setName(acc.name)
    setCookiesFile(null)
    setCookiesFileName('')
    setError('')
    if (editFileRef.current) editFileRef.current.value = ''
  }

  return (
    <div className="page-card">
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1>Quản trị kênh chat</h1>
          <p className="chat-acc-sub">Facebook: upload cookies.json · Zalo: quét QR trên điện thoại</p>
        </div>
        <Button variant="primary" onClick={() => { resetCreateForm(); setOpen(true) }}><Plus size={16} /> Thêm kênh</Button>
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
                <span>{a.platform === 'facebook' ? 'Facebook' : 'Zalo'}</span>
                {a.has_credentials
                  ? <span className="chat-acc-cred">Đã cấu hình session</span>
                  : <span className="chat-acc-cred chat-acc-cred--missing">Chưa có session</span>}
                {a.external_id && <span>ID: {a.external_id}</span>}
              </div>
            </div>
            <div className="chat-acc-card__actions">
              <Button variant="secondary" onClick={() => openEdit(a)}><Pencil size={14} /> Quản lý</Button>
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
        {accounts.length === 0 && (
          <p className="chat-acc-empty">
            Chưa có kênh chat. Thêm Facebook (upload cookies.json) hoặc Zalo (quét QR).
          </p>
        )}
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
            <Field label="File cookies.json">
              <div className="cookie-upload">
                <input ref={fileRef} type="file" accept=".json,application/json" className="cookie-upload__input" onChange={(e) => onCookiesPick(e)} />
                <div className="cookie-upload__box">
                  <Upload size={20} />
                  <span>{cookiesFileName || 'Chọn file cookies export từ trình duyệt'}</span>
                  <small>Dùng extension EditThisCookie / Cookie-Editor → Export JSON</small>
                </div>
              </div>
            </Field>
          )}
          {platform === 'zalo' && <p className="hint">Sau khi tạo, quét mã QR bằng app Zalo trên điện thoại.</p>}
          {error && <p className="form-error">{error}</p>}
          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
            <Button type="button" variant="secondary" onClick={() => setOpen(false)}>Hủy</Button>
            <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang tạo...' : 'Tạo kênh'}</Button>
          </div>
        </form>
      </Modal>

      <Modal open={!!editAccount} onClose={() => setEditAccount(null)} title="Quản lý kênh chat" icon={Pencil} tone="blue">
        {editAccount && (
          <form onSubmit={(e) => void handleUpdate(e)} className="form-stack">
            <Field label="Nền tảng">
              <Input value={editAccount.platform === 'facebook' ? 'Facebook Messenger' : 'Zalo'} readOnly />
            </Field>
            <Field label="Tên hiển thị"><Input value={name} onChange={(e) => setName(e.target.value)} required /></Field>
            {editAccount.platform === 'facebook' && (
              <Field label="Cập nhật cookies.json">
                <div className="cookie-upload">
                  <input ref={editFileRef} type="file" accept=".json,application/json" className="cookie-upload__input" onChange={(e) => onCookiesPick(e, true)} />
                  <div className="cookie-upload__box">
                    <Upload size={20} />
                    <span>{cookiesFileName || 'Chọn file mới để thay session (tuỳ chọn)'}</span>
                    <small>{editAccount.has_credentials ? 'Session hiện tại đang hoạt động' : 'Chưa có session — cần upload'}</small>
                  </div>
                </div>
              </Field>
            )}
            {editAccount.platform === 'zalo' && (
              <p className="hint">Zalo dùng QR login. Bấm nút QR trên thẻ kênh để đăng nhập lại.</p>
            )}
            {error && <p className="form-error">{error}</p>}
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <Button type="button" variant="secondary" onClick={() => setEditAccount(null)}>Đóng</Button>
              <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu thay đổi'}</Button>
            </div>
          </form>
        )}
      </Modal>

      <Modal open={!!qrModal} onClose={() => setQrModal(null)} title="Đăng nhập Zalo QR" icon={QrCode} tone="green">
        {qrModal?.image ? (
          <div className="qr-wrap">
            <img src={qrModal.image.startsWith('data:') ? qrModal.image : `data:image/png;base64,${qrModal.image}`} alt="Zalo QR" />
            <p>Trạng thái: {chatAccountStatusLabel(qrModal.status) || 'Chờ quét'}</p>
          </div>
        ) : (
          <p>Đang tạo mã QR...</p>
        )}
      </Modal>
    </div>
  )
}
