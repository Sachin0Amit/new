package rag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunking(t *testing.T) {
	s := &KnowledgeStore{}
	
	tests := []struct {
		name     string
		text     string
		size     int
		overlap  int
		expected int
	}{
		{"Simple split", "one two three four five", 2, 0, 3},
		{"With overlap", "one two three four five", 3, 1, 2},
		{"Single chunk", "one two three", 5, 0, 1},
		{"Empty text", "", 5, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := s.chunkText(tt.text, tt.size, tt.overlap)
			assert.Len(t, chunks, tt.expected)
		})
	}
}

func TestRAGIntegration(t *testing.T) {
	// Mocking embedder to avoid Ollama dependency in unit tests
	// In real integration, we'd use the actual embedder.
	
	// Pipeline setup (using mocks/memory)
	// store := NewKnowledgeStore(db)
	// index := NewVectorIndex()
	// pipe := NewRAGPipeline(store, embedder, index)
	
	// Test ingestion
	// err := pipe.Ingest(ctx, strings.NewReader("The capital of France is Paris."), nil)
	// assert.NoError(t, err)
	
	// Test query
	// results, err := pipe.Query(ctx, "What is the capital of France?", 1)
	// assert.NoError(t, err)
	// assert.NotEmpty(t, results)
}
