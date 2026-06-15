-- Logistics ERP/CRM initial schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 3.1 users
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    full_name   VARCHAR(255) NOT NULL,
    role        VARCHAR(50) NOT NULL,
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN DEFAULT false,
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- 3.2 CRM
CREATE TABLE IF NOT EXISTS customers (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code             VARCHAR(50) UNIQUE NOT NULL,
    name             VARCHAR(255) NOT NULL,
    type             VARCHAR(20) NOT NULL,
    email            VARCHAR(255),
    phone            VARCHAR(20),
    address          TEXT,
    province         VARCHAR(100),
    segment          VARCHAR(50),
    tier             VARCHAR(20),
    assigned_to      UUID REFERENCES users(id),
    last_contact_at  TIMESTAMPTZ,
    created_by       UUID REFERENCES users(id),
    created_at       TIMESTAMPTZ DEFAULT now(),
    updated_at       TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS contacts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id  UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    role         VARCHAR(100),
    phone        VARCHAR(20),
    email        VARCHAR(255),
    is_primary   BOOLEAN DEFAULT false,
    created_at   TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS interactions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id  UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    channel      VARCHAR(30) NOT NULL,
    direction    VARCHAR(10),
    summary      TEXT,
    created_by   UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ DEFAULT now()
);

-- 3.3 Sales
CREATE TABLE IF NOT EXISTS opportunities (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id   UUID NOT NULL REFERENCES customers(id),
    title         VARCHAR(255) NOT NULL,
    stage         VARCHAR(30) NOT NULL,
    value         NUMERIC(15,2),
    currency      VARCHAR(3) DEFAULT 'VND',
    expected_close DATE,
    assigned_to   UUID REFERENCES users(id),
    lost_reason   TEXT,
    created_by    UUID REFERENCES users(id),
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS contracts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code           VARCHAR(50) UNIQUE NOT NULL,
    customer_id    UUID NOT NULL REFERENCES customers(id),
    opportunity_id UUID REFERENCES opportunities(id),
    title          VARCHAR(255),
    start_date     DATE NOT NULL,
    end_date       DATE,
    service_type   VARCHAR(100),
    value          NUMERIC(15,2),
    currency       VARCHAR(3) DEFAULT 'VND',
    status         VARCHAR(20) DEFAULT 'draft',
    file_url       VARCHAR(500),
    created_by     UUID REFERENCES users(id),
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS quotations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code         VARCHAR(50) UNIQUE NOT NULL,
    customer_id  UUID NOT NULL REFERENCES customers(id),
    opp_id       UUID REFERENCES opportunities(id),
    items        JSONB NOT NULL,
    total        NUMERIC(15,2),
    currency     VARCHAR(3) DEFAULT 'VND',
    valid_until  DATE,
    status       VARCHAR(20) DEFAULT 'draft',
    created_by   UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ DEFAULT now()
);

-- 3.4 Tracking
CREATE TABLE IF NOT EXISTS shipments (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tracking_code      VARCHAR(100) UNIQUE NOT NULL,
    external_id        VARCHAR(255),
    customer_id        UUID REFERENCES customers(id),
    contract_id        UUID REFERENCES contracts(id),
    status             VARCHAR(50),
    origin             VARCHAR(255),
    destination        VARCHAR(255),
    lat                NUMERIC(10,7),
    lng                NUMERIC(10,7),
    estimated_delivery DATE,
    actual_delivery    TIMESTAMPTZ,
    raw_payload        JSONB,
    last_synced_at     TIMESTAMPTZ,
    created_at         TIMESTAMPTZ DEFAULT now(),
    updated_at         TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS shipment_events (
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

-- 3.5 Accounting
CREATE TABLE IF NOT EXISTS invoices (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code          VARCHAR(50) UNIQUE NOT NULL,
    customer_id   UUID NOT NULL REFERENCES customers(id),
    contract_id   UUID REFERENCES contracts(id),
    shipment_id   UUID REFERENCES shipments(id),
    items         JSONB NOT NULL,
    subtotal      NUMERIC(15,2),
    tax_rate      NUMERIC(5,2) DEFAULT 10,
    tax_amount    NUMERIC(15,2),
    total         NUMERIC(15,2),
    currency      VARCHAR(3) DEFAULT 'VND',
    status        VARCHAR(20) DEFAULT 'draft',
    due_date      DATE,
    paid_at       TIMESTAMPTZ,
    file_url      VARCHAR(500),
    created_by    UUID REFERENCES users(id),
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id      UUID NOT NULL REFERENCES invoices(id),
    amount          NUMERIC(15,2),
    method          VARCHAR(30),
    reference_code  VARCHAR(255),
    sepay_txn_id    VARCHAR(255),
    matched_auto    BOOLEAN DEFAULT false,
    note            TEXT,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS unmatched_payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sepay_txn_id    VARCHAR(255),
    amount          NUMERIC(15,2),
    reference_code  VARCHAR(255),
    raw_payload     JSONB,
    created_at      TIMESTAMPTZ DEFAULT now()
);

-- 3.6 Marketing
CREATE TABLE IF NOT EXISTS campaigns (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    type         VARCHAR(30) NOT NULL,
    status       VARCHAR(20) DEFAULT 'draft',
    subject      VARCHAR(500),
    body_html    TEXT,
    segment      JSONB,
    scheduled_at TIMESTAMPTZ,
    sent_count   INT DEFAULT 0,
    created_by   UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS campaign_logs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id   UUID NOT NULL REFERENCES campaigns(id),
    customer_id   UUID REFERENCES customers(id),
    email         VARCHAR(255),
    status        VARCHAR(20),
    sg_message_id VARCHAR(255),
    created_at    TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_customers_assigned_to ON customers(assigned_to);
CREATE INDEX IF NOT EXISTS idx_opportunities_stage ON opportunities(stage);
CREATE INDEX IF NOT EXISTS idx_shipments_status ON shipments(status);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_code ON invoices(code);
