package adapter

import (
	"encoding/json"

	"github.com/seu-org/motor-leads/internal/canonical"
)

// SiteAdapter trata leads enviados por formulários do site.
type SiteAdapter struct{}

func init() { Register("site", &SiteAdapter{}) }

type sitePayload struct {
	Nome     string `json:"nome"`
	Email    string `json:"email"`
	Telefone string `json:"telefone"`
	Campaign string `json:"campaign"`
	UTM      struct {
		Campaign string `json:"utm_campaign"`
	} `json:"utm"`
}

func (a *SiteAdapter) Parse(payload []byte) (*canonical.Lead, error) {
	var p sitePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}

	campaign := p.Campaign
	if campaign == "" {
		campaign = p.UTM.Campaign
	}

	return &canonical.Lead{
		Source:        "site",
		Campaign:      campaign,
		Nome:          p.Nome,
		Telefone:      p.Telefone,
		Email:         p.Email,
		SourcePayload: payload,
	}, nil
}
