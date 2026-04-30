package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Sachin0Amit/new/internal/llm"
	"github.com/dgraph-io/badger/v4"
)

// BadgerMemoryStore implements MemoryStore using BadgerDB
type BadgerMemoryStore struct {
	db *badger.DB
	mu sync.RWMutex
}

// NewBadgerMemoryStore creates a new memory store backed by BadgerDB
func NewBadgerMemoryStore(db *badger.DB) *BadgerMemoryStore {
	return &BadgerMemoryStore{db: db}
}

// StoreEpisode stores episodic memory
func (bms *BadgerMemoryStore) StoreEpisode(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	bms.mu.Lock()
	defer bms.mu.Unlock()

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	entry := badger.NewEntry([]byte(key), bytes)
	if ttl > 0 {
		entry = entry.WithTTL(ttl)
	}

	return bms.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(entry)
	})
}

// RetrieveEpisode retrieves episodic memory
func (bms *BadgerMemoryStore) RetrieveEpisode(ctx context.Context, key string) (interface{}, error) {
	bms.mu.RLock()
	defer bms.mu.RUnlock()

	var result interface{}
	err := bms.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &result)
		})
	})

	return result, err
}

// StoreSemanticMemory stores vector embeddings
func (bms *BadgerMemoryStore) StoreSemanticMemory(ctx context.Context, text string, embedding []float32) error {
	bms.mu.Lock()
	defer bms.mu.Unlock()

	data := map[string]interface{}{
		"text":      text,
		"embedding": embedding,
		"timestamp": time.Now().Unix(),
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	key := fmt.Sprintf("semantic:%d", time.Now().UnixNano())
	entry := badger.NewEntry([]byte(key), bytes).WithTTL(24 * time.Hour)

	return bms.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(entry)
	})
}

// SearchSemanticMemory searches semantic memory
func (bms *BadgerMemoryStore) SearchSemanticMemory(ctx context.Context, query string, limit int) ([]interface{}, error) {
	bms.mu.RLock()
	defer bms.mu.RUnlock()

	var results []interface{}

	err := bms.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("semantic:")

		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		for it.Rewind(); it.Valid() && count < limit; it.Next() {
			item := it.Item()
			var data interface{}
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &data)
			})
			if err == nil {
				results = append(results, data)
				count++
			}
		}
		return nil
	})

	return results, err
}

// GetShortTermMemory returns the recent messages (short-term memory)
func (bms *BadgerMemoryStore) GetShortTermMemory(ctx context.Context) ([]llm.Message, error) {
	// This would be better implemented by integrating with context manager
	// For now, return empty
	return make([]llm.Message, 0), nil
}

// InMemoryMemoryStore is a simple in-memory implementation
type InMemoryMemoryStore struct {
	episodes map[string]interface{}
	semantic []map[string]interface{}
	mu       sync.RWMutex
}

// NewInMemoryMemoryStore creates a new in-memory memory store
func NewInMemoryMemoryStore() *InMemoryMemoryStore {
	return &InMemoryMemoryStore{
		episodes: make(map[string]interface{}),
		semantic: make([]map[string]interface{}, 0),
	}
}

// StoreEpisode stores episodic memory
func (ims *InMemoryMemoryStore) StoreEpisode(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	ims.episodes[key] = data

	// Simple TTL simulation
	if ttl > 0 {
		go func() {
			<-time.After(ttl)
			ims.mu.Lock()
			defer ims.mu.Unlock()
			delete(ims.episodes, key)
		}()
	}

	return nil
}

// RetrieveEpisode retrieves episodic memory
func (ims *InMemoryMemoryStore) RetrieveEpisode(ctx context.Context, key string) (interface{}, error) {
	ims.mu.RLock()
	defer ims.mu.RUnlock()

	if val, ok := ims.episodes[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("not found")
}

// StoreSemanticMemory stores vector embeddings
func (ims *InMemoryMemoryStore) StoreSemanticMemory(ctx context.Context, text string, embedding []float32) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	ims.semantic = append(ims.semantic, map[string]interface{}{
		"text":      text,
		"embedding": embedding,
		"timestamp": time.Now().Unix(),
	})

	return nil
}

// SearchSemanticMemory searches semantic memory
func (ims *InMemoryMemoryStore) SearchSemanticMemory(ctx context.Context, query string, limit int) ([]interface{}, error) {
	ims.mu.RLock()
	defer ims.mu.RUnlock()

	results := make([]interface{}, 0)
	for i, item := range ims.semantic {
		if i >= limit {
			break
		}
		results = append(results, item)
	}

	return results, nil
}

// GetShortTermMemory returns short-term memory
func (ims *InMemoryMemoryStore) GetShortTermMemory(ctx context.Context) ([]llm.Message, error) {
	return make([]llm.Message, 0), nil
}
