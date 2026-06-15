import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { customerApi } from '@/entities/customer/api/customerApi'
import { contactApi } from '@/entities/contact/api/contactApi'
import { interactionApi } from '@/entities/interaction/api/interactionApi'
import type { CustomerDetail } from '@/entities/customer/model/types'
import type { Contact } from '@/entities/contact/model/types'
import type { Interaction } from '@/entities/interaction/model/types'
import { Button } from '@/shared/ui/Button/Button'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { ContactModal } from '@/features/contact-modal/ui/ContactModal'
import { InteractionModal } from '@/features/interaction-modal/ui/InteractionModal'
import { parseMeta } from '@/shared/api/types'
import './customer-detail.css'

export function CustomerDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [customer, setCustomer] = useState<CustomerDetail | null>(null)
  const [contacts, setContacts] = useState<Contact[]>([])
  const [interactions, setInteractions] = useState<Interaction[]>([])
  const [intPage, setIntPage] = useState(1)
  const [intMeta, setIntMeta] = useState({ page: 1, limit: 10, total: 0 })
  const [loading, setLoading] = useState(true)
  const [contactModal, setContactModal] = useState<Contact | null | 'new'>(null)
  const [interactionOpen, setInteractionOpen] = useState(false)

  const loadCustomer = useCallback(async () => {
    if (!id) return
    const res = await customerApi.get(id)
    setCustomer(res.data)
  }, [id])

  const loadContacts = useCallback(async () => {
    if (!id) return
    const res = await contactApi.list(id)
    setContacts(Array.isArray(res.data) ? res.data : [])
  }, [id])

  const loadInteractions = useCallback(async () => {
    if (!id) return
    const res = await interactionApi.list(id, intPage, 10)
    setInteractions(Array.isArray(res.data) ? res.data : [])
    setIntMeta(parseMeta(res, intPage, 10))
  }, [id, intPage])

  const reload = useCallback(async () => {
    setLoading(true)
    try {
      await Promise.all([loadCustomer(), loadContacts(), loadInteractions()])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }, [loadCustomer, loadContacts, loadInteractions])

  useEffect(() => { void reload() }, [reload])

  async function deleteContact(contactId: string) {
    if (!id || !window.confirm('Xóa liên hệ này?')) return
    await contactApi.remove(id, contactId)
    void reload()
  }

  const contactColumns: DataTableColumn<Contact>[] = [
    { key: 'name', label: 'Tên' },
    { key: 'role', label: 'Chức vụ', render: (r) => r.role || '—' },
    { key: 'email', label: 'Email', render: (r) => r.email || '—' },
    { key: 'phone', label: 'ĐT', render: (r) => r.phone || '—' },
    { key: 'is_primary', label: 'Chính', render: (r) => r.is_primary ? '✓' : '' },
    {
      key: '_actions', label: '', render: (r) => (
        <div className="row-actions">
          <Button variant="secondary" onClick={() => setContactModal(r)}>Sửa</Button>
          <Button variant="secondary" onClick={() => void deleteContact(r.id)}>Xóa</Button>
        </div>
      ),
    },
  ]

  const interactionColumns: DataTableColumn<Interaction>[] = [
    { key: 'occurred_at', label: 'Thời gian', render: (r) => r.occurred_at?.slice(0, 16).replace('T', ' ') || '—' },
    { key: 'channel', label: 'Kênh' },
    { key: 'direction', label: 'Hướng', render: (r) => r.direction || '—' },
    { key: 'summary', label: 'Nội dung' },
    { key: 'created_by', label: 'Người ghi', render: (r) => r.created_by?.full_name || '—' },
  ]

  if (loading && !customer) return <div className="page-card">Đang tải...</div>
  if (!customer) return <div className="page-card">Không tìm thấy khách hàng. <Link to="/customers">Quay lại</Link></div>

  return (
    <div className="page-card customer-detail">
      <div className="page-header">
        <div>
          <Link to="/customers" className="back-link">← Danh sách KH</Link>
          <h1>{customer.name} <small className="muted">{customer.code}</small></h1>
        </div>
      </div>

      <div className="detail-grid">
        <section className="detail-card">
          <h2 className="section-title">Thông tin</h2>
          <dl className="detail-dl">
            <dt>Loại</dt><dd>{customer.type}</dd>
            <dt>Email</dt><dd>{customer.email || '—'}</dd>
            <dt>Điện thoại</dt><dd>{customer.phone || '—'}</dd>
            <dt>Tỉnh/TP</dt><dd>{customer.province || '—'}</dd>
            <dt>MST</dt><dd>{customer.tax_code || '—'}</dd>
            <dt>Segment / Tier</dt><dd>{customer.segment} / {customer.tier}</dd>
            <dt>Liên hệ cuối</dt><dd>{customer.last_contact_at ? customer.last_contact_at.slice(0, 10) : '—'}</dd>
          </dl>
        </section>
        <section className="detail-card">
          <h2 className="section-title">Tổng quan</h2>
          <div className="kpi-mini">
            <div><strong>{customer.open_tickets}</strong><span>Ticket mở</span></div>
            <div><strong>{customer.active_contracts}</strong><span>HĐ active</span></div>
          </div>
          {customer.primary_contact && (
            <div className="primary-contact">
              <strong>Liên hệ chính:</strong> {customer.primary_contact.name}
              {customer.primary_contact.phone && ` · ${customer.primary_contact.phone}`}
            </div>
          )}
        </section>
      </div>

      <section className="detail-section">
        <div className="section-header">
          <h2 className="section-title">Liên hệ</h2>
          <Button variant="primary" onClick={() => setContactModal('new')}>+ Thêm</Button>
        </div>
        <DataTable columns={contactColumns} rows={contacts} empty="Chưa có liên hệ" />
      </section>

      <section className="detail-section">
        <div className="section-header">
          <h2 className="section-title">Lịch sử tương tác</h2>
          <Button variant="primary" onClick={() => setInteractionOpen(true)}>+ Ghi nhận</Button>
        </div>
        <DataTable columns={interactionColumns} rows={interactions} empty="Chưa có tương tác" />
        <Pagination page={intPage} limit={intMeta.limit} total={intMeta.total} onChange={setIntPage} />
      </section>

      {id && (
        <>
          <ContactModal
            open={contactModal !== null}
            customerId={id}
            edit={contactModal && contactModal !== 'new' ? contactModal : null}
            onClose={() => setContactModal(null)}
            onSaved={reload}
          />
          <InteractionModal
            open={interactionOpen}
            customerId={id}
            onClose={() => setInteractionOpen(false)}
            onSaved={reload}
          />
        </>
      )}
    </div>
  )
}
