package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/seu-org/motor-leads/internal/adapter"
	"github.com/seu-org/motor-leads/internal/canonical"
	"github.com/seu-org/motor-leads/internal/config"
	"github.com/seu-org/motor-leads/internal/metrics"
	"github.com/seu-org/motor-leads/internal/queue"
	"github.com/seu-org/motor-leads/internal/schema"
	"github.com/seu-org/motor-leads/internal/store"
	"github.com/seu-org/motor-leads/internal/tenant"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := store.Connect(ctx, cfg.DSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	queue.Init(db)

	// Carrega de-para origem/campanha -> tenant e mantém atualizado.
	if err := tenant.LoadFromDB(ctx, db); err != nil {
		log.Printf("aviso: não foi possível carregar tenant_routes: %v", err)
	}
	go tenant.WatchDB(ctx, db, 60*time.Second)

	// Worker de reprocessamento.
	go queue.ProcessPending(ctx, cfg.Workers)

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/meta", webhookHandler(db, "meta"))
	mux.HandleFunc("/webhook/chatwoot", webhookHandler(db, "chatwoot"))
	mux.HandleFunc("/webhook/site", webhookHandler(db, "site"))
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })

	// API + UI de reprocessamento.
	mux.HandleFunc("/api/leads", apiListLeads(db))
	mux.HandleFunc("/api/leads/retry", apiRetryLead(db))
	mux.Handle("/", http.FileServer(http.Dir("frontend")))

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: logMW(mux)}
	go func() {
		log.Printf("ouvindo em %s | origens: %s", cfg.HTTPAddr, strings.Join(adapter.Sources(), ", "))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	<-ctx.Done()
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
	log.Println("encerrado")
}

func webhookHandler(db *sqlx.DB, source string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "erro ao ler body", http.StatusBadRequest)
			return
		}

		a, ok := adapter.Get(source)
		if !ok {
			http.Error(w, "origem desconhecida", http.StatusBadRequest)
			return
		}

		// Detecta mudança de schema (não bloqueia o fluxo).
		go schema.Detect(context.Background(), db, source, body)

		lead, err := a.Parse(body)
		if err != nil {
			metrics.ParseErrors.WithLabelValues(source).Inc()
			http.Error(w, "falha no parse", http.StatusUnprocessableEntity)
			return
		}

		lead.TenantID, err = tenant.Resolve(lead.Source, lead.Campaign)
		if err != nil {
			// Não perde o lead: sinaliza mapeamento inesperado.
			lead.TenantID = "unknown"
			lead.ErrorLog = err.Error()
			lead.Status = canonical.StatusFailed
		} else {
			lead.Status = canonical.StatusPending
		}

		if err := queue.Enqueue(lead); err != nil {
			http.Error(w, "erro ao enfileirar", http.StatusInternalServerError)
			return
		}

		metrics.LeadsReceived.WithLabelValues(source).Inc()
		w.WriteHeader(http.StatusAccepted)
	}
}

func apiListLeads(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		leads, err := queue.List(r.Context(), status, 200)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(leads)
	}
}

func apiRetryLead(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "use POST", http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id obrigatório", http.StatusBadRequest)
			return
		}
		if err := queue.Retry(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func logMW(next http.Handler) http.Handler {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
