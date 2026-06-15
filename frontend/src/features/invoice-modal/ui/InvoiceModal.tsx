import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { CustomerSelect } from '@/shared/ui/CustomerSelect/CustomerSelect'
import { invoiceApi } from '@/entities/invoice/api/invoiceApi'
import type { Invoice } from '@/entities/invoice/model/types'
import { useToken } from '@/app/providers/AuthProvider'

interface LineItemForm {
  description: string
  qty: number | string
  unit_price: number | string
}

interface InvoiceFormState {
  customer_id: string
  contract_id: string
  tax_rate: string | number
  currency: string
  due_date: string
  items: LineItemForm[]
}

const emptyItem: LineItemForm = { description: '', qty: 1, unit_price: 0 }

function calcTotal(items: LineItemForm[]): number {
  return items.reduce((s, it) => s + (Number(it.qty) || 0) * (Number(it.unit_price) || 0), 0)
}

interface InvoiceModalProps {
  open: boolean
  onClose: () => void
  onSaved?: () => void
  edit: Invoice | null
}

export function InvoiceModal({ open, onClose, onSaved, edit }: InvoiceModalProps) {
  const token = useToken()
  const [form, setForm] = useState<InvoiceFormState>({
    customer_id: '', contract_id: '', tax_rate: 10, currency: 'VND', due_date: '', items: [{ ...emptyItem }],
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return
    if (edit?.id) {
      setForm({
        customer_id: edit.customer_id || '',
        contract_id: '',
        tax_rate: edit.tax_rate ?? 10,
        currency: edit.currency || 'VND',
        due_date: edit.due_date ? edit.due_date.slice(0, 10) : '',
        items: edit.items?.length
          ? edit.items.map((it) => ({ description: it.description, qty: it.qty, unit_price: it.unit_price }))
          : [{ ...emptyItem }],
      })
    } else {
      setForm({
        customer_id: '', contract_id: '', tax_rate: 10, currency: 'VND', due_date: '', items: [{ ...emptyItem }],
      })
    }
  }, [open, edit])

  const setItem = (i: number, k: keyof LineItemForm, v: string | number) => setForm((f) => {
    const items = [...f.items]
    items[i] = { ...items[i], [k]: v } as LineItemForm
    return { ...f, items }
  })

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!token) return
    setLoading(true)
    setError('')
    try {
      const items = form.items.map((it) => ({
        description: it.description,
        qty: Number(it.qty) || 1,
        unit_price: Number(it.unit_price) || 0,
        amount: (Number(it.qty) || 1) * (Number(it.unit_price) || 0),
      }))
      const body = {
        invoice: {
          customer_id: form.customer_id,
          contract_id: form.contract_id || null,
          tax_rate: Number(form.tax_rate) || 0,
          currency: form.currency,
          due_date: form.due_date || null,
          status: edit?.status || 'draft',
        },
        items,
      }
      if (edit?.id) await invoiceApi.update(token, edit.id, body)
      else await invoiceApi.create(token, body)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  async function handleSend() {
    if (!token || !edit?.id) return
    setLoading(true)
    setError('')
    try {
      await invoiceApi.send(token, edit.id)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  async function handleDownload() {
    if (!token || !edit?.id) return
    setLoading(true)
    setError('')
    try {
      const blob = await invoiceApi.download(token, edit.id)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `invoice-${edit.code}.pdf`
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  async function handleCancel() {
    if (!token || !edit?.id) return
    setLoading(true)
    setError('')
    try {
      await invoiceApi.cancel(token, edit.id)
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const subtotal = calcTotal(form.items)
  const taxAmount = subtotal * (Number(form.tax_rate) || 0) / 100

  return (
    <Modal open={open} onClose={onClose} title={edit ? 'Sửa hóa đơn' : 'Tạo hóa đơn'} wide>
      <form onSubmit={submit}>
        <div className="form-grid">
          <Field label="Khách hàng" required>
            <CustomerSelect value={form.customer_id} onChange={(v) => setForm((f) => ({ ...f, customer_id: v }))} required />
          </Field>
          <Field label="Hạn thanh toán"><Input type="date" value={form.due_date} onChange={(e) => setForm((f) => ({ ...f, due_date: e.target.value }))} /></Field>
          <Field label="Thuế (%)"><Input type="number" value={form.tax_rate} onChange={(e) => setForm((f) => ({ ...f, tax_rate: e.target.value }))} /></Field>
        </div>
        <div className="line-items">
          <strong>Dòng hàng</strong>
          {form.items.map((it, i) => (
            <div key={i} className="line-item-row">
              <Input placeholder="Mô tả" value={it.description} onChange={(e) => setItem(i, 'description', e.target.value)} />
              <Input type="number" placeholder="SL" value={it.qty} onChange={(e) => setItem(i, 'qty', e.target.value)} style={{ width: 70 }} />
              <Input type="number" placeholder="Đơn giá" value={it.unit_price} onChange={(e) => setItem(i, 'unit_price', e.target.value)} style={{ width: 120 }} />
              {form.items.length > 1 && (
                <Button variant="secondary" onClick={() => setForm((f) => ({ ...f, items: f.items.filter((_, j) => j !== i) }))}>×</Button>
              )}
            </div>
          ))}
          <Button variant="secondary" onClick={() => setForm((f) => ({ ...f, items: [...f.items, { ...emptyItem }] }))}>+ Dòng</Button>
          <p className="line-total">
            Tạm tính: {subtotal.toLocaleString('vi-VN')} ₫ · Thuế: {taxAmount.toLocaleString('vi-VN')} ₫ ·
            Tổng: {(subtotal + taxAmount).toLocaleString('vi-VN')} ₫
          </p>
        </div>
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          {edit?.id && edit.status === 'draft' && (
            <Button variant="secondary" onClick={handleSend} disabled={loading}>Gửi</Button>
          )}
          {edit?.id && edit.status !== 'cancelled' && (
            <Button variant="secondary" onClick={handleDownload} disabled={loading}>Tải PDF</Button>
          )}
          {edit?.id && edit.status !== 'cancelled' && edit.status !== 'paid' && (
            <Button variant="secondary" onClick={handleCancel} disabled={loading}>Hủy HĐ</Button>
          )}
          <Button variant="secondary" onClick={onClose}>Đóng</Button>
          <Button type="submit" variant="primary" disabled={loading}>{loading ? 'Đang lưu...' : 'Lưu'}</Button>
        </div>
      </form>
    </Modal>
  )
}
