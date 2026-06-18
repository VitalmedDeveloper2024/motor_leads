package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/seu-org/motor-leads/internal/canonical"
)

// OdooConfig — credenciais por tenant (carregar de tenant_credentials).
type OdooConfig struct {
	BaseURL   string
	DB        string
	UID       int
	APIKey    string
	CompanyID int
}

// odooConfigFor retorna a config do tenant.
// TODO(negócio): buscar de tenant_credentials no PostgreSQL.
var odooConfigFor = func(tenantID string) (OdooConfig, error) {
	return OdooConfig{}, fmt.Errorf("credenciais Odoo não configuradas para tenant=%s", tenantID)
}

// SetOdooResolver permite injetar a fonte de credenciais (ex.: do banco).
func SetOdooResolver(fn func(tenantID string) (OdooConfig, error)) { odooConfigFor = fn }

var httpClient = &http.Client{Timeout: 15 * time.Second}

// CreateOdooLead cria um crm.lead no Odoo v18 via JSON-RPC.
func CreateOdooLead(ctx context.Context, lead canonical.Lead) error {
	cfg, err := odooConfigFor(lead.TenantID)
	if err != nil {
		return err
	}

	values := map[string]any{
		"name":         fmt.Sprintf("Lead %s (%s)", lead.Nome, lead.Source),
		"contact_name": lead.Nome,
		"phone":        lead.Telefone,
		"email_from":   lead.Email,
		"company_id":   cfg.CompanyID, // empresa correta do tenant
		"description":  fmt.Sprintf("Origem: %s | Campanha: %s", lead.Source, lead.Campaign),
	}

	payload := map[string]any{
		"jsonrpc": "2.0",
		"method":  "call",
		"params": map[string]any{
			"service": "object",
			"method":  "execute_kw",
			"args": []any{
				cfg.DB, cfg.UID, cfg.APIKey,
				"crm.lead", "create", []any{values},
			},
		},
	}
	return rpcCall(ctx, cfg.BaseURL+"/jsonrpc", payload)
}

func rpcCall(ctx context.Context, url string, payload any) error {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var out struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("odoo: resposta inválida: %w", err)
	}
	if out.Error != nil {
		return fmt.Errorf("odoo: %s", out.Error.Message)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("odoo: status http %d", resp.StatusCode)
	}
	return nil
}
