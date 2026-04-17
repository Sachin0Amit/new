package api

import (
	"context"
	"runtime"
	"time"
)

// TelemetryData represents the real-time state of a Sovereign node.
type TelemetryData struct {
	Timestamp      int64   `json:"timestamp"`
	NodeID         string  `json:"node_id"`
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsed     uint64  `json:"memory_used"`
	TaskThroughput float64 `json:"task_throughput"`
	ActiveInference bool    `json:"active_inference"`
}

// StartBroadcaster begins a background loop that pushes core telemetry to the Hub.
func StartBroadcaster(ctx context.Context, h *Hub) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			data := collectTelemetry()
			h.Broadcast(Message{
				Type:      EventMetrics,
				Payload:   data,
				Timestamp: time.Now().UnixMilli(),
			})
		}
	}
}

func collectTelemetry() TelemetryData {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return TelemetryData{
		Timestamp:      time.Now().UnixMilli(),
		NodeID:         "sovereign-node-alpha", // Would be configurable in prod
		CPUUsage:       float64(runtime.NumGoroutine()), // Proxy for load in this prototype
		MemoryUsed:     m.Alloc,
		TaskThroughput: 12.5, // Mock value, would link to pkg/metrics
		ActiveInference: true,
	}
}
