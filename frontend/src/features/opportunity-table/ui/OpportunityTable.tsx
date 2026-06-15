import { useCallback, useMemo, useState } from 'react'
import { opportunityApi } from '@/entities/opportunity/api/opportunityApi'
import type { Opportunity } from '@/entities/opportunity/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { OpportunityModal } from '@/features/opportunity-modal/ui/OpportunityModal'

import { OPPORTUNITY_STAGE_OPTIONS, opportunityStageLabel } from '@/shared/lib/labels'

type OpportunityFilters = { stage: string }

export function OpportunityTable() {
  const [stage, setStage] = useState('')
  const [modal, setModal] = useState<Opportunity | Record<string, never> | null>(null)

  const columns = useMemo<DataTableColumn<Opportunity>[]>(() => [
    { key: 'code', label: 'Mã', render: (r) => r.code || '—' },
    { key: 'title', label: 'Tiêu đề' },
    { key: 'customer', label: 'Khách hàng', render: (r) => r.customer?.name || '—' },
    { key: 'stage', label: 'Giai đoạn', render: (r) => <span className={`badge badge--${r.stage === 'won' ? 'gold' : 'open'}`}>{opportunityStageLabel(r.stage)}</span> },
    { key: 'value', label: 'Giá trị', render: (r) => r.value ? `${Number(r.value).toLocaleString('vi-VN')} ₫` : '—' },
    { key: 'expected_close', label: 'Dự kiến đóng', render: (r) => r.expected_close ? r.expected_close.slice(0, 10) : '—' },
    { key: '_actions', label: '', render: (r) => <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button> },
  ], [])

  const fetchPage = useCallback(
    (page: number, limit: number, filters: OpportunityFilters) => opportunityApi.list(page, limit, filters),
    [],
  )

  const { rows, meta, page, setPage, loading, reload } = usePaginated<Opportunity, OpportunityFilters>(
    fetchPage,
    { filters: { stage } },
  )

  return (
    <>
      <div className="page-header">
        <h1>Cơ hội bán hàng</h1>
        <Button variant="primary" onClick={() => setModal({})}>+ Thêm cơ hội</Button>
      </div>
      <div className="toolbar">
        <select className="field-input" value={stage} onChange={(e) => { setStage(e.target.value); setPage(1) }}>
          <option value="">Tất cả giai đoạn</option>
          {OPPORTUNITY_STAGE_OPTIONS.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
        </select>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <OpportunityModal
        open={modal !== null}
        edit={modal && 'id' in modal && modal.id ? (modal as Opportunity) : null}
        onClose={() => setModal(null)}
        onSaved={reload}
      />
    </>
  )
}
