# Motor de Leads

Microserviço em Go para ingestão, normalização e roteamento de leads de múltiplas
origens (Meta Ads, Chatwoot/WhatsApp, Site; extensível para Google Ads, TikTok)
para **Odoo v18** (multiempresa) e **Chatwoot** (multiconta).

## Subir tudo
```bash
docker compose up --build
# UI de reprocessamento: http://localhost:8080
# Métricas:             http://localhost:8080/metrics
```

## Rodar local (sem Docker)
```bash
go mod tidy
createdb leads
for f in db/migrations/*.sql; do psql leads -f "$f"; done
export DATABASE_URL=postgres://leads:leads@localhost:5432/leads?sslmode=disable
make run
```

## Endpoints
| Método | Rota                     | Descrição                  |
|--------|--------------------------|----------------------------|
| POST   | /webhook/meta            | Webhook Meta Lead Ads      |
| POST   | /webhook/chatwoot        | Webhook Chatwoot (origem)  |
| POST   | /webhook/site            | Formulário do site         |
| GET    | /api/leads?status=failed | Lista leads (UI)           |
| POST   | /api/leads/retry?id=...  | Reprocessa um lead falho   |
| GET    | /metrics                 | Prometheus                 |
| GET    | /healthz                 | Healthcheck                |

## Adicionar uma nova origem (ex.: Google Ads)
1. `internal/adapter/google_ads.go` implementando `Adapter` + `Register("google_ads", ...)` no `init()`.
2. Uma linha em `cmd/server/main.go`: `mux.HandleFunc("/webhook/google_ads", webhookHandler(db, "google_ads"))`.
3. Nada mais muda. O core não conhece origens.

## Regras do projeto
- Lead com tenant desconhecido vai para `failed` (nunca é descartado).
- `source_payload` (jsonb) é sempre salvo, mesmo em parse parcial.
- Reprocessamento usa semáforo (`chan struct{}`) — sem goroutines ilimitadas.
- `fetchPending` usa `FOR UPDATE SKIP LOCKED` (seguro com múltiplas instâncias).
- Toda falha vai para `error_log` + métrica Prometheus. Sem falha silenciosa.

## Credenciais por tenant (a implementar)
As integrações usam resolvers injetáveis. Implemente lendo `tenant_credentials`:

```go
integration.SetOdooResolver(func(tid string) (integration.OdooConfig, error) {
    // SELECT ... FROM tenant_credentials WHERE tenant_id = tid
})
integration.SetChatwootResolver(func(tid string) (integration.ChatwootConfig, error) {
    // idem
})
```

## A definir com o negócio
- Isolamento de tenant: filtro por `tenant_id` (atual) vs. row-level security.
- Campos canônicos extras (endereço, UTMs completas, etc.).
- Origem do de-para `tenant_routes`: banco (atual) vs. arquivo de config.

## Estrutura
```
motor-leads/
├── cmd/server/main.go
├── internal/
│   ├── adapter/      # interface + registry + meta/chatwoot/site
│   ├── canonical/    # modelo canônico de lead
│   ├── tenant/       # resolver + loader do de-para
│   ├── queue/        # fila PostgreSQL + retry paralelo
│   ├── integration/  # odoo + chatwoot (destinos)
│   ├── schema/       # detecção de mudança de schema
│   ├── metrics/      # Prometheus
│   ├── store/        # conexão Postgres
│   └── config/       # env
├── db/migrations/
├── frontend/         # UI de reprocessamento
├── docker-compose.yml
├── Dockerfile
└── Makefile
```
