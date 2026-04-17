package models

import (
	"context"

	"github.com/google/uuid"
)

// Orchestrator defines the primary control interface for the Sovereign system.
type Orchestrator interface {
	// SubmitTask asynchronously starts a new computational or inference task.
	SubmitTask(ctx context.Context, task *Task) (uuid.UUID, error)

	// GetTask returns the current state and results of a specific task.
	GetTask(ctx context.Context, id uuid.UUID) (*Task, error)

	// Shutdown gracefully stops all processing.
	Shutdown() error
}

// InferenceEngine defines the contract for local processing (C++ bridge).
type InferenceEngine interface {
	// Infer performs a mathematical or linguistic derivation.
	Infer(ctx context.Context, payload map[string]interface{}) (*TaskResult, error)

	// GetStatus returns the performance metrics of the engine.
	GetStatus() (InferenceMetrics, error)
}

// StorageManager defines the contract for local state persistence.
type StorageManager interface {
	// Save persists an object to the local LSM-tree.
	Save(ctx context.Context, key string, data interface{}) error

	// Load retrieves an object from the local LSM-tree.
	Load(ctx context.Context, key string, out interface{}) error
}
