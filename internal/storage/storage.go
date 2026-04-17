package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/papi-ai/sovereign-core/internal/models"
)

// SovereignStorage implements the StorageManager interface using BadgerDB.
type SovereignStorage struct {
	db *badger.DB
}

// New initializes a new BadgerDB-backed storage engine at the specified path.
func New(path string) (*SovereignStorage, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	opts := badger.DefaultOptions(path).
		WithLogger(nil). // We'll handle logging via pkg/logger
		WithInMemory(false).
		WithSyncWrites(true)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &SovereignStorage{db: db}, nil
}

// Save persists a Sovereign object to the LSM-tree.
func (s *SovereignStorage) Save(ctx context.Context, key string, data interface{}) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), buf)
	})
}

// SaveIndexedTask persists a task and maintains its secondary status index.
func (s *SovereignStorage) SaveIndexedTask(ctx context.Context, task *models.Task) error {
	key := fmt.Sprintf("task:%s", task.ID.String())
	idxKey := fmt.Sprintf("idx:status:%s:%s", task.Status, task.ID.String())

	return s.db.Update(func(txn *badger.Txn) error {
		// 1. Save actual task
		buf, _ := json.Marshal(task)
		if err := txn.Set([]byte(key), buf); err != nil {
			return err
		}
		
		// 2. Set index pointer
		return txn.Set([]byte(idxKey), []byte(key))
	})
}

// QueryTasks retrieves a slice of tasks matching a specific status.
func (s *SovereignStorage) QueryTasks(ctx context.Context, status models.TaskStatus, limit int) ([]*models.Task, error) {
	var tasks []*models.Task
	prefix := []byte(fmt.Sprintf("idx:status:%s:", status))

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if len(tasks) >= limit {
				break
			}

			var t models.Task
			item := it.Item()
			err := item.Value(func(val []byte) error {
				// val is the pointer to the actual task key
				taskItem, err := txn.Get(val)
				if err != nil {
					return err
				}
				
				return taskItem.Value(func(taskBuf []byte) error {
					return json.Unmarshal(taskBuf, &t)
				})
			})
			if err != nil {
				continue
			}
			tasks = append(tasks, &t)
		}
		return nil
	})

	return tasks, err
}

// Delete removes a key and its value from the LSM-tree.
func (s *SovereignStorage) Delete(ctx context.Context, key string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}
func (s *SovereignStorage) Load(ctx context.Context, key string, out interface{}) error {
	return s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, out)
		})
	})
}

// SaveTask persists a Sovereign task to the LSM-tree (Legacy wrapper).
func (s *SovereignStorage) SaveTask(ctx context.Context, task *models.Task) error {
	key := fmt.Sprintf("task:%s", task.ID.String())
	return s.Save(ctx, key, task)
}

// LoadTask retrieves a Sovereign task from the LSM-tree (Legacy wrapper).
func (s *SovereignStorage) LoadTask(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	var task models.Task
	key := fmt.Sprintf("task:%s", id.String())
	err := s.Load(ctx, key, &task)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("task not found: %s", id.String())
		}
		return nil, err
	}
	return &task, nil
}

// Close gracefully shuts down the storage engine.
func (s *SovereignStorage) Close(ctx context.Context) error {
	return s.db.Close()
}
