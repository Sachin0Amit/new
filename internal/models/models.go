package models

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the current state of a sovereign task.
type TaskStatus string

const (
	StatusPending   TaskStatus = "PENDING"
	StatusRunning   TaskStatus = "RUNNING"
	StatusCompleted TaskStatus = "COMPLETED"
	StatusFailed    TaskStatus = "FAILED"
	StatusSensory   TaskStatus = "SENSORY_WAIT"
)

// HookPoint defines specific stages in the task lifecycle where plugins can intercept data.
type HookPoint string

const (
	HookPreProcessing  HookPoint = "PRE_PROCESSING"
	HookPreInference   HookPoint = "PRE_INFERENCE"
	HookPostInference  HookPoint = "POST_INFERENCE"
	HookToolExecution  HookPoint = "TOOL_EXECUTION"
)

// Task represents a unit of work for the Sovereign engine.
type Task struct {
	ID        uuid.UUID              `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Status    TaskStatus             `json:"status"`
	ReflexDepth int                  `json:"reflex_depth"` // Number of self-correction attempts
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Result    *TaskResult            `json:"result,omitempty"`
}

// CompletedAt returns the timestamp when the task was finalized.
func (t *Task) CompletedAt() time.Time {
	if t.Result != nil {
		return t.Result.Completed
	}
	return t.UpdatedAt
}

// TaskResult contains the output of a completed Task.
type TaskResult struct {
	Data      map[string]interface{} `json:"data"`
	Metrics   InferenceMetrics       `json:"metrics"`
	AuditTrail []AuditEntry           `json:"audit_trail,omitempty"`
	Signature []byte                 `json:"signature,omitempty"`
	Completed time.Time              `json:"completed_at"`
}

// AuditEntry represents a single verifiable step in the reasoning process.
type AuditEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Stage     string                 `json:"stage"`  // e.g., "RETRIEVAL", "PLUGIN", "INFERENCE"
	Action    string                 `json:"action"` // e.g., "Retrieved chunk 01", "Payload modified by plugin X"
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// InferenceMetrics tracks technical performance of the C++ engine.
type InferenceMetrics struct {
	TokensPerSecond float64 `json:"tokens_per_sec"`
	LatencyMS       int64   `json:"latency_ms"`
	MemoryUsedBytes int64   `json:"memory_used"`
	HardwareAgent   string  `json:"hardware_agent"`
}

// Document represents a high-level file ingested into the Sovereign Core.
type Document struct {
	ID        uuid.UUID `json:"id"`
	Path      string    `json:"path"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// Chunk represents a semantically coherent fragment of a Document.
type Chunk struct {
	ID         uuid.UUID `json:"id"`
	DocID      uuid.UUID `json:"doc_id"`
	Content    string    `json:"content"`
	Metadata   map[string]interface{} `json:"metadata"`
	Embedding  []float32 `json:"embedding,omitempty"`
}

// EmbeddingRequest defines the payload for generating vector representations.
type EmbeddingRequest struct {
	Text   string `json:"text"`
	Model  string `json:"model"`
	Config map[string]interface{} `json:"config"`
}

// NodePeer represents a discovered Sovereign node in the distributed fleet.
type NodePeer struct {
	ID        uuid.UUID `json:"id"`
	Address   string    `json:"address"`
	LastSeen  time.Time `json:"last_seen"`
	IsActive  bool      `json:"is_active"`
	Load      float64   `json:"cpu_load"`
}

// SensoryData encapsulates raw media buffers for the local vision/audio engines.
type SensoryData struct {
	Type     string `json:"type"` // "image", "audio", "video"
	MimeType string `json:"mime_type"`
	Buffer   []byte `json:"buffer"`
	Metadata map[string]interface{} `json:"metadata"`
}

// VisualFrame represents a single normalized frame for tensor processing.
type VisualFrame struct {
	Width    int       `json:"width"`
	Height   int       `json:"height"`
	Channels int       `json:"channels"`
	Data     []float32 `json:"data"` // Normalized pixel data
}

// PluginManifest provides telemetry data for an active system extension.
type PluginManifest struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	Capabilities []string `json:"capabilities"`
	CreatedAt   time.Time `json:"created_at"`
}
