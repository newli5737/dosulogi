# Tài liệu Kỹ thuật – Logistics ERP/CRM

**Stack:** Go (Gin) · PostgreSQL 16 · Redis · Nginx · VPS  
**Kiến trúc:** Monolith, chia package nội bộ  
**Tích hợp:** Tracking API (3rd party) · SePay Webhook · SendGrid · Leaflet/OSM

---

## 1. File / Folder Structure

```
/logistics-erp/
├── cmd/
│   └── server/
│       └── main.go                  # Entry point
├── internal/
│   ├── config/
│   │   └── config.go                # Load env, app config struct
│   ├── db/
│   │   ├── postgres.go              # Init PostgreSQL connection pool
│   │   └── redis.go                 # Init Redis client
│   ├── middleware/
│   │   ├── auth.go                  # JWT verify, inject user ctx
│   │   └── rbac.go                  # Role-based access check
│   ├── module/
│   │   ├── auth/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── crm/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── sales/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── marketing/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── accounting/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   └── tracking/
│   │       ├── handler.go           # Nhận webhook từ Tracking API
│   │       ├── poller.go            # Polling job (cron)
│   │       ├── service.go
│   │       └── repository.go
│   ├── integration/
│   │   ├── sepay/
│   │   │   └── webhook.go           # Parse + verify SePay payload
│   │   ├── sendgrid/
│   │   │   └── client.go            # Gửi email
│   │   └── tracking3p/
│   │       └── client.go            # Gọi Tracking API bên thứ 3
│   ├── router/
│   │   └── router.go                # Đăng ký tất cả routes
│   └── util/
│       ├── jwt.go                   # Sign / parse JWT
│       ├── password.go              # bcrypt hash/verify
│       └── pdf.go                   # Generate Invoice PDF
├── migrations/
│   └── *.sql                        # Migration files đánh số thứ tự
├── .env.example
├── go.mod
└── go.sum
```

---

## 2. Environment Variables (.env)

```env
# App
APP_PORT=8080
APP_ENV=production

# PostgreSQL
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=logistics_erp
DB_USER=erp_user
DB_PASSWORD=secret

# Redis
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=

# JWT
JWT_ACCESS_SECRET=<random-256bit>
JWT_REFRESH_SECRET=<random-256bit>
JWT_ACCESS_TTL_MIN=15
JWT_REFRESH_TTL_DAY=7

# SePay
SEPAY_WEBHOOK_SECRET=<from-sepay-dashboard>

# SendGrid
SENDGRID_API_KEY=SG.xxx
SENDGRID_FROM_EMAIL=no-reply@company.com

# Tracking 3rd party
TRACKING_API_BASE_URL=https://api.trackingprovider.com
TRACKING_API_KEY=xxx
TRACKING_POLL_INTERVAL_SEC=300
```

---

## 3. Database Schema

### 3.1 users
```sql
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,           -- bcrypt
    full_name   VARCHAR(255) NOT NULL,
    role        VARCHAR(50) NOT NULL,            -- admin | director | sales_manager | sales_rep | marketing | accountant
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN DEFAULT false,
    created_at  TIMESTAMPTZ DEFAULT now()
);
```

### 3.2 CRM
```sql
CREATE TABLE customers (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code             VARCHAR(50) UNIQUE NOT NULL,   -- KH-0001
    name             VARCHAR(255) NOT NULL,
    type             VARCHAR(20) NOT NULL,           -- B2B | B2C
    email            VARCHAR(255),
    phone            VARCHAR(20),
    address          TEXT,
    province         VARCHAR(100),
    segment          VARCHAR(50),                   -- enterprise | sme | individual
    tier             VARCHAR(20),                   -- gold | silver | standard
    assigned_to      UUID REFERENCES users(id),
    last_contact_at  TIMESTAMPTZ,
    created_by       UUID REFERENCES users(id),
    created_at       TIMESTAMPTZ DEFAULT now(),
    updated_at       TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE contacts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id  UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    role         VARCHAR(100),
    phone        VARCHAR(20),
    email        VARCHAR(255),
    is_primary   BOOLEAN DEFAULT false,
    created_at   TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE interactions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id  UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    channel      VARCHAR(30) NOT NULL,    -- call | email | meeting | note
    direction    VARCHAR(10),             -- inbound | outbound
    summary      TEXT,
    created_by   UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ DEFAULT now()
);
```

