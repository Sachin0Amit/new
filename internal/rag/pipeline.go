package rag

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/google/uuid"
)

type RAGPipeline struct {
	Store    *KnowledgeStore
	Embedder *Embedder
	Index    *VectorIndex
}

func NewRAGPipeline(store *KnowledgeStore, embedder *Embedder, index *VectorIndex) *RAGPipeline {
	return &RAGPipeline{
		Store:    store,
		Embedder: embedder,
		Index:    index,
	}
}

func (p *RAGPipeline) Ingest(ctx context.Context, source io.Reader, metadata map[string]string) error {
	content, err := ioutil.ReadAll(source)
	if err != nil {
		return err
	}

	docID := uuid.New().String()
	doc := Document{
		ID:       docID,
		Content:  string(content),
		Metadata: metadata,
	}

	// 1. Persist to Badger
	if err := p.Store.Store(ctx, doc); err != nil {
		return err
	}

	// 2. Load back chunks (to get IDs/metadata) and embed
	chunks, err := p.Store.Retrieve(docID)
	if err != nil {
		return err
	}

	for i := range chunks {
		vec, err := p.Embedder.Embed(chunks[i].Text)
		if err != nil {
			return err
		}
		chunks[i].Embedding = vec
		
		// 3. Add to HNSW Index
		p.Index.Add(chunks[i], vec)
	}

	return nil
}

func (p *RAGPipeline) Query(ctx context.Context, question string, topK int) ([]SearchResult, error) {
	vec, err := p.Embedder.Embed(question)
	if err != nil {
		return nil, err
	}

	return p.Index.Search(vec, topK)
}
