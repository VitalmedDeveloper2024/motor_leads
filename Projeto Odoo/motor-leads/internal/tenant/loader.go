package tenant

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type mapping struct {
	Source   string `db:"source"`
	Campaign string `db:"campaign"`
	TenantID string `db:"tenant_id"`
}

// LoadFromDB carrega a tabela tenant_routes para o mapa em memória.
func LoadFromDB(ctx context.Context, db *sqlx.DB) error {
	var rows []mapping
	if err := db.SelectContext(ctx, &rows,
		`SELECT source, COALESCE(campaign,'') AS campaign, tenant_id FROM tenant_routes`); err != nil {
		return err
	}
	m := make(TenantMap, len(rows))
	for _, r := range rows {
		m[r.Source+":"+r.Campaign] = r.TenantID
	}
	Load(m)
	return nil
}

// WatchDB recarrega o de-para periodicamente.
func WatchDB(ctx context.Context, db *sqlx.DB, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = LoadFromDB(ctx, db)
		}
	}
}
