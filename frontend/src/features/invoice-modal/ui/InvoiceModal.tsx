import { useEffect, useState, type FormEvent } from 'react'
import { Download, FileText, Plus, Send, Trash2 } from 'lucide-react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input, Select } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { CustomerSelect } from '@/shared/ui/CustomerSelect/CustomerSelect'
import { invoiceApi } from '@/entities/invoice/api/invoiceApi'
import type { Invoice } from '@/entities/invoice/model/types'
import { invoiceStatusLabel } from '@/shared/lib/labels'
import './invoice-modal.css'

interface LineItemForm {
  description: string
  unit: string
  qty: number | string
  unit_price: number | string
}

interface InvoiceFormState {
  customer_id: string
  contract_id: string
  tax_rate: string | number
  currency: string
  due_date: string
  note: string
  items: LineItemForm[]
}

const emptyItem: LineItemForm = { description: '', unit: 'Chuyến', qty: 1, unit_price: 0 }

function lineAmount(it: LineItemForm): number {
  return (Number(it.qty) || 0) * (Number(it.unit_price) || 0)
}

function calcSubtotal(items: LineItemForm[]): number {
  return items.reduce((s, it) => s + lineAmount(it), 0)
}

function fmtVnd(n: number): string {
  return `${n.toLocaleString('vi-VN')} ₫`
}

interface InvoiceModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
  edit: Invoice | null
}

