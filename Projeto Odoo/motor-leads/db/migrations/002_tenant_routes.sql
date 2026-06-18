-- De-para: (source, campaign) -> tenant_id (empresa Odoo / conta Chatwoot).
-- campaign NULL = fallback por source.
CREATE TABLE IF NOT EXISTS tenant_routes (
    id         BIGSERIAL PRIMARY KEY,
    source     TEXT NOT NULL,
    campaign   TEXT,
    tenant_id  TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (source, campaign)
);

-- Exemplos:
-- INSERT INTO tenant_routes (source, campaign, tenant_id) VALUES ('meta', 'black-friday', 'empresa_a');
-- INSERT INTO tenant_routes (source, campaign, tenant_id) VALUES ('site', NULL, 'empresa_b');
