package auditor

import (
	"context"
	"time"
)

// Entry represents a single audit record.
type Entry struct {
	Timestamp time.Time
	Actor     string
	Action    string
	Resource  string
	Status    string
	Metadata  map[string]interface{}
}

// Auditor defines the contract for recording system actions and states.
type Auditor interface {
	// Log records a new system action to the audit trail.
	Log(ctx context.Context, entry Entry) error
	
	// GetHistory retrieves audit entries matching the specified criteria.
	GetHistory(ctx context.Context, actor string, limit int) ([]Entry, error)
	
	// Flush ensures all pending audit entries are persisted.
	Flush() error
}
