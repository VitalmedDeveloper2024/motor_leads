package queue

import (
	"context"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/seu-org/motor-leads/internal/canonical"
	"github.com/seu-org/motor-leads/internal/integration"
	"github.com/seu-org/motor-leads/internal/metrics"
)

var db *sqlx.DB

// Init injeta a conexão usada pelas funções de pacote.
func Init(d *sqlx.DB) { db = d }

// Enqueue persiste o lead (status já definido pelo handler).
func Enqueue(lead *canonical.Lead) error {
	const q = `
		INSERT INTO leads (tenant_id, source, campaign, nome, telefone, email,
		                   source_payload, status, error_log)
		VALUES (:tenant_id, :source, :campaign, :nome, :telefone, :email,
		        :source_payload, :status, :error_log)
		RETURNING id`
	rows, err := db.NamedQuery(q, lead)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		_ = rows.Scan(&lead.ID)
	}
	return nil
}

// ProcessPending roda em loop, com paralelismo controlado por semáforo.
func ProcessPending(ctx context.Context, workers int) {
	sem := make(chan struct{}, workers)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		leads, err := fetchPending(ctx, 100)
		if err != nil || len(leads) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}
		for _, lead := range leads {
			sem <- struct{}{}
			go func(l canonical.Lead) {
				defer func() { <-sem }()
				processLead(ctx, l)
			}(lead)
		}
	}
}

func processLead(ctx context.Context, lead canonical.Lead) {
	err := integration.CreateOdooLead(ctx, lead)
	if err == nil {
		err = integration.CreateChatwootContact(ctx, lead)
	}

	if err != nil {
		markFailed(ctx, lead.ID, err.Error())
		metrics.LeadsFailed.WithLabelValues(lead.Source).Inc()
		return
	}
	markProcessed(ctx, lead.ID)
	metrics.LeadsProcessed.WithLabelValues(lead.Source).Inc()
}

func fetchPending(ctx context.Context, limit int) ([]canonical.Lead, error) {
	// SKIP LOCKED evita que múltiplas instâncias peguem o mesmo lead.
	const q = `
		UPDATE leads SET status = 'pending', attempts = attempts + 1, updated_at = NOW()
		WHERE id IN (
			SELECT id FROM leads
			WHERE status = 'pending' AND tenant_id <> 'unknown'
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		RETURNING id, tenant_id, source, campaign, nome, telefone, email,
		          source_payload, status, attempts, error_log, created_at, updated_at`
	var leads []canonical.Lead
	if err := db.SelectContext(ctx, &leads, q, limit); err != nil {
		return nil, err
	}
	return leads, nil
}

func markProcessed(ctx context.Context, id string) {
	_, _ = db.ExecContext(ctx,
		`UPDATE leads SET status='processed', error_log='', updated_at=NOW() WHERE id=$1`, id)
}

func markFailed(ctx context.Context, id, errLog string) {
	_, _ = db.ExecContext(ctx,
		`UPDATE leads SET status='failed', error_log=$2, updated_at=NOW() WHERE id=$1`, id, errLog)
}

// ---- API usada pela interface de reprocessamento -------------------------

// List devolve leads filtrados por status (vazio = todos).
func List(ctx context.Context, status string, limit int) ([]canonical.Lead, error) {
	var leads []canonical.Lead
	q := `SELECT id, tenant_id, source, campaign, nome, telefone, email,
	             source_payload, status, attempts, error_log, created_at, updated_at
	      FROM leads`
	args := []any{}
	if status != "" {
		q += ` WHERE status = $1`
		args = append(args, status)
	}
	q += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(len(args)+1)
	args = append(args, limit)
	if err := db.SelectContext(ctx, &leads, q, args...); err != nil {
		return nil, err
	}
	return leads, nil
}

// Retry recoloca um lead 'failed' como 'pending' para o worker reprocessar.
func Retry(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE leads SET status='pending', updated_at=NOW() WHERE id=$1 AND status='failed'`, id)
	return err
}
