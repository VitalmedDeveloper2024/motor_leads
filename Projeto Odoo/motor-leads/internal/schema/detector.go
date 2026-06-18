package schema

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/seu-org/motor-leads/internal/metrics"
)

// Detect compara as chaves de primeiro nível do payload com o schema
// esperado (tabela source_schemas) e registra mudanças.
func Detect(ctx context.Context, db *sqlx.DB, source string, payload []byte) {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return
	}
	fields := keysOf(m)

	stored, err := loadSchema(ctx, db, source)
	if err != nil {
		return
	}
	if len(stored) == 0 { // primeira vez: grava baseline
		_ = saveSchema(ctx, db, source, fields)
		return
	}

	added, removed := diff(stored, fields)
	if len(added) > 0 || len(removed) > 0 {
		_ = logSchemaChange(ctx, db, source, added, removed)
		metrics.SchemaChanges.WithLabelValues(source).Inc()
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func diff(old, cur []string) (added, removed []string) {
	o := toSet(old)
	c := toSet(cur)
	for k := range c {
		if _, ok := o[k]; !ok {
			added = append(added, k)
		}
	}
	for k := range o {
		if _, ok := c[k]; !ok {
			removed = append(removed, k)
		}
	}
	return
}

func toSet(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	for _, k := range s {
		m[k] = struct{}{}
	}
	return m
}

func loadSchema(ctx context.Context, db *sqlx.DB, source string) ([]string, error) {
	var fields pq.StringArray
	err := db.GetContext(ctx, &fields,
		`SELECT fields FROM source_schemas WHERE source=$1 ORDER BY updated_at DESC LIMIT 1`, source)
	if err != nil {
		return nil, nil // sem baseline ainda
	}
	return []string(fields), nil
}

func saveSchema(ctx context.Context, db *sqlx.DB, source string, fields []string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO source_schemas (source, fields) VALUES ($1, $2)`,
		source, pq.StringArray(fields))
	return err
}

func logSchemaChange(ctx context.Context, db *sqlx.DB, source string, added, removed []string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO source_schema_changes (source, added, removed) VALUES ($1, $2, $3)`,
		source, pq.StringArray(added), pq.StringArray(removed))
	return err
}
