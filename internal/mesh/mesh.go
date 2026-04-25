package mesh

import (
	"context"
)

// Result represents a knowledge retrieval unit.
type Result struct {
	ID       string
	Content  string
	Distance float64
	Metadata map[string]interface{}
}

// KnowledgeMesh defines the distributed RAG and state persistence layer.
type KnowledgeMesh interface {
	// Store persists a value associated with a key in the Knowledge Mesh.
	Store(ctx context.Context, key string, value interface{}) error
	
	// Retrieve fetches a value from the Knowledge Mesh.
	Retrieve(ctx context.Context, key string, out interface{}) error
	
	// Query performs a semantic or keyword search across the Knowledge Mesh.
	Query(ctx context.Context, query string, limit int) ([]Result, error)
	
	// Close safely shuts down the Knowledge Mesh storage engine.
	Close() error
}
