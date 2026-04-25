package retrieval

import (
	"context"
	"fmt"
	"strings"

	"github.com/Sachin0Amit/new/internal/models"
	"github.com/Sachin0Amit/new/pkg/vector"
)

// KnowledgeMesh orchestrates semantic search across various memory domains.
type KnowledgeMesh struct {
	vectorStore vector.VectorStore
	similarity  vector.VectorSimilarity
	broker      syncBroker // Direct interface dependency
}

type syncBroker interface {
	FederatedSearch(ctx context.Context, query []float32, topK int) ([]models.Chunk, error)
}

// NewKnowledgeMesh initializes the retrieval engine with a backing vector store and optional mesh broker.
func NewKnowledgeMesh(store vector.VectorStore, broker interface{}) *KnowledgeMesh {
	k := &KnowledgeMesh{
		vectorStore: store,
		similarity:  &vector.CosineSimilarity{},
	}
	
	if b, ok := broker.(syncBroker); ok {
		k.broker = b
	}
	
	return k
}

// Search retrieves the top-K relevant chunks for a given query embedding.
func (k *KnowledgeMesh) Search(ctx context.Context, query []float32, topK int) ([]models.Chunk, error) {
	// For simulation, we would list all vectors, compare, and sort.
	// In production, this utilizes an HNSW index or similar.
	all, _ := k.vectorStore.ListAll(ctx)
	if len(all) == 0 {
		return nil, nil
	}

	var candidates [][]float32
	var ids []string
	for id, vec := range all {
		candidates = append(candidates, vec)
		ids = append(ids, id)
	}

	indices, _, err := vector.Search(query, candidates, k.similarity)
	if err != nil {
		return nil, err
	}
	var results []models.Chunk
	for i := 0; i < len(indices) && i < topK; i++ {
		// Mock reconstruction of chunks for the prototype
		results = append(results, models.Chunk{
			Content: fmt.Sprintf("Retrieved Context from %s", ids[indices[i]]),
		})
	}

	// 2. Perform Federated Search if mesh connectivity is available
	if k.broker != nil && len(results) < topK {
		remoteResults, _ := k.broker.FederatedSearch(ctx, query, topK-len(results))
		results = append(results, remoteResults...)
	}

	return results, nil
}

// BuildContext formats retrieved chunks into a standard memory block for derivation.
func BuildContext(chunks []models.Chunk) string {
	if len(chunks) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("### RELEVANT SYSTEM MEMORY ###\n")
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, chunk.Content))
	}
	sb.WriteString("##############################\n")
	
	return sb.String()
}
