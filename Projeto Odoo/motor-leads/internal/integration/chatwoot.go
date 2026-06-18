package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/seu-org/motor-leads/internal/canonical"
)

// ChatwootConfig — credenciais por tenant/conta (carregar de tenant_credentials).
type ChatwootConfig struct {
	BaseURL   string
	AccountID int
	APIToken  string
	InboxID   int
}

var chatwootConfigFor = func(tenantID string) (ChatwootConfig, error) {
	return ChatwootConfig{}, fmt.Errorf("credenciais Chatwoot não configuradas para tenant=%s", tenantID)
}

func SetChatwootResolver(fn func(tenantID string) (ChatwootConfig, error)) { chatwootConfigFor = fn }

// CreateChatwootContact cria (ou atualiza) um contato na conta do tenant.
func CreateChatwootContact(ctx context.Context, lead canonical.Lead) error {
	cfg, err := chatwootConfigFor(lead.TenantID)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/accounts/%d/contacts", cfg.BaseURL, cfg.AccountID)
	body, _ := json.Marshal(map[string]any{
		"name":         lead.Nome,
		"email":        lead.Email,
		"phone_number": lead.Telefone,
		"inbox_id":     cfg.InboxID,
		"custom_attributes": map[string]any{
			"source":   lead.Source,
			"campaign": lead.Campaign,
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_access_token", cfg.APIToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("chatwoot: status http %d", resp.StatusCode)
	}
	return nil
}
