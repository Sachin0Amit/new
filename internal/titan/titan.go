package titan

import (
	"context"
)

// Engine is the unified interface for the Titan Inference Engine.
type Engine interface {
	Derive(ctx context.Context, prompt string, maxTokens int) (string, error)
	Close() error
}
