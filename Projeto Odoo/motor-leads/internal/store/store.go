package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Connect abre a conexão com o PostgreSQL com retry simples.
func Connect(ctx context.Context, dsn string) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sqlx.Open("postgres", dsn)
		if err == nil {
			if pErr := db.PingContext(ctx); pErr == nil {
				db.SetMaxOpenConns(20)
				db.SetMaxIdleConns(10)
				db.SetConnMaxLifetime(time.Hour)
				return db, nil
			} else {
				err = pErr
			}
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("não foi possível conectar ao postgres: %w", err)
}
