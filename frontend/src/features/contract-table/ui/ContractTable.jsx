import { useCallback, useMemo, useState } from 'react'
import { contractApi } from '../../../entities/contract/api/contractApi'
import { usePaginated } from '../../../shared/hooks/usePaginated'
import { useToken } from '../../../app/providers/AuthProvider'
import { DataTable } from '../../../shared/ui/DataTable/DataTable'
import { Pagination } from '../../../shared/ui/Pagination/Pagination'
import { Button } from '../../../shared/ui/Button/Button'
import { ContractModal } from '../../contract-modal/ui/ContractModal'

export function ContractTable() {
  const token = useToken()
  const [modal, setModal] = useState(null)

  const columns = useMemo(() => [
    { key: 'code', label: 'Mã' },
    { key: 'title', label: 'Tiêu đề', render: (r) => r.title || '—' },
    { key: 'status', label: 'Trạng thái' },
    { key: 'service_type', label: 'Dịch vụ', render: (r) => r.service_type || '—' },
    { key: 'value', label: 'Giá trị', render: (r) => r.value ? `${Number(r.value).toLocaleString('vi-VN')} ₫` : '—' },
    { key: '_actions', label: '', render: (r) => <Button variant="secondary" onClick={() => setModal(r)}>Sửa</Button> },
  ], [])

  const fetchPage = useCallback((page, limit) => contractApi.list(token, page, limit), [token])
  const { rows, meta, page, setPage, loading, reload } = usePaginated(fetchPage)

  return (
    <>
      <div className="page-header">
        <h1>Hợp đồng</h1>
        <Button variant="primary" onClick={() => setModal({})}>+ Thêm hợp đồng</Button>
      </div>
      <DataTable columns={columns} rows={rows} loading={loading} />
      <Pagination page={page} limit={meta.limit} total={meta.total} onChange={setPage} />
      <ContractModal open={!!modal} edit={modal?.id ? modal : null} onClose={() => setModal(null)} onSaved={reload} />
    </>
  )
}
