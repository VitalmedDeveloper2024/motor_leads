package adapter

import (
	"encoding/json"

	"github.com/seu-org/motor-leads/internal/canonical"
)

type MetaAdapter struct{}

func init() { Register("meta", &MetaAdapter{}) }

type metaPayload struct {
	LeadgenID string `json:"leadgen_id"`
	FieldData []struct {
		Name   string   `json:"name"`
		Values []string `json:"values"`
	} `json:"field_data"`
	CampaignName string `json:"campaign_name"`
	FormID       string `json:"form_id"`
}

func (a *MetaAdapter) Parse(payload []byte) (*canonical.Lead, error) {
	var p metaPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}

	lead := &canonical.Lead{
		Source:        "meta",
		Campaign:      p.CampaignName,
		SourcePayload: payload,
	}

	for _, f := range p.FieldData {
		val := ""
		if len(f.Values) > 0 {
			val = f.Values[0]
		}
		switch f.Name {
		case "full_name", "nome":
			lead.Nome = val
		case "phone_number", "telefone":
			lead.Telefone = val
		case "email":
			lead.Email = val
		}
	}

	return lead, nil
}
