package etl

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/Sachin0Amit/new/internal/models"
)

// NewTextChunker creates a stage that splits raw text into semantic chunks of a fixed size.
func NewTextChunker(chunkSize int) Stage {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		text, ok := input.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected input type: %T", input)
		}

		var chunks []models.Chunk
		docID := uuid.New() // Simplified for demo
		
		words := strings.Fields(text)
		for i := 0; i < len(words); i += chunkSize {
			end := i + chunkSize
			if end > len(words) {
				end = len(words)
			}
			
			chunks = append(chunks, models.Chunk{
				ID:      uuid.New(),
				DocID:   docID,
				Content: strings.Join(words[i:end], " "),
			})
		}

		return chunks, nil
	}
}

// NewMockEmbeddingStage creates a stage that simulates vector generation for testing.
func NewMockEmbeddingStage(dim int) Stage {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		chunks, ok := input.([]models.Chunk)
		if !ok {
			return nil, fmt.Errorf("unexpected input type: %T", input)
		}

		for i := range chunks {
			// Generate deterministic mock embedding based on content
			chunks[i].Embedding = make([]float32, dim)
			for d := 0; d < dim; d++ {
				chunks[i].Embedding[d] = float32(d) / float32(dim)
			}
		}

		return chunks, nil
	}
}