### 3.3 Sales
```sql
CREATE TABLE opportunities (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id   UUID NOT NULL REFERENCES customers(id),
    title         VARCHAR(255) NOT NULL,
    stage         VARCHAR(30) NOT NULL,     -- lead | qualified | proposal | negotiation | won | lost
    value         NUMERIC(15,2),
    currency      VARCHAR(3) DEFAULT 'VND',
    expected_close DATE,
    assigned_to   UUID REFERENCES users(id),
    lost_reason   TEXT,
    created_by    UUID REFERENCES users(id),
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE contracts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code           VARCHAR(50) UNIQUE NOT NULL,     -- HD-2026-001
    customer_id    UUID NOT NULL REFERENCES customers(id),
    opportunity_id UUID REFERENCES opportunities(id),
    title          VARCHAR(255),
    start_date     DATE NOT NULL,
    end_date       DATE,
    service_type   VARCHAR(100),                    -- FCL | LCL | air | express
    value          NUMERIC(15,2),
    currency       VARCHAR(3) DEFAULT 'VND',
    status         VARCHAR(20) DEFAULT 'draft',     -- draft | active | expired | terminated
    file_url       VARCHAR(500),                    -- uploaded contract PDF
    created_by     UUID REFERENCES users(id),
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE quotations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code         VARCHAR(50) UNIQUE NOT NULL,       -- BG-2026-001
    customer_id  UUID NOT NULL REFERENCES customers(id),
    opp_id       UUID REFERENCES opportunities(id),
    items        JSONB NOT NULL,                    -- [{description, qty, unit_price, amount}]
    total        NUMERIC(15,2),
    currency     VARCHAR(3) DEFAULT 'VND',
    valid_until  DATE,
    status       VARCHAR(20) DEFAULT 'draft',       -- draft | sent | accepted | rejected
    created_by   UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ DEFAULT now()
);
```

### 3.4 Tracking
```sql
CREATE TABLE shipments (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tracking_code      VARCHAR(100) UNIQUE NOT NULL,  -- mã từ 3rd party
    external_id        VARCHAR(255),                  -- ID bên tracking provider
    customer_id        UUID REFERENCES customers(id),
    contract_id        UUID REFERENCES contracts(id),
    status             VARCHAR(50),                   -- pending | in_transit | delivered | failed
    origin             VARCHAR(255),
    destination        VARCHAR(255),
    lat                NUMERIC(10,7),
    lng                NUMERIC(10,7),
    estimated_delivery DATE,
    actual_delivery    TIMESTAMPTZ,
    raw_payload        JSONB,                         -- full response từ API
    last_synced_at     TIMESTAMPTZ,
    created_at         TIMESTAMPTZ DEFAULT now(),
    updated_at         TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE shipment_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shipment_id  UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    status       VARCHAR(50),
    description  TEXT,
    location     VARCHAR(255),
    lat          NUMERIC(10,7),
    lng          NUMERIC(10,7),
    event_time   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ DEFAULT now()
);
```

