-- Baseline de schema por origem + histórico de mudanças.
CREATE TABLE IF NOT EXISTS source_schemas (
    id         BIGSERIAL PRIMARY KEY,
    source     TEXT NOT NULL,
    fields     TEXT[] NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS source_schema_changes (
    id          BIGSERIAL PRIMARY KEY,
    source      TEXT NOT NULL,
    added       TEXT[],
    removed     TEXT[],
    detected_at TIMESTAMPTZ DEFAULT NOW()
);
