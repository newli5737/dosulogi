import { useCallback, useMemo, useState } from 'react'
import { contractApi } from '@/entities/contract/api/contractApi'
import type { Contract } from '@/entities/contract/model/types'
import { usePaginated } from '@/shared/hooks/usePaginated'
import { DataTable, type DataTableColumn } from '@/shared/ui/DataTable/DataTable'
import { Pagination } from '@/shared/ui/Pagination/Pagination'
import { Button } from '@/shared/ui/Button/Button'
import { contractStatusLabel, serviceTypeLabel } from '@/shared/lib/labels'
import { ContractModal } from '@/features/contract-modal/ui/ContractModal'

export function ContractTable() {
  const [modal, setModal] = useState<Contract | Record<string, never> | null>(null)

  const columns = useMemo<DataTableColumn<Contract>[]>(() => [
    { key: 'code', label: 'Mã' },
    { key: 'title', label: 'Tiêu đề', render: (r) => r.title || '—' },
    { key: 'status', label: 'Trạng thái', render: (r) => contractStatusLabel(r.status) },
    { key: 'service_type', label: 'Dịch vụ', render: (r) => serviceTypeLabel(r.service_type) },
    { key: 'value', label: 'Giá trị', render: (r) => r.value ? `${Number(r.value).toLocaleString('vi-VN')} ₫` : '—' },
    { key: '_actions', label: '', render: (r) => <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button> },
  ], [])

  const fetchPage = useCallback(
    (page: number, limit: number) => contractApi.list(page, limit),
    [],
  )
  const { rows, meta, page, setPage, loading, reload } = usePaginated<Contract>(fetchPage)

  return (
    <>
      <div className="page-header">
        <h1>Hợp đồng</h1>
        <Button variant="primary" onClick={() => setModal({})}>+ Thêm hợp đồng</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <ContractModal
        open={modal !== null}
        edit={modal && 'id' in modal && modal.id ? (modal as Contract) : null}
        onClose={() => setModal(null)}
        onSaved={reload}
      />
    </>
  )
}
