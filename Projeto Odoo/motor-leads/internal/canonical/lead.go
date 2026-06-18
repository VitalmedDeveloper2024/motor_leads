package canonical

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusProcessed Status = "processed"
	StatusFailed    Status = "failed"
)

// Lead é o modelo canônico, independente da origem.
type Lead struct {
	ID            string          `db:"id"             json:"id"`
	TenantID      string          `db:"tenant_id"      json:"tenant_id"`
	Source        string          `db:"source"         json:"source"`   // "meta", "chatwoot", "site"
	Campaign      string          `db:"campaign"       json:"campaign"`
	Nome          string          `db:"nome"           json:"nome"`
	Telefone      string          `db:"telefone"       json:"telefone"`
	Email         string          `db:"email"          json:"email"`
	SourcePayload json.RawMessage `db:"source_payload" json:"source_payload"` // payload bruto (jsonb)
	Status        Status          `db:"status"         json:"status"`
	Attempts      int             `db:"attempts"       json:"attempts"`
	ErrorLog      string          `db:"error_log"      json:"error_log"`
	CreatedAt     time.Time       `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"     json:"updated_at"`
}
