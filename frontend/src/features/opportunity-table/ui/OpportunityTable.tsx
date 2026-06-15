import { useCallback, useMemo, useState } from 'react'
import { opportunityApi } from '@/entities/opportunity/api/opportunityApi'
import type { Opportunity } from '@/entities/opportunity/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { useToken } from '@/app/providers/AuthProvider'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { OpportunityModal } from '@/features/opportunity-modal/ui/OpportunityModal'

const STAGES = ['lead', 'qualified', 'proposal', 'negotiation', 'won', 'lost'] as const

type OpportunityFilters = { stage: string }

export function OpportunityTable() {
  const token = useToken()
  const [stage, setStage] = useState('')
  const [modal, setModal] = useState<Opportunity | Record<string, never> | null>(null)

  const columns = useMemo<DataTableColumn<Opportunity>[]>(() => [
    { key: 'code', label: 'Mã', render: (r) => r.code || '—' },
    { key: 'title', label: 'Tiêu đề' },
    { key: 'customer', label: 'Khách hàng', render: (r) => r.customer?.name || '—' },
    { key: 'stage', label: 'Stage', render: (r) => <span className={`badge badge--${r.stage === 'won' ? 'gold' : 'open'}`}>{r.stage}</span> },
    { key: 'value', label: 'Giá trị', render: (r) => r.value ? `${Number(r.value).toLocaleString('vi-VN')} ₫` : '—' },
    { key: 'expected_close', label: 'Dự kiến đóng', render: (r) => r.expected_close ? r.expected_close.slice(0, 10) : '—' },
    { key: '_actions', label: '', render: (r) => <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button> },
  ], [])

  const fetchPage = useCallback(
    (page: number, limit: number, filters: OpportunityFilters) => opportunityApi.list(token!, page, limit, filters),
    [token],
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
          <option value="">Tất cả stage</option>
          {STAGES.map((s) => <option key={s} value={s}>{s}</option>)}
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
