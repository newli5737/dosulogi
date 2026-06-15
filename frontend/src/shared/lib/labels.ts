const OPPORTUNITY_PIPELINE: Record<string, string> = {
  lead: 'Tiềm năng',
  qualified: 'Đã đánh giá',
  proposal: 'Báo giá',
  negotiation: 'Đàm phán',
}

const OPPORTUNITY_OUTCOMES: Record<string, string> = {
  won: 'Chốt thắng',
  lost: 'Bỏ cuộc',
}

const OPPORTUNITY_STAGES: Record<string, string> = {
  ...OPPORTUNITY_PIPELINE,
  ...OPPORTUNITY_OUTCOMES,
}

const SERVICE_TYPES: Record<string, string> = {
  FCL: 'FCL — Nguyên cont',
  LCL: 'LCL — Ghép cont',
  air: 'Hàng không',
  express: 'Chuyển phát nhanh',
  road: 'Đường bộ',
  domestic: 'Vận chuyển nội địa',
  cold_chain: 'Chuỗi lạnh (Cold chain)',
  warehouse: 'Kho bãi / 3PL',
  last_mile: 'Giao last-mile',
  sea: 'Đường biển',
}

const INVOICE_STATUSES: Record<string, string> = {
  draft: 'Nháp',
  sent: 'Đã gửi',
  paid: 'Đã thanh toán',
  overdue: 'Quá hạn',
  cancelled: 'Đã hủy',
}

const SHIPMENT_STATUSES: Record<string, string> = {
  pending: 'Chờ lấy hàng',
  picked_up: 'Đã lấy hàng',
  in_transit: 'Đang vận chuyển',
  out_for_delivery: 'Đang giao',
  delivered: 'Đã giao',
  failed: 'Giao thất bại',
  returned: 'Hoàn trả',
  cancelled: 'Đã hủy',
}

const TICKET_STATUSES: Record<string, string> = {
  open: 'Mở',
  in_progress: 'Đang xử lý',
  pending: 'Chờ phản hồi',
  resolved: 'Đã giải quyết',
  closed: 'Đóng',
}

const TICKET_PRIORITIES: Record<string, string> = {
  low: 'Thấp',
  medium: 'Trung bình',
  high: 'Cao',
  urgent: 'Khẩn cấp',
}

const CONTRACT_STATUSES: Record<string, string> = {
  draft: 'Nháp',
  active: 'Hiệu lực',
  expired: 'Hết hạn',
  terminated: 'Chấm dứt',
  cancelled: 'Đã hủy',
}

const QUOTATION_STATUSES: Record<string, string> = {
  draft: 'Nháp',
  sent: 'Đã gửi',
  accepted: 'Chấp nhận',
  rejected: 'Từ chối',
  expired: 'Hết hạn',
  converted: 'Đã chuyển HĐ',
}

const CAMPAIGN_TYPES: Record<string, string> = {
  email: 'Email marketing',
  sms: 'SMS / ZNS',
  zalo_oa: 'Zalo OA',
  zalo_zns: 'Zalo ZNS (template)',
  facebook: 'Facebook Messenger',
  push: 'Push notification (app)',
  in_app: 'Thông báo in-app',
  webhook: 'Webhook / API',
  multi: 'Đa kênh (Email + SMS + Zalo)',
}

const CAMPAIGN_STATUSES: Record<string, string> = {
  draft: 'Nháp',
  scheduled: 'Đã lên lịch',
  sending: 'Đang gửi',
  sent: 'Đã gửi',
  cancelled: 'Đã hủy',
}

const CHAT_ACCOUNT_STATUSES: Record<string, string> = {
  active: 'Hoạt động',
  inactive: 'Ngưng',
  error: 'Lỗi',
  connected: 'Đã kết nối',
  pending: 'Chờ kết nối',
  starting: 'Đang khởi tạo',
  qr_ready: 'Quét QR',
  scanned: 'Đã quét',
  failed: 'Thất bại',
}

function label(map: Record<string, string>, key?: string | null, fallback = '—'): string {
  if (!key) return fallback
  return map[key] ?? key
}

export function isOpportunityClosed(stage?: string | null): boolean {
  return stage === 'won' || stage === 'lost'
}

export function opportunityStageLabel(stage?: string | null): string {
  return label(OPPORTUNITY_STAGES, stage)
}

export function opportunityPipelineLabel(stage?: string | null): string {
  if (!stage || isOpportunityClosed(stage)) return '—'
  return label(OPPORTUNITY_PIPELINE, stage)
}

export function opportunityOutcomeLabel(stage?: string | null): string {
  if (!isOpportunityClosed(stage)) return '—'
  return label(OPPORTUNITY_OUTCOMES, stage)
}

export function serviceTypeLabel(type?: string | null): string {
  return label(SERVICE_TYPES, type)
}

export function invoiceStatusLabel(status?: string | null): string {
  return label(INVOICE_STATUSES, status)
}

export function shipmentStatusLabel(status?: string | null): string {
  return label(SHIPMENT_STATUSES, status)
}

export function ticketStatusLabel(status?: string | null): string {
  return label(TICKET_STATUSES, status)
}

export function ticketPriorityLabel(priority?: string | null): string {
  return label(TICKET_PRIORITIES, priority)
}

export function contractStatusLabel(status?: string | null): string {
  return label(CONTRACT_STATUSES, status)
}

export function quotationStatusLabel(status?: string | null): string {
  return label(QUOTATION_STATUSES, status)
}

export function campaignTypeLabel(type?: string | null): string {
  return label(CAMPAIGN_TYPES, type)
}

export function campaignStatusLabel(status?: string | null): string {
  return label(CAMPAIGN_STATUSES, status)
}

export function chatAccountStatusLabel(status?: string | null): string {
  return label(CHAT_ACCOUNT_STATUSES, status)
}

export function statusCountLabel(status: string, domain: 'ticket' | 'shipment'): string {
  if (domain === 'ticket') return ticketStatusLabel(status)
  return shipmentStatusLabel(status)
}

export const CAMPAIGN_TYPE_OPTIONS = Object.entries(CAMPAIGN_TYPES).map(([value, label]) => ({ value, label }))
export const OPPORTUNITY_PIPELINE_OPTIONS = Object.entries(OPPORTUNITY_PIPELINE).map(([value, label]) => ({ value, label }))
export const OPPORTUNITY_OUTCOME_OPTIONS = Object.entries(OPPORTUNITY_OUTCOMES).map(([value, label]) => ({ value, label }))
export const OPPORTUNITY_STAGE_OPTIONS = Object.entries(OPPORTUNITY_STAGES).map(([value, label]) => ({ value, label }))
export const SERVICE_TYPE_OPTIONS = Object.entries(SERVICE_TYPES).map(([value, label]) => ({ value, label }))
export const TICKET_STATUS_OPTIONS = Object.entries(TICKET_STATUSES).map(([value, label]) => ({ value, label }))
export const TICKET_PRIORITY_OPTIONS = Object.entries(TICKET_PRIORITIES).map(([value, label]) => ({ value, label }))
export const INVOICE_STATUS_OPTIONS = Object.entries(INVOICE_STATUSES).map(([value, label]) => ({ value, label }))
export const CONTRACT_STATUS_OPTIONS = Object.entries(CONTRACT_STATUSES).map(([value, label]) => ({ value, label }))
export const QUOTATION_STATUS_OPTIONS = Object.entries(QUOTATION_STATUSES).map(([value, label]) => ({ value, label }))
