# GPS Tracker Mock (bên thứ 3)

Giả lập nhà cung cấp tracking GPS cho Dosu Logi: API pull + webhook push có HMAC.

## Chạy local

```bash
go build -o bin/gps-tracker ./cmd/gps-tracker

set GPS_PORT=8091
set GPS_API_KEY=dev-gps-api-key
set GPS_WEBHOOK_URL=http://127.0.0.1:8089/api/v1/webhooks/tracking
set GPS_WEBHOOK_SECRET=dev-tracking-secret
set GPS_PUSH_INTERVAL_SEC=30

bin/gps-tracker
```

## API (xác thực `X-Api-Key`)

| Method | Path | Mô tả |
|--------|------|--------|
| GET | `/health` | Health check |
| GET | `/v1/shipments` | Danh sách xe demo |
| GET | `/v1/shipments/:tracking_code` | Chi tiết GPS + events |

## Webhook → Dosu Logi

Mỗi 30s (mặc định) service gửi POST tới `GPS_WEBHOOK_URL` với header:

- `Content-Type: application/json`
- `X-Hmac-Signature`: HMAC-SHA256 hex của body, secret = `GPS_WEBHOOK_SECRET`

Payload mẫu:

```json
{
  "tracking_code": "VD-DEMO-001",
  "status": "in_transit",
  "location": "51C-123.45 (13.8500, 109.0500)",
  "lat": 13.85,
  "lng": 109.05,
  "description": "Cập nhật GPS xe 51C-123.45",
  "event_time": "2026-06-15T10:00:00Z"
}
```

## Cấu hình Dosu Logi (.env)

```env
TRACKING_API_BASE_URL=http://127.0.0.1:8091
TRACKING_API_KEY=dev-gps-api-key
TRACKING_WEBHOOK_SECRET=dev-tracking-secret
```

Tạo vận đơn với mã `VD-DEMO-001`, `VD-DEMO-002`, `VD-DEMO-003` để nhận dữ liệu demo.

## VPS (systemd)

File mẫu: `deploy/systemd/dosulogi-gps-tracker.service`

```bash
go build -o /home/dosulogi/dosulogi/bin/gps-tracker ./cmd/gps-tracker
sudo cp deploy/systemd/dosulogi-gps-tracker.service /etc/systemd/system/
sudo systemctl enable --now dosulogi-gps-tracker
```
