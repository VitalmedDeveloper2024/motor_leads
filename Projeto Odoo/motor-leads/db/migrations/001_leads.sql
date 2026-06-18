CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS leads (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    source          TEXT NOT NULL,
    campaign        TEXT,
    nome            TEXT,
    telefone        TEXT,
    email           TEXT,
    source_payload  JSONB NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    attempts        INT  NOT NULL DEFAULT 0,
    error_log       TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status);
CREATE INDEX IF NOT EXISTS idx_leads_tenant ON leads(tenant_id);
CREATE INDEX IF NOT EXISTS idx_leads_source ON leads(source);