export function InvoiceModal({ open, onClose, onSaved, edit }: InvoiceModalProps) {
  const [form, setForm] = useState<InvoiceFormState>({
    customer_id: '', contract_id: '', tax_rate: 10, currency: 'VND', due_date: '', note: '', items: [{ ...emptyItem }],
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    setError('')
    if (edit?.id) {
      setForm({
        customer_id: edit.customer_id || '',
        contract_id: '',
        tax_rate: edit.tax_rate ?? 10,
        currency: edit.currency || 'VND',
        due_date: edit.due_date ? edit.due_date.slice(0, 10) : '',
        note: '',
        items: edit.items?.length
          ? edit.items.map((it) => ({
            description: it.description,
            unit: 'Chuyến',
            qty: it.qty,
            unit_price: it.unit_price,
          }))
          : [{ ...emptyItem }],
      })
    } else {
      setForm({
        customer_id: '', contract_id: '', tax_rate: 10, currency: 'VND',
        due_date: new Date(Date.now() + 30 * 86400000).toISOString().slice(0, 10),
        note: '', items: [{ ...emptyItem }],
      })
    }
  }, [open, edit])

  const setItem = (i: number, k: keyof LineItemForm, v: string | number) => setForm((f) => {
    const items = [...f.items]
    items[i] = { ...items[i], [k]: v } as LineItemForm
    return { ...f, items }
  })

  const subtotal = calcSubtotal(form.items)
  const taxRate = Number(form.tax_rate) || 0
  const taxAmount = subtotal * taxRate / 100
  const total = subtotal + taxAmount

  async function submit(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const items = form.items.map((it) => ({
        description: it.description,
        qty: Number(it.qty) || 1,
        unit_price: Number(it.unit_price) || 0,
        amount: lineAmount(it),
      }))
      const body = {
        invoice: {
          customer_id: form.customer_id,
          contract_id: form.contract_id || null,
          tax_rate: taxRate,
          currency: form.currency,
          due_date: form.due_date || null,
          status: edit?.status || 'draft',
        },
        items,
      }
      if (edit?.id) await invoiceApi.update(edit.id, body)
      else await invoiceApi.create(body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  async function handleSend() {
    if (!edit?.id) return
    setLoading(true)
    setError('')
    try {
      await invoiceApi.send(edit.id)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  async function handleDownload() {
    if (!edit?.id) return
    setLoading(true)
    setError('')
    try {
      const blob = await invoiceApi.download(edit.id)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `hoa-don-${edit.code}.pdf`
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Không tải được PDF')
    } finally {
      setLoading(false)
    }
  }

  async function handleCancel() {
    if (!edit?.id || !window.confirm('Hủy hóa đơn này?')) return
    setLoading(true)
    setError('')
    try {
      await invoiceApi.cancel(edit.id)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={edit ? 'Sửa hóa đơn GTGT' : 'Tạo hóa đơn GTGT'}
      xlarge
      icon={FileText}
      tone="violet"
    >
      <form onSubmit={submit} className="invoice-modal__form">
        {edit?.id && (
          <div className="invoice-modal__meta">
            <span className="invoice-modal__code">{edit.code}</span>
            <span className={`badge badge--${edit.status === 'paid' ? 'gold' : 'open'}`}>
              {invoiceStatusLabel(edit.status)}
            </span>
          </div>
        )}

        <div className="invoice-modal__layout">
          <div>
            <section className="invoice-modal__section">
              <h4 className="invoice-modal__section-title">Thông tin chung</h4>
              <div className="invoice-modal__info-card">
                <div className="form-grid">
                <Field label="Khách hàng" required>
                  <CustomerSelect
                    value={form.customer_id}
                    onChange={(v) => setForm((f) => ({ ...f, customer_id: v }))}
                    required
                  />
                </Field>
                <Field label="Hạn thanh toán">
                  <Input type="date" value={form.due_date} onChange={(e) => setForm((f) => ({ ...f, due_date: e.target.value }))} />
                </Field>
                <Field label="Loại tiền">
                  <Select value={form.currency} onChange={(e) => setForm((f) => ({ ...f, currency: e.target.value }))}>
                    <option value="VND">VND</option>
                    <option value="USD">USD</option>
                  </Select>
                </Field>
                <Field label="Thuế GTGT (%)">
                  <Input type="number" min={0} max={100} step={0.1} value={form.tax_rate} onChange={(e) => setForm((f) => ({ ...f, tax_rate: e.target.value }))} />
                </Field>
                <div className="invoice-modal__note-field">
                  <Field label="Ghi chú">
                    <Input value={form.note} onChange={(e) => setForm((f) => ({ ...f, note: e.target.value }))} placeholder="Ghi chú nội bộ (không in lên hóa đơn)" />
                  </Field>
                </div>
                </div>
              </div>
            </section>

            <section className="invoice-modal__section">
              <h4 className="invoice-modal__section-title">Chi tiết dịch vụ</h4>
              <div className="invoice-lines-wrap">
              <div className="invoice-lines">
                <div className="invoice-lines__head">
                  <span>#</span>
                  <span>Mô tả</span>
                  <span>ĐVT</span>
                  <span>SL</span>
                  <span>Đơn giá</span>
                  <span>Tiền</span>
                  <span />
                </div>
                {form.items.map((it, i) => (
                  <div key={i} className="invoice-lines__row">
                    <span className="invoice-lines__idx">{i + 1}</span>
                    <Input
                      placeholder="Mô tả dịch vụ vận chuyển"
                      value={it.description}
                      onChange={(e) => setItem(i, 'description', e.target.value)}
                      required
                    />
                    <Input value={it.unit} onChange={(e) => setItem(i, 'unit', e.target.value)} />
                    <Input type="number" min={0} step={1} value={it.qty} onChange={(e) => setItem(i, 'qty', e.target.value)} />
                    <Input type="number" min={0} value={it.unit_price} onChange={(e) => setItem(i, 'unit_price', e.target.value)} />
                    <span className="invoice-lines__amount">{fmtVnd(lineAmount(it))}</span>
                    {form.items.length > 1 ? (
                      <button type="button" className="invoice-lines__remove" onClick={() => setForm((f) => ({ ...f, items: f.items.filter((_, j) => j !== i) }))} aria-label="Xóa dòng">
                        <Trash2 size={16} />
                      </button>
                    ) : <span />}
                  </div>
                ))}
                <div className="invoice-lines__add">
                  <Button type="button" variant="secondary" onClick={() => setForm((f) => ({ ...f, items: [...f.items, { ...emptyItem }] }))}>
                    <Plus size={16} /> Thêm dòng hàng
                  </Button>
                </div>
              </div>
              </div>
            </section>
          </div>

          <aside className="invoice-summary">
            <h4>Tổng hợp thanh toán</h4>
            <div className="invoice-summary__row">
              <span>Tạm tính</span>
              <strong>{fmtVnd(subtotal)}</strong>
            </div>
            <div className="invoice-summary__row">
              <span>Thuế GTGT ({taxRate}%)</span>
              <strong>{fmtVnd(taxAmount)}</strong>
            </div>
            <div className="invoice-summary__row invoice-summary__total">
              <span>Tổng cộng</span>
              <strong>{fmtVnd(total)}</strong>
            </div>
            {form.due_date && (
              <p className="invoice-summary__due">
                Hạn thanh toán
                <strong>{new Date(form.due_date).toLocaleDateString('vi-VN')}</strong>
              </p>
            )}
          </aside>
        </div>

        {error && <p className="form-error">{error}</p>}

        <div className="invoice-modal__actions">
          <div className="invoice-modal__actions-left">
            {edit?.id && edit.status !== 'cancelled' && (
              <Button type="button" variant="secondary" onClick={handleDownload} disabled={loading}>
                <Download size={16} /> Tải PDF
              </Button>
            )}
            {edit?.id && edit.status === 'draft' && (
              <Button type="button" variant="secondary" onClick={handleSend} disabled={loading}>
                <Send size={16} /> Gửi khách
              </Button>
            )}
            {edit?.id && edit.status !== 'cancelled' && edit.status !== 'paid' && (
              <Button type="button" variant="secondary" onClick={handleCancel} disabled={loading}>Hủy HĐ</Button>
            )}
          </div>
          <Button type="button" variant="secondary" onClick={onClose}>Đóng</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu hóa đơn'}</Button>
        </div>
      </form>
    </Modal>
  )
}
