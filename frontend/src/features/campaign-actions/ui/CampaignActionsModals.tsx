import { useEffect, useState } from 'react'
import { Modal } from '@/shared/ui/Modal/Modal'
import { Field, Input } from '@/shared/ui/Form/Form'
import { Button } from '@/shared/ui/Button/Button'
import { campaignApi } from '@/entities/campaign/api/campaignApi'
import type { CampaignLog } from '@/entities/campaign/model/types'
import { useToken } from '@/app/providers/AuthProvider'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { parseMeta } from '@/shared/api/types'

interface CampaignLogsModalProps {
  open: boolean
  campaignId: string | null
  campaignName?: string
  onClose: () => void
}

export function CampaignLogsModal({ open, campaignId, campaignName, onClose }: CampaignLogsModalProps) {
  const token = useToken()
  const [rows, setRows] = useState<CampaignLog[]>([])
  const [page, setPage] = useState(1)
  const [meta, setMeta] = useState({ page: 1, limit: 20, total: 0 })

  useEffect(() => {
    if (!open || !token || !campaignId) return
    campaignApi.logs(token, campaignId, page, 20)
      .then((res) => {
        setRows(Array.isArray(res.data) ? res.data : [])
        setMeta(parseMeta(res, page, 20))
      })
      .catch(console.error)
  }, [open, token, campaignId, page])

  const columns: DataTableColumn<CampaignLog>[] = [
    { key: 'email', label: 'Email', render: (r) => r.email || '—' },
    { key: 'status', label: 'Trạng thái', render: (r) => r.status || '—' },
    { key: 'created_at', label: 'Thời gian', render: (r) => r.created_at.slice(0, 16).replace('T', ' ') },
  ]

  return (
    <Modal open={open} onClose={onClose} title={`Log: ${campaignName || ''}`} wide>
      <DataTable columns={columns} rows={rows} empty="Chưa có log" />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
    </Modal>
  )
}

interface CampaignScheduleModalProps {
  open: boolean
  campaignId: string | null
  onClose: () => void
  onSaved?: () => void
}

export function CampaignScheduleModal({ open, campaignId, onClose, onSaved }: CampaignScheduleModalProps) {
  const token = useToken()
  const [at, setAt] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function submit() {
    if (!token || !campaignId || !at) return
    setLoading(true)
    setError('')
    try {
      await campaignApi.schedule(token, campaignId, new Date(at).toISOString())
      onSaved?.()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Lên lịch gửi">
      <Field label="Thời gian gửi" required>
        <Input type="datetime-local" value={at} onChange={(e) => setAt(e.target.value)} required />
      </Field>
      {error && <p className="form-error">{error}</p>}
      <div className="form-actions">
        <Button variant="secondary" onClick={onClose}>Hủy</Button>
        <Button variant="primary" onClick={() => void submit()} disabled={loading || !at}>
          {loading ? 'Đang lưu...' : 'Lên lịch'}
        </Button>
      </div>
    </Modal>
  )
}
