import { useCallback, useMemo, useState } from 'react'
import { campaignApi } from '@/entities/campaign/api/campaignApi'
import type { Campaign } from '@/entities/campaign/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { CampaignModal } from '@/features/campaign-modal/ui/CampaignModal'
import { campaignStatusLabel } from '@/shared/lib/labels'
import { CampaignLogsModal, CampaignScheduleModal } from '@/features/campaign-actions/ui/CampaignActionsModals'

export function CampaignTable() {
  const [modal, setModal] = useState<Campaign | Record<string, never> | null>(null)
  const [scheduleId, setScheduleId] = useState<string | null>(null)
  const [logsCampaign, setLogsCampaign] = useState<Campaign | null>(null)

  const fetchPage = useCallback(
    (page: number, limit: number) => campaignApi.list(page, limit),
    [],
  )
  const { rows, meta, page, setPage, loading, reload } = usePaginated<Campaign>(fetchPage)

  const columns = useMemo<DataTableColumn<Campaign>[]>(() => [
    { key: 'name', label: 'Tên chiến dịch' },
    { key: 'type', label: 'Loại' },
    { key: 'status', label: 'Trạng thái', render: (r) => campaignStatusLabel(r.status) },
    { key: 'sent_count', label: 'Đã gửi' },
    { key: 'scheduled_at', label: 'Lên lịch', render: (r) => r.scheduled_at ? r.scheduled_at.slice(0, 16).replace('T', ' ') : '—' },
    {
      key: '_actions', label: '', render: (r) => (
        <div className="row-actions">
          <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button>
          {r.status === 'draft' && (
            <>
              <Button variant="secondary" onClick={() => setScheduleId(r.id)}>Lên lịch</Button>
              <Button variant="primary" onClick={async () => { await campaignApi.send(r.id); reload() }}>Gửi</Button>
            </>
          )}
          <Button variant="secondary" onClick={() => setLogsCampaign(r)}>Logs</Button>
        </div>
      ),
    },
  ], [reload])

  return (
    <>
      <div className="page-header">
        <h1>Marketing</h1>
        <Button variant="primary" onClick={() => setModal({})}>+ Chiến dịch</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <CampaignModal
        open={modal !== null}
        edit={modal && 'id' in modal && modal.id ? (modal as Campaign) : null}
        onClose={() => setModal(null)}
        onSaved={reload}
      />
      <CampaignScheduleModal
        open={scheduleId !== null}
        campaignId={scheduleId}
        onClose={() => setScheduleId(null)}
        onSaved={reload}
      />
      <CampaignLogsModal
        open={logsCampaign !== null}
        campaignId={logsCampaign?.id ?? null}
        campaignName={logsCampaign?.name}
        onClose={() => setLogsCampaign(null)}
      />
    </>
  )
}
