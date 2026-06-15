import { useEffect, useState } from 'react'
import { listShipments, paginatedItems, syncShipment } from '../api'
import { DataTable, Page, useToken } from '../components/ui'

export default function ShipmentsPage() {
  const token = useToken()
  const [rows, setRows] = useState([])

  const load = () => listShipments(token).then((r) => setRows(paginatedItems(r)))
  useEffect(() => { load().catch(console.error) }, [token])

  return (
    <Page title="Vận đơn">
      <DataTable
        columns={[
          { key: 'tracking_code', label: 'Mã tracking' },
          { key: 'status', label: 'Trạng thái' },
          { key: 'origin', label: 'Đi' },
          { key: 'destination', label: 'Đến' },
          { key: 'sync', label: '', render: (r) => (
            <button type="button" className="btn-sm" onClick={() => syncShipment(token, r.id).then(load)}>Sync</button>
          )},
        ]}
        rows={rows}
      />
    </Page>
  )
}
