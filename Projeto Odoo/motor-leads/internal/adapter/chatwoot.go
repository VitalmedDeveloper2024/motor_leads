package adapter

import (
	"encoding/json"

	"github.com/seu-org/motor-leads/internal/canonical"
)

// ChatwootAdapter trata o Chatwoot no papel de ORIGEM (webhook de
// contato/conversa criada). O papel de DESTINO fica em internal/integration.
type ChatwootAdapter struct{}

func init() { Register("chatwoot", &ChatwootAdapter{}) }

type chatwootPayload struct {
	Event   string `json:"event"`
	Contact struct {
		Name        string `json:"name"`
		PhoneNumber string `json:"phone_number"`
		Email       string `json:"email"`
	} `json:"contact"`
	Conversation struct {
		AdditionalAttributes struct {
			Campaign string `json:"campaign"`
		} `json:"additional_attributes"`
	} `json:"conversation"`
	AdditionalAttributes struct {
		Campaign string `json:"campaign"`
	} `json:"additional_attributes"`
}

func (a *ChatwootAdapter) Parse(payload []byte) (*canonical.Lead, error) {
	var p chatwootPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}

	campaign := p.Conversation.AdditionalAttributes.Campaign
	if campaign == "" {
		campaign = p.AdditionalAttributes.Campaign
	}

	return &canonical.Lead{
		Source:        "chatwoot",
		Campaign:      campaign,
		Nome:          p.Contact.Name,
		Telefone:      p.Contact.PhoneNumber,
		Email:         p.Contact.Email,
		SourcePayload: payload,
	}, nil
}
