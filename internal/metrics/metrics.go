package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ActiveDerivations = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_derivations",
		Help: "Current number of active AI derivations",
	})

	DerivationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "derivation_duration_seconds",
		Help:    "Latency of AI derivations in seconds",
		Buckets: []float64{0.1, 0.5, 1, 5, 30, 120},
	})

	DerivationTokens = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "derivation_tokens_total",
		Help: "Total tokens generated across the fleet",
	}, []string{"node_id"})

	ReflexCorrections = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reflex_corrections_total",
		Help: "Total number of autonomous reflex self-healing actions",
	}, []string{"anomaly_type"})

	P2PPeersConnected = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "p2p_peers_connected",
		Help: "Number of active P2P mesh peers",
	}, []string{"node_id"})

	KnowledgeMeshEntries = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "knowledge_mesh_entries",
		Help: "Total entries in the BadgerDB knowledge mesh",
	})

	WebSocketConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "websocket_connections_active",
		Help: "Number of active telemetry WebSocket clients",
	})

	AuthFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_failures_total",
		Help: "Total number of authentication failures",
	}, []string{"reason"})
)

// Handler returns the Prometheus metrics HTTP handler.
func Handler() http.Handler {
	return promhttp.Handler()
}