### 3.5 Accounting
```sql
CREATE TABLE invoices (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code          VARCHAR(50) UNIQUE NOT NULL,       -- INV-2026-001
    customer_id   UUID NOT NULL REFERENCES customers(id),
    contract_id   UUID REFERENCES contracts(id),
    shipment_id   UUID REFERENCES shipments(id),
    items         JSONB NOT NULL,                    -- [{description, qty, unit_price, amount}]
    subtotal      NUMERIC(15,2),
    tax_rate      NUMERIC(5,2) DEFAULT 10,
    tax_amount    NUMERIC(15,2),
    total         NUMERIC(15,2),
    currency      VARCHAR(3) DEFAULT 'VND',
    status        VARCHAR(20) DEFAULT 'draft',       -- draft | sent | paid | overdue | cancelled
    due_date      DATE,
    paid_at       TIMESTAMPTZ,
    file_url      VARCHAR(500),                      -- generated PDF path
    created_by    UUID REFERENCES users(id),
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id      UUID NOT NULL REFERENCES invoices(id),
    amount          NUMERIC(15,2),
    method          VARCHAR(30),                     -- bank_transfer | cash
    reference_code  VARCHAR(255),                    -- nội dung chuyển khoản
    sepay_txn_id    VARCHAR(255),                    -- ID giao dịch từ SePay
    matched_auto    BOOLEAN DEFAULT false,
    note            TEXT,
    created_at      TIMESTAMPTZ DEFAULT now()
);
```

### 3.6 Marketing
```sql
CREATE TABLE campaigns (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    type         VARCHAR(30) NOT NULL,               -- email_blast | drip | transactional
    status       VARCHAR(20) DEFAULT 'draft',        -- draft | scheduled | sending | done | paused
    subject      VARCHAR(500),
    body_html    TEXT,
    segment      JSONB,                              -- filter: {tier, province, type, ...}
    scheduled_at TIMESTAMPTZ,
    sent_count   INT DEFAULT 0,
    created_by   UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE campaign_logs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id  UUID NOT NULL REFERENCES campaigns(id),
    customer_id  UUID REFERENCES customers(id),
    email        VARCHAR(255),
    status       VARCHAR(20),                        -- sent | delivered | opened | clicked | bounced
    sg_message_id VARCHAR(255),
    created_at   TIMESTAMPTZ DEFAULT now()
);
```

---

## 4. Authentication Flow

### 4.1 Login
```
POST /api/auth/login
Body: { email, password }

→ Verify password (bcrypt)
→ Generate Access Token (JWT, 15m): { sub: user_id, role, exp }
→ Generate Refresh Token (JWT, 7d): { sub: user_id, exp }
→ Hash refresh token → lưu vào refresh_tokens table
→ Response:
  - Body: { access_token, user: { id, email, full_name, role } }
  - Cookie: refresh_token (httpOnly, Secure, SameSite=Strict)
```

### 4.2 Refresh Token
```
POST /api/auth/refresh
Cookie: refresh_token

→ Parse JWT refresh token
→ Hash token → query refresh_tokens (check not revoked, not expired)
→ Revoke token cũ
→ Issue Access Token mới + Refresh Token mới
```

### 4.3 Middleware Auth + RBAC
```
Mỗi request protected:
1. Extract Authorization: Bearer <token>
2. Verify JWT signature + exp
3. Inject { user_id, role } vào Gin context
4. RBAC middleware kiểm tra role có quyền với route không
   → 403 nếu không đủ quyền
```

### 4.4 RBAC Route Mapping
```
admin          → tất cả routes
director       → GET tất cả, không POST/PUT/DELETE trên admin routes
sales_manager  → full /crm /sales, GET /tracking
sales_rep      → /crm (own), /sales (own), GET /tracking
marketing      → full /marketing, GET /crm
accountant     → full /accounting, GET /crm, GET /tracking
```

---

## 5. API Endpoints

> Tất cả request/response dùng JSON. Prefix: `/api/v1`

### 5.1 Auth
```
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
GET    /api/v1/auth/me
PUT    /api/v1/auth/me/password
```

### 5.2 Users (Admin only)
```
GET    /api/v1/users                  ?page&limit&role
POST   /api/v1/users
GET    /api/v1/users/:id
PUT    /api/v1/users/:id
DELETE /api/v1/users/:id
```

