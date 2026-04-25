package reflex

import (
	"context"
)

// Issue represents a detected system anomaly or failure.
type Issue struct {
	Component string
	Severity  string
	Message   string
	Context   map[string]interface{}
}

// SelfHealer defines the autonomous monitoring and repair contract.
type SelfHealer interface {
	// Monitor checks the system for anomalies and returns a list of detected issues.
	Monitor(ctx context.Context) ([]Issue, error)
	
	// Heal attempts to resolve a specific issue.
	Heal(ctx context.Context, issue Issue) error
	
	// Start begins the autonomous background monitoring loop.
	Start(ctx context.Context)
}
