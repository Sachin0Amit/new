package vector

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/Sachin0Amit/new/pkg/errors"
)

// VectorSimilarity defines the logic for comparing high-dimensional embeddings.
type VectorSimilarity interface {
	Compare(a, b []float32) (float32, error)
}

// CosineSimilarity implements the dot product based similarity metric.
type CosineSimilarity struct{}

func (c *CosineSimilarity) Compare(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New(errors.CodeValidation, "vector dimensions must match", nil)
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0, nil
	}

	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB))), nil
}

// EuclideanDistance implements the L2 distance metric.
type EuclideanDistance struct{}

func (e *EuclideanDistance) Compare(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New(errors.CodeValidation, "vector dimensions must match", nil)
	}

	var sum float64
	for i := range a {
		diff := float64(a[i] - b[i])
		sum += diff * diff
	}

	return float32(math.Sqrt(sum)), nil
}

// Search calculates the similarity of a query vector against a candidate set.
// It returns a slice of indices ordered by descending similarity (for Cosine).
func Search(query []float32, candidates [][]float32, metric VectorSimilarity) ([]int, []float32, error) {
	scores := make([]float32, len(candidates))
	indices := make([]int, len(candidates))

	for i, candidate := range candidates {
		score, err := metric.Compare(query, candidate)
		if err != nil {
			return nil, nil, err
		}
		scores[i] = score
		indices[i] = i
	}

	// Simple selection sort for demo, would use a max-heap/priority-queue for production
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i] < scores[j] { // Descending for Cosine
				scores[i], scores[j] = scores[j], scores[i]
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}

	return indices, scores, nil
}

// VectorStore defines the interface for persisting and querying embeddings.
type VectorStore interface {
	Save(ctx context.Context, id string, vec []float32) error
	Get(ctx context.Context, id string) ([]float32, error)
	ListAll(ctx context.Context) (map[string][]float32, error)
}

// BadgerFlatStore implements VectorStore using the project's primary BadgerDB instance.
type BadgerFlatStore struct {
	SaveFunc func(ctx context.Context, key string, data interface{}) error
	LoadFunc func(ctx context.Context, key string, out interface{}) error
}

func (b *BadgerFlatStore) Save(ctx context.Context, id string, vec []float32) error {
	key := fmt.Sprintf("vec:%s", id)
	return b.SaveFunc(ctx, key, vec)
}

func (b *BadgerFlatStore) Get(ctx context.Context, id string) ([]float32, error) {
	key := fmt.Sprintf("vec:%s", id)
	var vec []float32
	err := b.LoadFunc(ctx, key, &vec)
	return vec, err
}

func (b *BadgerFlatStore) ListAll(ctx context.Context) (map[string][]float32, error) {
	// For prototype, this would utilize a specialized iterator in internal/storage
	return nil, nil 
}

// Float32ToBytes converts a vector to a raw byte buffer for storage.
func Float32ToBytes(vec []float32) []byte {
	buf := make([]byte, len(vec)*4)
	for i, v := range vec {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}
