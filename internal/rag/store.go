package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

type Document struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

type Chunk struct {
	DocID      string            `json:"doc_id"`
	Index      int               `json:"index"`
	Text       string            `json:"text"`
	Embedding  []float32         `json:"embedding,omitempty"`
	Metadata   map[string]string `json:"metadata"`
}

type KnowledgeStore struct {
	db *badger.DB
}

func NewKnowledgeStore(db *badger.DB) *KnowledgeStore {
	return &KnowledgeStore{db: db}
}

// Store chunks a document and persists it.
func (s *KnowledgeStore) Store(ctx context.Context, doc Document) error {
	chunks := s.chunkText(doc.Content, 512, 64)
	
	return s.db.Update(func(txn *badger.Txn) error {
		var chunkKeys []string
		for i, text := range chunks {
			chunk := Chunk{
				DocID:    doc.ID,
				Index:    i,
				Text:     text,
				Metadata: doc.Metadata,
			}
			
			key := fmt.Sprintf("chunk:%s:%d", doc.ID, i)
			val, _ := json.Marshal(chunk)
			if err := txn.Set([]byte(key), val); err != nil {
				return err
			}
			chunkKeys = append(chunkKeys, key)
		}
		
		// Store document index
		idxKey := fmt.Sprintf("doc:%s", doc.ID)
		idxVal, _ := json.Marshal(chunkKeys)
		return txn.Set([]byte(idxKey), idxVal)
	})
}

func (s *KnowledgeStore) Retrieve(docID string) ([]Chunk, error) {
	var chunks []Chunk
	err := s.db.View(func(txn *badger.Txn) error {
		idxKey := fmt.Sprintf("doc:%s", docID)
		item, err := txn.Get([]byte(idxKey))
		if err != nil {
			return err
		}
		
		var chunkKeys []string
		_ = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &chunkKeys)
		})
		
		for _, key := range chunkKeys {
			chunkItem, err := txn.Get([]byte(key))
			if err != nil {
				continue
			}
			var c Chunk
			_ = chunkItem.Value(func(val []byte) error {
				return json.Unmarshal(val, &c)
			})
			chunks = append(chunks, c)
		}
		return nil
	})
	return chunks, err
}

func (s *KnowledgeStore) chunkText(text string, chunkSize, overlap int) []string {
	// Simplified tokenization: split by whitespace
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var chunks []string
	for i := 0; i < len(words); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		
		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)
		
		if end == len(words) {
			break
		}
	}
	return chunks
}
