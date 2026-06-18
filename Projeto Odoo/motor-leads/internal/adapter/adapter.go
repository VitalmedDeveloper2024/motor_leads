package adapter

import "github.com/seu-org/motor-leads/internal/canonical"

// Adapter é o contrato que toda origem deve implementar.
// Adicionar nova origem = novo struct que implementa esta interface.
type Adapter interface {
	Parse(payload []byte) (*canonical.Lead, error)
}

var registry = map[string]Adapter{}

func Register(source string, a Adapter) {
	registry[source] = a
}

func Get(source string) (Adapter, bool) {
	a, ok := registry[source]
	return a, ok
}

// Sources devolve as origens registradas (útil para health/admin).
func Sources() []string {
	out := make([]string, 0, len(registry))
	for s := range registry {
		out = append(out, s)
	}
	return out
}
