package rag

import (
	"math"
	"sync"

	"github.com/coder/hnsw"
)

type SearchResult struct {
	Chunk    Chunk   `json:"chunk"`
	Distance float32 `json:"distance"`
}

type VectorIndex struct {
	mu     sync.RWMutex
	graph  *hnsw.Graph[uint32]
	chunks map[uint32]Chunk
	nextID uint32
}

func NewVectorIndex() *VectorIndex {
	return &VectorIndex{
		graph:  hnsw.NewGraph[uint32](),
		chunks: make(map[uint32]Chunk),
	}
}

func (vi *VectorIndex) Add(chunk Chunk, vec []float32) {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	id := vi.nextID
	vi.nextID++
	
	node := hnsw.MakeNode(id, vec)
	vi.graph.Add(node)
	vi.chunks[id] = chunk
}

func (vi *VectorIndex) Search(queryVec []float32, topK int) ([]SearchResult, error) {
	vi.mu.RLock()
	defer vi.mu.RUnlock()

	// If fewer than 1000 vectors, do a flat scan for accuracy (as per requirements)
	if len(vi.chunks) < 1000 {
		return vi.flatSearch(queryVec, topK), nil
	}

	nodes := vi.graph.Search(queryVec, topK)
	
	var finalResults []SearchResult
	for _, node := range nodes {
		if chunk, ok := vi.chunks[node.Key]; ok {
			dist := vi.cosineDistance(queryVec, node.Value)
			finalResults = append(finalResults, SearchResult{
				Chunk:    chunk,
				Distance: dist,
			})
		}
	}
	return finalResults, nil
}

func (vi *VectorIndex) flatSearch(queryVec []float32, topK int) []SearchResult {
	var results []SearchResult
	for _, chunk := range vi.chunks {
		dist := vi.cosineDistance(queryVec, chunk.Embedding)
		results = append(results, SearchResult{
			Chunk:    chunk,
			Distance: dist,
		})
	}
	return results
}

func (vi *VectorIndex) cosineDistance(v1, v2 []float32) float32 {
	if len(v1) != len(v2) {
		return 2.0
	}
	var dot, n1, n2 float64
	for i := range v1 {
		dot += float64(v1[i] * v2[i])
		n1 += float64(v1[i] * v1[i])
		n2 += float64(v2[i] * v2[i])
	}
	if n1 == 0 || n2 == 0 {
		return 1.0
	}
	return float32(1.0 - (dot / (math.Sqrt(n1) * math.Sqrt(n2))))
}
