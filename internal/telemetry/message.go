package telemetry

import (
	"time"
)

// MessageType defines the classification of telemetry events.
type MessageType string

const (
	DerivationStarted   MessageType = "DERIVATION_STARTED"
	DerivationStep      MessageType = "DERIVATION_STEP"
	DerivationCompleted MessageType = "DERIVATION_COMPLETED"
	DerivationFailed    MessageType = "DERIVATION_FAILED"
	ReflexTriggered     MessageType = "REFLEX_TRIGGERED"
	NodeHeartbeat      MessageType = "NODE_HEARTBEAT"
	FleetUpdate        MessageType = "FLEET_UPDATE"
)

// TelemetryMessage is the standard envelope for all WebSocket communication.
type TelemetryMessage struct {
	Type      MessageType `json:"type"`
	NodeID    string      `json:"node_id"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// Payload schemas for specific event types
type DerivationPayload struct {
	TaskID string `json:"task_id"`
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
	Step   int    `json:"step,omitempty"`
}

type HeartbeatPayload struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	Uptime      int64   `json:"uptime"`
}
