import { useEffect, useState } from 'react'
import { opportunityApi } from '@/entities/opportunity/api/opportunityApi'
import type { Opportunity } from '@/entities/opportunity/model/types'
import { isOpportunityClosed, opportunityPipelineLabel, opportunityOutcomeLabel } from '@/shared/lib/labels'
import { Select } from '../Form/Form'

interface OpportunitySelectProps {
  value: string
  onChange: (value: string) => void
  customerId?: string
}

function oppLabel(o: Opportunity): string {
  const stage = isOpportunityClosed(o.stage)
    ? opportunityOutcomeLabel(o.stage)
    : opportunityPipelineLabel(o.stage)
  const val = o.value ? ` · ${Number(o.value).toLocaleString('vi-VN')} ₫` : ''
  return `${o.code} — ${o.title} (${stage}${val})`
}

export function OpportunitySelect({ value, onChange, customerId }: OpportunitySelectProps) {
  const [items, setItems] = useState<Opportunity[]>([])

  useEffect(() => {
    void opportunityApi.list(1, 200).then((res) => setItems(res.data)).catch(() => setItems([]))
  }, [])

  const filtered = customerId
    ? items.filter((o) => o.customer_id === customerId)
    : items

  return (
    <Select value={value} onChange={(e) => onChange(e.target.value)}>
      <option value="">— Không liên kết —</option>
      {filtered.map((o) => (
        <option key={o.id} value={o.id}>{oppLabel(o)}</option>
      ))}
    </Select>
  )
}
