import { useEffect, useState } from 'react'
import { createCampaign, listCampaigns, paginatedItems, sendCampaign } from '../api'
import { DataTable, Page, useToken } from '../components/ui'

export default function CampaignsPage() {
  const token = useToken()
  const [rows, setRows] = useState([])
  const [form, setForm] = useState({ name: '', type: 'email_blast', subject: '', body_html: '<p>Xin chào</p>' })

  const load = () => listCampaigns(token).then((r) => setRows(paginatedItems(r)))

  useEffect(() => { load().catch(console.error) }, [token])

  async function submit(e) {
    e.preventDefault()
    await createCampaign(token, form)
    load()
  }

  return (
    <Page title="Marketing · Campaigns">
      <form className="inline-form" onSubmit={submit}>
        <input placeholder="Tên campaign" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
        <input placeholder="Subject" value={form.subject} onChange={(e) => setForm({ ...form, subject: e.target.value })} />
        <button type="submit">Tạo</button>
      </form>
      <DataTable
        columns={[
          { key: 'name', label: 'Tên' },
          { key: 'type', label: 'Loại' },
          { key: 'status', label: 'Trạng thái' },
          { key: 'sent_count', label: 'Đã gửi' },
          { key: 'act', label: '', render: (r) => (
            <button type="button" className="btn-sm" onClick={() => sendCampaign(token, r.id).then(load)}>Gửi</button>
          )},
        ]}
        rows={rows}
      />
    </Page>
  )
}
