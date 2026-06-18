-- Credenciais Odoo/Chatwoot por tenant. (Colunas a confirmar com o negócio.)
CREATE TABLE IF NOT EXISTS tenant_credentials (
    tenant_id            TEXT PRIMARY KEY,
    odoo_base_url        TEXT,
    odoo_db              TEXT,
    odoo_uid             INT,
    odoo_api_key         TEXT,
    odoo_company_id      INT,
    chatwoot_base_url    TEXT,
    chatwoot_account_id  INT,
    chatwoot_api_token   TEXT,
    chatwoot_inbox_id    INT,
    updated_at           TIMESTAMPTZ DEFAULT NOW()
);
