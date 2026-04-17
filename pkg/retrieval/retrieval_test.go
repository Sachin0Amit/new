package retrieval

import (
	"context"
	"testing"

	"github.com/papi-ai/sovereign-core/internal/models"
	"github.com/stretchr/testify/assert"
)

// MockStore implements vector.VectorStore for testing.
type MockStore struct {
	data map[string][]float32
}

func (m *MockStore) Save(ctx context.Context, id string, vec []float32) error          { m.data[id] = vec; return nil }
func (m *MockStore) Get(ctx context.Context, id string) ([]float32, error)            { return m.data[id], nil }
func (m *MockStore) ListAll(ctx context.Context) (map[string][]float32, error)      { return m.data, nil }

func TestKnowledgeMeshSearch(t *testing.T) {
	store := &MockStore{data: map[string][]float32{
		"1": {1.0, 0.0},
		"2": {0.0, 1.0},
	}}

	mesh := NewKnowledgeMesh(store, nil)
	query := []float32{0.9, 0.1}

	chunks, err := mesh.Search(context.Background(), query, 1)
	assert.NoError(t, err)
	assert.Len(t, chunks, 1)
}

func TestBuildContext(t *testing.T) {
	chunks := []models.Chunk{
		{Content: "Sovereignty is the high road."},
		{Content: "Intelligence is local."},
	}

	ctxStr := BuildContext(chunks)
	assert.Contains(t, ctxStr, "### RELEVANT SYSTEM MEMORY ###")
	assert.Contains(t, ctxStr, "Sovereignty is the high road.")
	assert.Contains(t, ctxStr, "[1]")
}
