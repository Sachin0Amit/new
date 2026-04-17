package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// InferenceTasksTotal tracks the total number of sovereign inference tasks processed.
	InferenceTasksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sovereign_inference_tasks_total",
		Help: "The total number of inference tasks processed by the Titan core",
	}, []string{"status"})

	// InferenceLatency tracks the time spent on each derivation task.
	InferenceLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sovereign_inference_latency_seconds",
		Help:    "Time spent performing mathematical or linguistic derivations",
		Buckets: prometheus.DefBuckets,
	}, []string{"engine"})

	// MemoryUtilization tracks the local hardware memory consumption of the Titan engine.
	MemoryUtilization = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sovereign_memory_utilization_bytes",
		Help: "Current memory consumption of the Titan C++ core",
	})
)

// RecordTaskCompletion increments the success/failure counter.
func RecordTaskCompletion(status string) {
	InferenceTasksTotal.WithLabelValues(status).Inc()
}

// RecordLatency records the duration of a task.
func RecordLatency(engine string, durationSeconds float64) {
	InferenceLatency.WithLabelValues(engine).Observe(durationSeconds)
}

// SetMemoryUsage updates the memory gauge.
func SetMemoryUsage(bytes int64) {
	MemoryUtilization.Set(float64(bytes))
}
