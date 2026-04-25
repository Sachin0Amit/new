package auditor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

type AuditStore struct {
	db *badger.DB
}

func NewAuditStore(db *badger.DB) *AuditStore {
	return &AuditStore{db: db}
}

func (s *AuditStore) SaveEntry(ctx context.Context, entry AuditEntry) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("audit:%s:%d", entry.DerivationID, entry.Timestamp.UnixNano())
		val, _ := json.Marshal(entry)
		return txn.Set([]byte(key), val)
	})
}

func (s *AuditStore) GetDerivationTrail(derivationID string, auditor *ProofAuditor) ([]AuditEntry, error) {
	var trail []AuditEntry
	err := s.db.View(func(txn *badger.Txn) error {
		prefix := []byte(fmt.Sprintf("audit:%s:", derivationID))
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		var lastHash []byte
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var entry AuditEntry
			_ = item.Value(func(val []byte) error {
				return json.Unmarshal(val, &entry)
			})

			// Verify integrity of each link in the chain
			if err := auditor.Verify(entry, auditor.KeyStore.PublicKey(), lastHash); err != nil {
				return err
			}

			trail = append(trail, entry)
			lastHash = ComputeHash(entry)
		}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	return trail, nil
}
