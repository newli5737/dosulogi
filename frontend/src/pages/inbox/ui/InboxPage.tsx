import { useCallback, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { Loader2, MessageCircle, RefreshCw, Send, UserPlus } from 'lucide-react'
import { chatApi, type ChatAccount, type ChatMessage, type ChatThread, type ConversationMeta } from '@/entities/chat/api/chatApi'
import { customerApi } from '@/entities/customer/api/customerApi'
import type { Customer } from '@/entities/customer/model/types'
import { Button } from '@/shared/ui/Button/Button'
import { useAuth } from '@/app/providers/AuthProvider'
import './inbox-page.css'

type Platform = 'facebook' | 'zalo' | ''

export function InboxPage() {
  const { session } = useAuth()
  const [platform, setPlatform] = useState<Platform>('')
  const [accounts, setAccounts] = useState<ChatAccount[]>([])
  const [selectedAccount, setSelectedAccount] = useState('')
  const [threads, setThreads] = useState<ChatThread[]>([])
  const [selectedThread, setSelectedThread] = useState<ChatThread | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [conversation, setConversation] = useState<ConversationMeta | null>(null)
  const [assignees, setAssignees] = useState<{ id: string; full_name: string }[]>([])
  const [customers, setCustomers] = useState<Customer[]>([])
  const [reply, setReply] = useState('')
  const [loadingInbox, setLoadingInbox] = useState(false)
  const [loadingThread, setLoadingThread] = useState(false)
  const [sending, setSending] = useState(false)
  const [error, setError] = useState('')
  const pollRef = useRef<number | null>(null)
  const bottomRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    chatApi.listAccounts().then(setAccounts).catch(console.error)
    chatApi.assignees().then(setAssignees).catch(console.error)
    customerApi.list(1, 100).then((r) => setCustomers(r.data || [])).catch(console.error)
  }, [])

  const filteredAccounts = accounts.filter((a) => !platform || a.platform === platform)

  const loadInbox = useCallback(async (accountId: string) => {
    if (!accountId) return
    setLoadingInbox(true)
    setError('')
    try {
      const data = await chatApi.inbox(accountId)
      setThreads(data.threads || [])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Lỗi tải hộp thư')
    } finally {
      setLoadingInbox(false)
    }
  }, [])

  const loadThread = useCallback(async (accountId: string, thread: ChatThread) => {
    setLoadingThread(true)
    setError('')
    try {
      const data = await chatApi.thread(accountId, thread.thread_id || thread.thread_key)
      setMessages([...(data.messages || [])].reverse())
      if (thread.conversation_id) {
        const conv = await chatApi.getConversation(thread.conversation_id)
        setConversation(conv)
      } else {
        setConversation(null)
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Lỗi tải tin nhắn')
    } finally {
      setLoadingThread(false)
    }
  }, [])

  useEffect(() => {
    if (selectedAccount) void loadInbox(selectedAccount)
  }, [selectedAccount, loadInbox])

  useEffect(() => {
    if (selectedThread && selectedAccount) void loadThread(selectedAccount, selectedThread)
  }, [selectedThread, selectedAccount, loadThread])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  useEffect(() => {
    if (!selectedAccount) return
    pollRef.current = window.setInterval(() => { void loadInbox(selectedAccount) }, 30000)
    return () => { if (pollRef.current) clearInterval(pollRef.current) }
  }, [selectedAccount, loadInbox])

  async function handleSend() {
    if (!selectedAccount || !selectedThread || !reply.trim()) return
    setSending(true)
    try {
      await chatApi.send({
        account_id: selectedAccount,
        thread_id: selectedThread.thread_id || selectedThread.thread_key,
        text: reply.trim(),
      })
      setReply('')
      await loadThread(selectedAccount, selectedThread)
      await loadInbox(selectedAccount)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Gửi thất bại')
    } finally {
      setSending(false)
    }
  }

  async function saveConversation(patch: { customer_id?: string; assigned_user_id?: string }) {
    if (!conversation?.id) return
    const updated = await chatApi.updateConversation(conversation.id, patch)
    setConversation(updated)
    if (selectedAccount) await loadInbox(selectedAccount)
  }

  const canManage = session?.role === 'admin' || session?.role === 'director' ||
    !conversation?.assigned_user_id || conversation.assigned_user_id === session?.id

  return (
    <div className="page-card inbox-page">
      <div className="page-header">
        <h1><MessageCircle size={22} style={{ verticalAlign: 'middle', marginRight: 8 }} />Hộp thư đa kênh</h1>
        <p className="inbox-sub">Theo dõi khách hàng qua Facebook Messenger và Zalo</p>
      </div>

      <div className="inbox-toolbar">
        <select className="field-input" value={platform} onChange={(e) => { setPlatform(e.target.value as Platform); setSelectedAccount(''); setSelectedThread(null) }}>
          <option value="">Tất cả nền tảng</option>
          <option value="facebook">Facebook</option>
          <option value="zalo">Zalo</option>
        </select>
        <select className="field-input" value={selectedAccount} onChange={(e) => { setSelectedAccount(e.target.value); setSelectedThread(null) }}>
          <option value="">Chọn tài khoản</option>
          {filteredAccounts.map((a) => (
            <option key={a.id} value={a.id}>{a.name} ({a.platform})</option>
          ))}
        </select>
        <Button variant="secondary" onClick={() => selectedAccount && loadInbox(selectedAccount)} disabled={!selectedAccount || loadingInbox}>
          <RefreshCw size={16} /> Tải lại
        </Button>
      </div>

      {error && <div className="inbox-error">{error}</div>}

      <div className="inbox-layout">
        <aside className="inbox-threads">
          {loadingInbox && <div className="inbox-loading"><Loader2 className="spin" size={18} /> Đang tải...</div>}
          {!loadingInbox && threads.length === 0 && (
            <p className="inbox-empty">{selectedAccount ? 'Chưa có hội thoại' : 'Chọn tài khoản để xem tin nhắn'}</p>
          )}
          {threads.map((t) => (
            <button
              key={t.thread_id || t.thread_key}
              type="button"
              className={`inbox-thread ${selectedThread?.thread_id === t.thread_id ? 'active' : ''}`}
              onClick={() => setSelectedThread(t)}
            >
              <div className="inbox-thread__title">{t.title || 'Khách'}</div>
              <div className="inbox-thread__preview">{t.last_message || '—'}</div>
              <div className="inbox-thread__meta">
                {t.customer_name && <span className="tag">{t.customer_name}</span>}
                {t.assigned_name && <span className="tag tag--muted">{t.assigned_name}</span>}
              </div>
            </button>
          ))}
        </aside>

        <section className="inbox-chat">
          {!selectedThread ? (
            <div className="inbox-placeholder">Chọn hội thoại để xem chi tiết</div>
          ) : (
            <>
              <header className="inbox-chat__header">
                <div>
                  <strong>{selectedThread.title}</strong>
                  <small>{platform || selectedThread.platform}</small>
                </div>
                {canManage && conversation && (
                  <div className="inbox-assign">
                    <select
                      className="field-input"
                      value={conversation.customer_id || ''}
                      onChange={(e) => void saveConversation({ customer_id: e.target.value || undefined })}
                    >
                      <option value="">Liên kết KH...</option>
                      {customers.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                    </select>
                    <select
                      className="field-input"
                      value={conversation.assigned_user_id || ''}
                      onChange={(e) => void saveConversation({ assigned_user_id: e.target.value || undefined })}
                    >
                      <option value="">Gán sale...</option>
                      {assignees.map((u) => <option key={u.id} value={u.id}>{u.full_name}</option>)}
                    </select>
                    {conversation.customer_id && (
                      <Link to={`/customers/${conversation.customer_id}`} className="inbox-customer-link">
                        <UserPlus size={14} /> Customer 360
                      </Link>
                    )}
                  </div>
                )}
              </header>

              <div className="inbox-messages">
                {loadingThread && <div className="inbox-loading"><Loader2 className="spin" size={18} /></div>}
                {messages.map((m) => (
                  <div key={m.message_id} className={`inbox-msg ${m.is_self ? 'self' : ''}`}>
                    {!m.is_self && <span className="inbox-msg__author">{m.sender_name || 'Khách'}</span>}
                    <div className="inbox-msg__bubble">{m.text}</div>
                  </div>
                ))}
                <div ref={bottomRef} />
              </div>

              <footer className="inbox-compose">
                <textarea
                  className="field-input"
                  rows={2}
                  placeholder="Nhập tin nhắn..."
                  value={reply}
                  onChange={(e) => setReply(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); void handleSend() } }}
                  disabled={!canManage}
                />
                <Button variant="primary" onClick={() => void handleSend()} disabled={sending || !canManage || !reply.trim()}>
                  <Send size={16} /> Gửi
                </Button>
              </footer>
            </>
          )}
        </section>
      </div>
    </div>
  )
}
