import { useCallback, useState } from 'react'
import { campaignApi } from '../../../entities/campaign/api/campaignApi'
import { usePaginated } from '../../../shared/hooks/usePaginated'
import { useToken } from '../../../app/providers/AuthProvider'
import { DataTable } from '../../../shared/ui/DataTable/DataTable'
import { Pagination } from '../../../shared/ui/Pagination/Pagination'
import { Button } from '../../../shared/ui/Button/Button'
import { CampaignModal } from '../../campaign-modal/ui/CampaignModal'

export function CampaignTable() {
  const token = useToken()
  const [modal, setModal] = useState(null)

  const fetchPage = useCallback((page, limit) => campaignApi.list(token, page, limit), [token])
  const { rows, meta, page, setPage, loading, reload } = usePaginated(fetchPage)

  const columns = [
    { key: 'name', label: 'Tên chiến dịch' },
    { key: 'type', label: 'Loại' },
    { key: 'status', label: 'Trạng thái' },
    { key: 'sent_count', label: 'Đã gửi' },
    {
      key: '_actions', label: '', render: (r) => (
        <div className="row-actions">
          <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button>
          {r.status === 'draft' && (
            <Button variant="primary" onClick={async () => { await campaignApi.send(token, r.id); reload() }}>Gửi</Button>
          )}
        </div>
      ),
    },
  ]

  return (
    <>
      <div className="page-header">
        <h1>Marketing</h1>
        <Button variant="primary" onClick={() => setModal({})}>+ Chiến dịch</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <CampaignModal open={!!modal} edit={modal?.id ? modal : null} onClose={() => setModal(null)} onSaved={reload} />
    </>
  )
}
