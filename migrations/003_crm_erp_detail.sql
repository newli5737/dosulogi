-- CRM & Sales ERP detail schema upgrade

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- customers upgrades
ALTER TABLE customers ADD COLUMN IF NOT EXISTS tax_code VARCHAR(20);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE customers ALTER COLUMN segment SET DEFAULT 'standard';
ALTER TABLE customers ALTER COLUMN tier SET DEFAULT 'standard';

-- contacts upgrades
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS note TEXT;
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
DROP INDEX IF EXISTS idx_contacts_primary;
CREATE UNIQUE INDEX IF NOT EXISTS idx_contacts_primary ON contacts(customer_id) WHERE is_primary = true;

-- interactions upgrades
ALTER TABLE interactions ADD COLUMN IF NOT EXISTS occurred_at TIMESTAMPTZ NOT NULL DEFAULT now();
UPDATE interactions SET occurred_at = created_at WHERE occurred_at IS NULL;

-- tickets
CREATE TABLE IF NOT EXISTS tickets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code         VARCHAR(20) UNIQUE NOT NULL,
    customer_id  UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    title        VARCHAR(500) NOT NULL,
    description  TEXT,
    priority     VARCHAR(10) NOT NULL DEFAULT 'medium',
    status       VARCHAR(20) NOT NULL DEFAULT 'open',
    category     VARCHAR(50),
    assigned_to  UUID REFERENCES users(id) ON DELETE SET NULL,
    sla_deadline TIMESTAMPTZ,
    resolved_at  TIMESTAMPTZ,
    closed_at    TIMESTAMPTZ,
    created_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ticket_comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id   UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    body        TEXT NOT NULL,
    is_internal BOOLEAN NOT NULL DEFAULT false,
    created_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tickets_customer ON tickets(customer_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_assigned ON tickets(assigned_to);

-- opportunities upgrades
ALTER TABLE opportunities ADD COLUMN IF NOT EXISTS code VARCHAR(20);
ALTER TABLE opportunities ADD COLUMN IF NOT EXISTS note TEXT;

CREATE TABLE IF NOT EXISTS opportunity_stage_history (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    from_stage     VARCHAR(20),
    to_stage       VARCHAR(20) NOT NULL,
    note           TEXT,
    changed_by     UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS opportunity_shipments (
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    shipment_id    UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    PRIMARY KEY (opportunity_id, shipment_id)
);

-- quotations upgrades
ALTER TABLE quotations ADD COLUMN IF NOT EXISTS opportunity_id UUID REFERENCES opportunities(id) ON DELETE SET NULL;
ALTER TABLE quotations ADD COLUMN IF NOT EXISTS subtotal NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE quotations ADD COLUMN IF NOT EXISTS discount NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE quotations ADD COLUMN IF NOT EXISTS tax_rate NUMERIC(5,2) NOT NULL DEFAULT 10;
ALTER TABLE quotations ADD COLUMN IF NOT EXISTS tax_amount NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE quotations ADD COLUMN IF NOT EXISTS note TEXT;
ALTER TABLE quotations ADD COLUMN IF NOT EXISTS sent_at TIMESTAMPTZ;
ALTER TABLE quotations ADD COLUMN IF NOT EXISTS file_url VARCHAR(500);
ALTER TABLE quotations ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
UPDATE quotations SET opportunity_id = opp_id WHERE opportunity_id IS NULL AND opp_id IS NOT NULL;

-- contracts upgrades
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS quotation_id UUID REFERENCES quotations(id) ON DELETE SET NULL;
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS payment_terms VARCHAR(100);
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS note TEXT;
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS signed_at DATE;

CREATE TABLE IF NOT EXISTS contract_shipments (
    contract_id UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    PRIMARY KEY (contract_id, shipment_id)
);

CREATE INDEX IF NOT EXISTS idx_customers_name_trgm ON customers USING gin (name gin_trgm_ops);
