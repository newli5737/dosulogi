import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Select, Textarea } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { ticketApi } from '@/entities/ticket/api/ticketApi'
import { TICKET_PRIORITY_OPTIONS, TICKET_STATUS_OPTIONS } from '@/shared/lib/labels'
import type { Ticket, TicketComment } from '@/entities/ticket/model/types'

interface TicketDetailModalProps {
  open: boolean
  ticketId: string | null
  onClose: () => void
  onSaved?: () => void
}

export function TicketDetailModal({ open, ticketId, onClose, onSaved }: TicketDetailModalProps) {
  const [ticket, setTicket] = useState<Ticket | null>(null)
  const [comments, setComments] = useState<TicketComment[]>([])
  const [status, setStatus] = useState('open')
  const [priority, setPriority] = useState('medium')
  const [commentBody, setCommentBody] = useState('')
  const [isInternal, setIsInternal] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open || !ticketId) return
    setLoading(true)
    ticketApi.get(ticketId)
      .then((res) => {
        setTicket(res.data.ticket)
        setComments(Array.isArray(res.data.comments) ? res.data.comments : [])
        setStatus(res.data.ticket.status)
        setPriority(res.data.ticket.priority)
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Error'))
      .finally(() => setLoading(false))
  }, [open, ticketId])

  async function saveTicket(e: FormEvent) {
    e.preventDefault()
    if (!ticketId) return
    setLoading(true)
    setError('')
    try {
      await ticketApi.update(ticketId, { status, priority })
      onSaved?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  async function postComment(e: FormEvent) {
    e.preventDefault()
    if (!ticketId || !commentBody.trim()) return
    setLoading(true)
    try {
      const res = await ticketApi.addComment(ticketId, { body: commentBody, is_internal: isInternal })
      setComments((c) => [...c, res.data])
      setCommentBody('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={ticket ? `Ticket ${ticket.code}` : 'Chi tiết ticket'} wide>
      {loading && !ticket ? <p>Đang tải...</p> : (
        <div className="ticket-detail">
          {ticket && (
            <>
              <div className="ticket-detail__meta">
                <p><strong>{ticket.title}</strong></p>
                <p className="muted">KH: {ticket.customer?.name || '—'} · SLA: {ticket.is_overdue ? '⚠ Quá hạn' : 'OK'}</p>
                {ticket.description && <p>{ticket.description}</p>}
              </div>
              <form onSubmit={saveTicket} className="form-grid" style={{ marginBottom: 16 }}>
                <Field label="Trạng thái">
                  <Select value={status} onChange={(e) => setStatus(e.target.value)}>
                    {TICKET_STATUS_OPTIONS.map((s) => (
                      <option key={s.value} value={s.value}>{s.label}</option>
                    ))}
                  </Select>
                </Field>
                <Field label="Ưu tiên">
                  <Select value={priority} onChange={(e) => setPriority(e.target.value)}>
                    {TICKET_PRIORITY_OPTIONS.map((p) => (
                      <option key={p.value} value={p.value}>{p.label}</option>
                    ))}
                  </Select>
                </Field>
                <div className="form-actions" style={{ gridColumn: '1 / -1' }}>
                  <Button type="submit" variant="primary" disabled={loading}>Cập nhật</Button>
                </div>
              </form>
            </>
          )}

          <h4 className="section-title">Bình luận</h4>
          <div className="comment-thread">
            {comments.map((c) => (
              <div key={c.id} className={`comment ${c.is_internal ? 'comment--internal' : ''}`}>
                <div className="comment__head">
                  <strong>{c.created_by?.full_name || 'Hệ thống'}</strong>
                  <span>{c.created_at?.slice(0, 16).replace('T', ' ')}</span>
                  {c.is_internal && <span className="badge badge--open">Nội bộ</span>}
                </div>
                <p>{c.body}</p>
              </div>
            ))}
            {!comments.length && <p className="muted">Chưa có bình luận</p>}
          </div>

          <form onSubmit={postComment} style={{ marginTop: 16 }}>
            <Field label="Thêm bình luận">
              <Textarea value={commentBody} onChange={(e) => setCommentBody(e.target.value)} rows={3} />
            </Field>
            <label style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 12 }}>
              <input type="checkbox" checked={isInternal} onChange={(e) => setIsInternal(e.target.checked)} />
              Bình luận nội bộ
            </label>
            <Button type="submit" variant="primary" disabled={loading || !commentBody.trim()}>Gửi</Button>
          </form>
          {error && <p className="form-error">{error}</p>}
        </div>
      )}
    </Modal>
  )
}
