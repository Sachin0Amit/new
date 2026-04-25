package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v4"
)

type meshStore struct {
	db *badger.DB
}

func NewKnowledgeMesh(path string) (KnowledgeMesh, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create mesh storage: %w", err)
	}

	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open mesh db: %w", err)
	}

	return &meshStore{db: db}, nil
}

func (m *meshStore) Store(ctx context.Context, key string, value interface{}) error {
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return m.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), buf)
	})
}

func (m *meshStore) Retrieve(ctx context.Context, key string, out interface{}) error {
	return m.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, out)
		})
	})
}

func (m *meshStore) Query(ctx context.Context, query string, limit int) ([]Result, error) {
	// Simple key-prefix search for now, extensible to vector search
	var results []Result
	prefix := []byte(query)
	err := m.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if len(results) >= limit {
				break
			}
			item := it.Item()
			err := item.Value(func(val []byte) error {
				results = append(results, Result{
					ID:      string(item.Key()),
					Content: string(val),
				})
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return results, err
}

func (m *meshStore) Close() error {
	return m.db.Close()
}
