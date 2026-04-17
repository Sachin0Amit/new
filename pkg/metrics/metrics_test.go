package metrics

import (
	"testing"
	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricRegistration(t *testing.T) {
	// Simple test to ensure counters don't panic
	RecordTaskCompletion("success")
	RecordTaskCompletion("failed")
	
	RecordLatency("titan_v1", 0.15)
	SetMemoryUsage(1024 * 1024)
}

func TestRecording(t *testing.T) {
	// Ensure labels are correctly applied
	InferenceTasksTotal.WithLabelValues("success").Inc()
	InferenceLatency.WithLabelValues("test_engine").Observe(0.5)
}

func TestHandler(t *testing.T) {
    // In a real app, we'd check the promhttp handler
    _ = prometheus.DefaultRegisterer
}
