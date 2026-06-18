package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	LeadsReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "leads_received_total", Help: "Leads recebidos por origem"},
		[]string{"source"},
	)
	LeadsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "leads_processed_total", Help: "Leads processados com sucesso"},
		[]string{"source"},
	)
	LeadsFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "leads_failed_total", Help: "Leads que falharam no processamento"},
		[]string{"source"},
	)
	ParseErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "leads_parse_errors_total", Help: "Erros de parse por origem"},
		[]string{"source"},
	)
	SchemaChanges = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "leads_schema_changes_total", Help: "Mudanças de schema detectadas"},
		[]string{"source"},
	)
)

func init() {
	prometheus.MustRegister(LeadsReceived, LeadsProcessed, LeadsFailed, ParseErrors, SchemaChanges)
}