### 5.3 CRM – Customers
```
GET    /api/v1/customers              ?page&limit&segment&tier&assigned_to&q
POST   /api/v1/customers
GET    /api/v1/customers/:id
PUT    /api/v1/customers/:id
DELETE /api/v1/customers/:id

GET    /api/v1/customers/:id/contacts
POST   /api/v1/customers/:id/contacts
PUT    /api/v1/customers/:id/contacts/:contact_id
DELETE /api/v1/customers/:id/contacts/:contact_id

GET    /api/v1/customers/:id/interactions
POST   /api/v1/customers/:id/interactions

GET    /api/v1/customers/:id/shipments
GET    /api/v1/customers/:id/invoices
GET    /api/v1/customers/:id/contracts
```

### 5.4 Sales
```
GET    /api/v1/opportunities          ?page&limit&stage&assigned_to
POST   /api/v1/opportunities
GET    /api/v1/opportunities/:id
PUT    /api/v1/opportunities/:id
DELETE /api/v1/opportunities/:id

GET    /api/v1/contracts              ?page&limit&status&customer_id
POST   /api/v1/contracts
GET    /api/v1/contracts/:id
PUT    /api/v1/contracts/:id
POST   /api/v1/contracts/:id/upload   -- upload file PDF hợp đồng

GET    /api/v1/quotations             ?page&limit&status&customer_id
POST   /api/v1/quotations
GET    /api/v1/quotations/:id
PUT    /api/v1/quotations/:id
POST   /api/v1/quotations/:id/send    -- gửi email báo giá
POST   /api/v1/quotations/:id/convert -- chuyển thành contract
```

### 5.5 Tracking
```
GET    /api/v1/shipments              ?page&limit&status&customer_id
GET    /api/v1/shipments/:id
GET    /api/v1/shipments/:id/events
POST   /api/v1/shipments/:id/sync     -- manual trigger sync từ 3rd party API
POST   /api/v1/webhooks/tracking      -- nhận webhook từ Tracking provider (no auth, verify HMAC)
```

### 5.6 Accounting
```
GET    /api/v1/invoices               ?page&limit&status&customer_id&from&to
POST   /api/v1/invoices
GET    /api/v1/invoices/:id
PUT    /api/v1/invoices/:id
POST   /api/v1/invoices/:id/send      -- gửi invoice PDF qua email
POST   /api/v1/invoices/:id/cancel
GET    /api/v1/invoices/:id/download  -- trả về PDF binary

GET    /api/v1/payments
POST   /api/v1/payments               -- ghi nhận thanh toán thủ công
POST   /api/v1/webhooks/sepay         -- nhận webhook SePay (no auth, verify secret header)

GET    /api/v1/reports/revenue        ?from&to&group_by=month|customer
GET    /api/v1/reports/ar             -- báo cáo công nợ hiện tại
```

### 5.7 Marketing
```
GET    /api/v1/campaigns              ?page&limit&status
POST   /api/v1/campaigns
GET    /api/v1/campaigns/:id
PUT    /api/v1/campaigns/:id
POST   /api/v1/campaigns/:id/send     -- gửi ngay
POST   /api/v1/campaigns/:id/schedule -- lên lịch gửi
GET    /api/v1/campaigns/:id/logs     ?page&limit&status

GET    /api/v1/sendgrid/webhook       -- SendGrid event webhook (delivered, open, click, bounce)
```

### 5.8 Dashboard
```
GET    /api/v1/dashboard/summary      -- KPI tổng: doanh thu, đơn hàng, KH mới, công nợ
GET    /api/v1/dashboard/sales-funnel -- Pipeline theo stage
GET    /api/v1/dashboard/shipment-map -- Danh sách shipment có lat/lng đang active
```

---

## 6. Integration Specs

### 6.1 Tracking API (3rd party)

