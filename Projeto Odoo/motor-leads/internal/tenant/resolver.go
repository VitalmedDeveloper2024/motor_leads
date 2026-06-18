package tenant

import (
	"fmt"
	"sync"
)

// TenantMap mapeia (source, campaign) -> tenant_id.
// Chave: "source:campaign". Fallback: "source:".
type TenantMap map[string]string

var (
	mu         sync.RWMutex
	defaultMap = TenantMap{}
)

// Load substitui o mapa de-para em memória (chamar na startup e em reloads).
func Load(m TenantMap) {
	mu.Lock()
	defer mu.Unlock()
	defaultMap = m
}

func Resolve(source, campaign string) (string, error) {
	mu.RLock()
	defer mu.RUnlock()

	if tid, ok := defaultMap[source+":"+campaign]; ok {
		return tid, nil
	}
	if tid, ok := defaultMap[source+":"]; ok { // fallback por source
		return tid, nil
	}
	return "", fmt.Errorf("tenant não encontrado para source=%s campaign=%s", source, campaign)
}
