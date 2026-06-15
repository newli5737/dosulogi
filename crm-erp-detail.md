# Tài liệu Kỹ thuật Chi tiết – Module CRM & Sales ERP

**Stack:** Go (Gin) · PostgreSQL 16 · Redis  
**Phạm vi:** CRM (Customer, Contact, Interaction, Ticket) + Sales ERP (Opportunity, Quotation, Contract)

---

## 1. Database Schema

### 1.1 customers
```sql
CREATE TABLE customers (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code             VARCHAR(20) UNIQUE NOT NULL,       -- tự sinh: KH-00001
    name             VARCHAR(255) NOT NULL,
    type             VARCHAR(10)  NOT NULL               CHECK (type IN ('B2B','B2C')),
    email            VARCHAR(255),
    phone            VARCHAR(20),
    address          TEXT,
    province         VARCHAR(100),
    tax_code         VARCHAR(20),                        -- mã số thuế (B2B)
    segment          VARCHAR(20)  NOT NULL DEFAULT 'standard'
                                  CHECK (segment IN ('enterprise','sme','individual')),
    tier             VARCHAR(20)  NOT NULL DEFAULT 'standard'
                                  CHECK (tier IN ('gold','silver','standard')),
    assigned_to      UUID         REFERENCES users(id) ON DELETE SET NULL,
    last_contact_at  TIMESTAMPTZ,
    is_active        BOOLEAN      NOT NULL DEFAULT true,
    created_by       UUID         REFERENCES users(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_customers_assigned_to  ON customers(assigned_to);
CREATE INDEX idx_customers_segment_tier ON customers(segment, tier);
CREATE INDEX idx_customers_name_trgm    ON customers USING gin (name gin_trgm_ops);
-- Cần: CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

### 1.2 contacts
```sql
CREATE TABLE contacts (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id  UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    role         VARCHAR(100),                           -- Giám đốc, Kế toán, Logistics Manager...
    phone        VARCHAR(20),
    email        VARCHAR(255),
    is_primary   BOOLEAN     NOT NULL DEFAULT false,
    note         TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Mỗi customer chỉ có 1 primary contact
CREATE UNIQUE INDEX idx_contacts_primary
    ON contacts(customer_id)
    WHERE is_primary = true;

CREATE INDEX idx_contacts_customer ON contacts(customer_id);
```

### 1.3 interactions
```sql
CREATE TABLE interactions (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id  UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    channel      VARCHAR(20) NOT NULL CHECK (channel IN ('call','email','meeting','note','other')),
    direction    VARCHAR(10)          CHECK (direction IN ('inbound','outbound')),
    summary      TEXT        NOT NULL,
    occurred_at  TIMESTAMPTZ NOT NULL DEFAULT now(),    -- thời điểm thực tế xảy ra
    created_by   UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_interactions_customer ON interactions(customer_id);
CREATE INDEX idx_interactions_occurred ON interactions(occurred_at DESC);
```

### 1.4 tickets
```sql
CREATE TABLE tickets (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code         VARCHAR(20) UNIQUE NOT NULL,            -- TK-00001
    customer_id  UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    title        VARCHAR(500) NOT NULL,
    description  TEXT,
    priority     VARCHAR(10) NOT NULL DEFAULT 'medium'
                              CHECK (priority IN ('low','medium','high','urgent')),
    status       VARCHAR(20) NOT NULL DEFAULT 'open'
                              CHECK (status IN ('open','in_progress','pending_customer','resolved','closed')),
    category     VARCHAR(50),                            -- billing | shipment | complaint | other
    assigned_to  UUID        REFERENCES users(id) ON DELETE SET NULL,
    sla_deadline TIMESTAMPTZ,                            -- tính từ created_at theo priority
    resolved_at  TIMESTAMPTZ,
    closed_at    TIMESTAMPTZ,
    created_by   UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- SLA deadline tính theo priority (business logic ở service layer):
-- urgent: +4h, high: +8h, medium: +24h, low: +72h

CREATE INDEX idx_tickets_customer    ON tickets(customer_id);
CREATE INDEX idx_tickets_assigned_to ON tickets(assigned_to);
CREATE INDEX idx_tickets_status      ON tickets(status);
CREATE INDEX idx_tickets_sla         ON tickets(sla_deadline) WHERE status NOT IN ('resolved','closed');

CREATE TABLE ticket_comments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id  UUID        NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    body       TEXT        NOT NULL,
    is_internal BOOLEAN    NOT NULL DEFAULT false,       -- true = note nội bộ, không gửi KH
    created_by UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 1.5 opportunities
```sql
CREATE TABLE opportunities (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    code          VARCHAR(20)  UNIQUE NOT NULL,          -- OPP-00001
    customer_id   UUID         NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    title         VARCHAR(500) NOT NULL,
    stage         VARCHAR(20)  NOT NULL DEFAULT 'lead'
                               CHECK (stage IN ('lead','qualified','proposal','negotiation','won','lost')),
    value         NUMERIC(15,2),
    currency      VARCHAR(3)   NOT NULL DEFAULT 'VND',
    expected_close DATE,
    assigned_to   UUID         REFERENCES users(id) ON DELETE SET NULL,
    lost_reason   TEXT,
    note          TEXT,
    created_by    UUID         REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_opp_customer    ON opportunities(customer_id);
CREATE INDEX idx_opp_stage       ON opportunities(stage);
CREATE INDEX idx_opp_assigned_to ON opportunities(assigned_to);

-- Gắn opportunity với shipment cụ thể (many-to-many)
CREATE TABLE opportunity_shipments (
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    shipment_id    UUID NOT NULL REFERENCES shipments(id)    ON DELETE CASCADE,
    PRIMARY KEY (opportunity_id, shipment_id)
);
```

### 1.6 quotations
```sql
CREATE TABLE quotations (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(20) UNIQUE NOT NULL,         -- BG-2026-00001
    customer_id     UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    opportunity_id  UUID        REFERENCES opportunities(id) ON DELETE SET NULL,
    items           JSONB       NOT NULL DEFAULT '[]',
    -- items format:
    -- [{ "description": "Cước FCL 20ft HCM-HN", "qty": 1, "unit": "chuyến",
    --    "unit_price": 15000000, "amount": 15000000 }]
    subtotal        NUMERIC(15,2) NOT NULL DEFAULT 0,
    discount        NUMERIC(15,2) NOT NULL DEFAULT 0,
    tax_rate        NUMERIC(5,2)  NOT NULL DEFAULT 10,   -- %
    tax_amount      NUMERIC(15,2) NOT NULL DEFAULT 0,
    total           NUMERIC(15,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(3)   NOT NULL DEFAULT 'VND',
    valid_until     DATE,
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft'
                                 CHECK (status IN ('draft','sent','accepted','rejected','expired')),
    note            TEXT,
    sent_at         TIMESTAMPTZ,
    file_url        VARCHAR(500),                        -- PDF đã generate
    created_by      UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_quotations_customer ON quotations(customer_id);
CREATE INDEX idx_quotations_opp      ON quotations(opportunity_id);
```

### 1.7 contracts
```sql
CREATE TABLE contracts (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(20) UNIQUE NOT NULL,         -- HD-2026-00001
    customer_id     UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    opportunity_id  UUID        REFERENCES opportunities(id) ON DELETE SET NULL,
    quotation_id    UUID        REFERENCES quotations(id) ON DELETE SET NULL,
    title           VARCHAR(500),
    service_type    VARCHAR(20) NOT NULL
                                CHECK (service_type IN ('FCL','LCL','air','express','road')),
    start_date      DATE        NOT NULL,
    end_date        DATE,
    value           NUMERIC(15,2),
    currency        VARCHAR(3)  NOT NULL DEFAULT 'VND',
    payment_terms   VARCHAR(100),                        -- "30 ngày kể từ ngày xuất hóa đơn"
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
                                CHECK (status IN ('draft','active','expired','terminated')),
    file_url        VARCHAR(500),                        -- file hợp đồng PDF scan/upload
    note            TEXT,
    signed_at       DATE,
    created_by      UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Một customer có thể có nhiều contract active cùng lúc (FCL + LCL song song)
-- Không có unique constraint trên (customer_id, status)

CREATE INDEX idx_contracts_customer ON contracts(customer_id);
CREATE INDEX idx_contracts_status   ON contracts(status);

-- Contract gắn với nhiều shipment
CREATE TABLE contract_shipments (
    contract_id UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    PRIMARY KEY (contract_id, shipment_id)
);
```

### 1.8 stage_history (audit trail pipeline)
```sql
CREATE TABLE opportunity_stage_history (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    opportunity_id UUID        NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    from_stage     VARCHAR(20),
    to_stage       VARCHAR(20) NOT NULL,
    note           TEXT,
    changed_by     UUID        REFERENCES users(id) ON DELETE SET NULL,
    changed_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

## 2. API Endpoints – CRM

> Prefix: `/api/v1`  
> Auth: `Authorization: Bearer <access_token>` (tất cả routes bên dưới)

---

### 2.1 Customers

#### GET /customers
```
Query params:
  page        int     default 1
  limit       int     default 20, max 100
  q           string  full-text search trên name, email, phone, code
  type        string  B2B | B2C
  segment     string  enterprise | sme | individual
  tier        string  gold | silver | standard
  assigned_to uuid    lọc theo sales rep
  province    string
  is_active   bool    default true

Response 200:
{
  "data": [
    {
      "id": "uuid",
      "code": "KH-00001",
      "name": "Công ty TNHH ABC",
      "type": "B2B",
      "email": "contact@abc.vn",
      "phone": "0901234567",
      "province": "Hồ Chí Minh",
      "segment": "sme",
      "tier": "gold",
      "assigned_to": { "id": "uuid", "full_name": "Nguyễn Văn A" },
      "last_contact_at": "2026-06-10T08:00:00Z",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 350 }
}
```

#### POST /customers
```
Body:
{
  "name":       "Công ty TNHH ABC",       -- required
  "type":       "B2B",                    -- required
  "email":      "contact@abc.vn",
  "phone":      "0901234567",
  "address":    "123 Nguyễn Huệ, Q1",
  "province":   "Hồ Chí Minh",
  "tax_code":   "0312345678",
  "segment":    "sme",
  "tier":       "standard",
  "assigned_to": "uuid"
}

Response 201:
{ "data": { <customer object đầy đủ> } }

Errors:
  400 nếu name hoặc type thiếu
  409 nếu email đã tồn tại
```

#### GET /customers/:id
```
Response 200:
{
  "data": {
    ...customer fields...,
    "primary_contact": { id, name, role, phone, email },
    "open_tickets":    2,
    "active_contracts": 1
  }
}
  404 nếu không tìm thấy
```

#### PUT /customers/:id
```
Body: bất kỳ field nào cần cập nhật (partial update)
Response 200: { "data": { <customer object> } }
```

#### DELETE /customers/:id
```
Soft delete: set is_active = false
Response 204
403 nếu role không phải admin hoặc sales_manager
```

---

### 2.2 Contacts

#### GET /customers/:id/contacts
```
Response 200:
{
  "data": [
    { "id", "name", "role", "phone", "email", "is_primary" }
  ]
}
```

#### POST /customers/:id/contacts
```
Body:
{
  "name":       "Trần Thị B",   -- required
  "role":       "Kế toán",
  "phone":      "0912345678",
  "email":      "b@abc.vn",
  "is_primary": false
}

Logic: nếu is_primary = true → set is_primary = false cho tất cả contact cũ của customer đó
Response 201: { "data": { <contact> } }
```

#### PUT /customers/:id/contacts/:contact_id
```
Body: partial update
Logic: nếu is_primary = true → unset primary contact cũ
Response 200
```

#### DELETE /customers/:id/contacts/:contact_id
```
Hard delete
403 nếu đây là primary contact duy nhất
Response 204
```

---

### 2.3 Interactions

#### GET /customers/:id/interactions
```
Query params:
  page    int
  limit   int   default 20
  channel string  call | email | meeting | note
  from    date
  to      date

Response 200:
{
  "data": [
    {
      "id", "channel", "direction", "summary",
      "occurred_at",
      "created_by": { "id", "full_name" }
    }
  ],
  "meta": { "page", "limit", "total" }
}
```

#### POST /customers/:id/interactions
```
Body:
{
  "channel":     "call",        -- required
  "direction":   "outbound",
  "summary":     "Trao đổi về lịch vận chuyển tháng 7", -- required
  "occurred_at": "2026-06-15T10:00:00Z"   -- default now() nếu không truyền
}

Side effect: cập nhật customers.last_contact_at = occurred_at
Response 201
```

---

### 2.4 Tickets

#### GET /tickets
```
Query params:
  page        int
  limit       int     default 20
  status      string  open | in_progress | pending_customer | resolved | closed
  priority    string  low | medium | high | urgent
  assigned_to uuid
  customer_id uuid
  overdue     bool    lọc ticket đã quá sla_deadline mà chưa resolved
  from        date
  to          date

Response 200:
{
  "data": [
    {
      "id", "code", "title", "priority", "status", "category",
      "customer": { "id", "name", "code" },
      "assigned_to": { "id", "full_name" },
      "sla_deadline": "2026-06-16T10:00:00Z",
      "is_overdue": false,
      "created_at"
    }
  ],
  "meta": { "page", "limit", "total" }
}
```

#### POST /tickets
```
Body:
{
  "customer_id":  "uuid",         -- required
  "title":        "Đơn hàng bị delay không có thông báo", -- required
  "description":  "...",
  "priority":     "high",         -- default medium
  "category":     "shipment"
}

Logic sau khi insert:
  - Tự động tính sla_deadline từ priority:
      urgent → now() + 4h
      high   → now() + 8h
      medium → now() + 24h
      low    → now() + 72h
  - Tự động assign cho sales rep của customer (assigned_to = customer.assigned_to)
  - Tạo auto comment: "Ticket được tạo tự động từ hệ thống"

Response 201: { "data": { <ticket đầy đủ> } }
```

#### GET /tickets/:id
```
Response 200:
{
  "data": {
    ...ticket fields...,
    "customer": { <customer summary> },
    "comments": [
      { "id", "body", "is_internal", "created_by": { "id", "full_name" }, "created_at" }
    ]
  }
}
```

#### PUT /tickets/:id
```
Body (partial):
{
  "status":      "in_progress",
  "assigned_to": "uuid",
  "priority":    "urgent"
}

Logic:
  - Nếu status → resolved: set resolved_at = now()
  - Nếu status → closed:   set closed_at = now()
  - Nếu priority thay đổi: recalculate sla_deadline

Response 200
```

#### POST /tickets/:id/comments
```
Body:
{
  "body":        "Đã liên hệ nhà xe, dự kiến giao ngày mai",
  "is_internal": false
}

Response 201: { "data": { <comment> } }
```

---

## 3. API Endpoints – Sales ERP

### 3.1 Opportunities

#### GET /opportunities
```
Query params:
  page          int
  limit         int     default 20
  stage         string  lead | qualified | proposal | negotiation | won | lost
  assigned_to   uuid
  customer_id   uuid
  from          date    lọc theo expected_close
  to            date

Response 200:
{
  "data": [
    {
      "id", "code", "title", "stage", "value", "currency",
      "expected_close",
      "customer":    { "id", "name", "code" },
      "assigned_to": { "id", "full_name" },
      "shipments":   [ { "id", "tracking_code", "status" } ],
      "created_at",  "updated_at"
    }
  ],
  "meta": { "page", "limit", "total" }
}
```

#### POST /opportunities
```
Body:
{
  "customer_id":    "uuid",               -- required
  "title":          "Hợp đồng FCL Q3/2026", -- required
  "stage":          "lead",               -- default lead
  "value":          500000000,
  "currency":       "VND",
  "expected_close": "2026-07-31",
  "assigned_to":    "uuid",
  "shipment_ids":   ["uuid1", "uuid2"],   -- gắn với shipment cụ thể
  "note":           ""
}

Response 201: { "data": { <opportunity> } }
```

#### PUT /opportunities/:id
```
Body (partial):
{
  "stage":          "proposal",
  "value":          600000000,
  "shipment_ids":   ["uuid1", "uuid2", "uuid3"],  -- replace toàn bộ danh sách
  "lost_reason":    ""   -- bắt buộc nếu stage = lost
}

Logic:
  - Khi stage thay đổi → insert opportunity_stage_history
  - Nếu stage = won → gợi ý tạo contract (không tự động tạo)

Response 200
```

#### GET /opportunities/:id/stage-history
```
Response 200:
{
  "data": [
    { "from_stage", "to_stage", "note", "changed_by": { "full_name" }, "changed_at" }
  ]
}
```

---

### 3.2 Quotations

#### GET /quotations
```
Query params:
  page, limit, status, customer_id, opportunity_id, from, to

Response 200: list quotation summary
```

#### POST /quotations
```
Body:
{
  "customer_id":    "uuid",         -- required
  "opportunity_id": "uuid",
  "valid_until":    "2026-07-15",
  "items": [
    {
      "description": "Cước FCL 20ft tuyến HCM-HN",
      "qty":         2,
      "unit":        "chuyến",
      "unit_price":  15000000
    }
  ],
  "discount":  0,
  "tax_rate":  10,
  "note":      ""
}

Logic:
  - Tự tính amount = qty * unit_price cho từng item
  - subtotal = sum(amount)
  - tax_amount = (subtotal - discount) * tax_rate / 100
  - total = subtotal - discount + tax_amount
  - Tự sinh code: BG-2026-00001

Response 201: { "data": { <quotation đầy đủ, bao gồm items đã tính amount> } }
```

#### POST /quotations/:id/send
```
Body: không cần (hoặc override email nếu muốn)

Logic:
  1. Generate PDF từ quotation data
  2. Lưu file vào /uploads/quotations/{code}.pdf → set file_url
  3. Gửi email qua SendGrid kèm PDF attachment đến customer.email
  4. Set status = sent, sent_at = now()

Response 200: { "data": { "file_url": "...", "sent_at": "..." } }
```

#### POST /quotations/:id/convert
```
Chuyển quotation → contract

Body:
{
  "start_date":    "2026-07-01",   -- required
  "end_date":      "2027-06-30",
  "service_type":  "FCL",          -- required
  "payment_terms": "30 ngày kể từ ngày xuất hóa đơn"
}

Logic:
  1. Tạo contract mới lấy thông tin từ quotation
  2. Set quotation.status = accepted
  3. Nếu có opportunity_id → set opportunity.stage = won + log stage history

Response 201: { "data": { <contract mới> } }
```

---

### 3.3 Contracts

#### GET /contracts
```
Query params:
  page, limit, status, customer_id, service_type, from (start_date), to

Response 200: list contract summary
```

#### POST /contracts
```
Body:
{
  "customer_id":    "uuid",         -- required
  "opportunity_id": "uuid",
  "quotation_id":   "uuid",
  "title":          "HĐ vận chuyển FCL 2026",
  "service_type":   "FCL",          -- required
  "start_date":     "2026-07-01",   -- required
  "end_date":       "2027-06-30",
  "value":          120000000,
  "payment_terms":  "30 ngày kể từ ngày xuất hóa đơn",
  "note":           ""
}

Response 201
```

#### PUT /contracts/:id
```
Body (partial): bất kỳ field nào
Logic: nếu status → expired hoặc terminated → không cho sửa nữa (trả 409)
Response 200
```

#### POST /contracts/:id/upload
```
Content-Type: multipart/form-data
Field: file (PDF, max 10MB)

Logic: lưu file → /uploads/contracts/{code}.pdf → cập nhật file_url
Response 200: { "data": { "file_url": "..." } }
```

#### GET /contracts/:id/shipments
```
Trả về danh sách shipment gắn với contract

Response 200:
{
  "data": [
    { "id", "tracking_code", "status", "origin", "destination",
      "lat", "lng", "estimated_delivery", "last_synced_at" }
  ]
}
```

#### POST /contracts/:id/shipments
```
Gắn shipment vào contract

Body: { "shipment_ids": ["uuid1", "uuid2"] }
Logic: insert contract_shipments (ignore nếu đã tồn tại)
Response 200
```

---

## 4. Business Logic quan trọng

### 4.1 Tự sinh code
```
Tất cả code (KH-, TK-, OPP-, BG-, HD-) sinh theo pattern:
  PREFIX-YYYY-NNNNN (5 chữ số, tự tăng theo năm)

Ví dụ:
  KH-00001          (customer, không có năm)
  OPP-00001
  BG-2026-00001
  HD-2026-00001
  TK-00001

Dùng PostgreSQL sequence hoặc SELECT MAX + 1 trong transaction.
Đặt trong utility function: GenerateCode(prefix, withYear bool) string
```

### 4.2 SLA Ticket
```
Khi tạo ticket:
  sla_deadline = created_at + duration theo priority
  urgent  → +4h
  high    → +8h
  medium  → +24h
  low     → +72h

Cron job chạy mỗi 5 phút:
  SELECT tickets WHERE sla_deadline < now() AND status NOT IN ('resolved','closed')
  → Gửi email cảnh báo cho assigned_to và sales_manager
  → (không tự đóng, chỉ cảnh báo)
```

### 4.3 Cập nhật last_contact_at
```
Mỗi khi INSERT interaction:
  UPDATE customers SET last_contact_at = occurred_at
  WHERE id = customer_id AND (last_contact_at IS NULL OR last_contact_at < occurred_at)
```

### 4.4 Tính toán Quotation
```
Thực hiện ở service layer, KHÔNG để frontend tính:
  item.amount  = item.qty * item.unit_price
  subtotal     = SUM(item.amount)
  tax_amount   = ROUND((subtotal - discount) * tax_rate / 100, 0)
  total        = subtotal - discount + tax_amount
```

### 4.5 Chuyển Opportunity stage
```
Luôn INSERT opportunity_stage_history khi stage thay đổi.
Stage cho phép chuyển:
  lead → qualified
  qualified → proposal | lost
  proposal → negotiation | lost
  negotiation → won | lost
  won / lost → không thể chuyển ngược (trả 400 nếu cố tình)
```

---

## 5. Error Response Format

```json
{
  "error": {
    "code":    "CUSTOMER_NOT_FOUND",
    "message": "Khách hàng không tồn tại"
  }
}
```

| HTTP | Code | Trường hợp |
|---|---|---|
| 400 | VALIDATION_ERROR | Thiếu field bắt buộc, sai format |
| 400 | INVALID_STAGE_TRANSITION | Chuyển stage không hợp lệ |
| 401 | UNAUTHORIZED | Token thiếu hoặc hết hạn |
| 403 | FORBIDDEN | Không đủ quyền |
| 404 | NOT_FOUND | Resource không tồn tại |
| 409 | CONFLICT | Trùng email/code, contract đã expired |
| 500 | INTERNAL_ERROR | Lỗi server |

---

## 6. Redis Usage

```
Key schema:

rate_limit:{user_id}:{route}     → TTL 60s, counter
session:refresh:{token_hash}     → TTL theo refresh token, value = user_id
sla_notified:{ticket_id}         → TTL 1h, tránh gửi email SLA trùng lặp
```