```
Polling job chạy mỗi TRACKING_POLL_INTERVAL_SEC giây (default: 300)
→ Lấy danh sách shipments có status != delivered|failed
→ Gọi GET {TRACKING_API_BASE_URL}/shipments?ids=...
  Header: X-Api-Key: {TRACKING_API_KEY}
→ Update bảng shipments + insert shipment_events nếu có status mới

Webhook nhận (nếu provider hỗ trợ):
POST /api/v1/webhooks/tracking
Header: X-Hmac-Signature: <sha256 của body + secret>
Body: { tracking_code, status, location, lat, lng, event_time, description }
→ Verify HMAC → upsert shipment + insert event
```

### 6.2 SePay Webhook

```
POST /api/v1/webhooks/sepay
Header: Authorization: Apikey {SEPAY_WEBHOOK_SECRET}

Body (SePay format):
{
  "id": "txn_123",
  "gateway": "VietcomBank",
  "transactionDate": "2026-06-15 10:00:00",
  "accountNumber": "1234567890",
  "code": "INV-2026-001",      -- nội dung chuyển khoản, chứa mã invoice
  "content": "Thanh toan INV-2026-001",
  "transferType": "in",
  "transferAmount": 5000000,
  "accumulated": 5000000,
  "referenceCode": "FT26166...",
  "description": ""
}

Xử lý:
1. Verify header Authorization
2. Parse code → tìm invoice theo code
3. Kiểm tra transferAmount >= invoice.total
4. Nếu match → tạo payment record, cập nhật invoice.status = paid, invoice.paid_at
5. Gửi email xác nhận thanh toán (SendGrid)
6. Nếu không match → log vào bảng riêng để review thủ công
```

### 6.3 SendGrid

```
Dùng SendGrid Dynamic Templates.
Các loại email:
- invoice_sent:       gửi kèm PDF attachment
- payment_confirmed:  xác nhận đã nhận tiền
- quotation_sent:     gửi báo giá
- campaign_blast:     email marketing bulk
- shipment_update:    cập nhật trạng thái đơn hàng

Webhook SendGrid → POST /api/v1/sendgrid/webhook
Events: delivered | open | click | bounce | spam_report
→ Update campaign_logs.status
```

### 6.4 Leaflet Map (Frontend)

```
Backend chỉ cần cung cấp endpoint:
GET /api/v1/dashboard/shipment-map
Response: [{ tracking_code, status, lat, lng, customer_name, destination }]

Frontend dùng Leaflet.js + tile OpenStreetMap (không cần API key):
- Render markers theo lat/lng
- Popup hiện tracking_code, status, customer
- Không call Google Maps API
```

---

## 7. Nginx Config (tham khảo)

```nginx
server {
    listen 443 ssl;
    server_name erp.company.com;

    ssl_certificate     /etc/letsencrypt/live/erp.company.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/erp.company.com/privkey.pem;

    # Static frontend (React build)
    location / {
        root /var/www/logistics-erp/dist;
        try_files $uri $uri/ /index.html;
    }

    # API reverse proxy
    location /api/ {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_read_timeout 60s;
    }

    # File downloads (invoices, contracts)
    location /files/ {
        alias /var/www/logistics-erp/uploads/;
        internal;
    }
}
```

---

## 8. systemd Service

```ini
# /etc/systemd/system/logistics-erp.service
[Unit]
Description=Logistics ERP API
After=network.target

[Service]
Type=simple
User=erp
WorkingDirectory=/srv/logistics-erp
EnvironmentFile=/srv/logistics-erp/.env
ExecStart=/srv/logistics-erp/server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable logistics-erp
systemctl start logistics-erp
```

---

## 9. GitHub Actions CI/CD

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        run: GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

      - name: Test
        run: go test ./...

      - name: Deploy to VPS
        uses: appleboy/scp-action@v0.1.7
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          source: server
          target: /srv/logistics-erp/

      - name: Restart service
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          script: systemctl restart logistics-erp
```
