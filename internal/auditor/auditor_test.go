package auditor

import (
	"crypto/ed25519"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestAdversarialAuditor(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := tmpDir + "/sovereign.key"
	ks, _ := NewKeyStore(keyPath, "super-secret-passphrase")
	auditor := &ProofAuditor{KeyStore: ks, NodeID: "test-node"}

	// Setup Badger
	opts := badger.DefaultOptions(tmpDir + "/badger")
	opts.Logger = nil
	db, _ := badger.Open(opts)
	defer db.Close()
	store := NewAuditStore(db)

	derivationID := "task-123"

	// 1. Create a valid chain of 3 entries
	e1, _ := auditor.CreateEntry(derivationID, "start", []byte("init"), nil)
	store.SaveEntry(nil, e1)
	
	h1 := ComputeHash(e1)
	e2, _ := auditor.CreateEntry(derivationID, "step", []byte("logic"), h1)
	store.SaveEntry(nil, e2)

	h2 := ComputeHash(e2)
	e3, _ := auditor.CreateEntry(derivationID, "complete", []byte("done"), h2)
	store.SaveEntry(nil, e3)

	t.Run("Happy Path", func(t *testing.T) {
		trail, err := store.GetDerivationTrail(derivationID, auditor)
		assert.NoError(t, err)
		assert.Len(t, trail, 3)
	})

	t.Run("Token Tampering", func(t *testing.T) {
		// Tamper with E2 payload directly in DB
		tamperedE2 := e2
		tamperedE2.StepPayload = []byte("MALICIOUS")
		store.SaveEntry(nil, tamperedE2)

		_, err := store.GetDerivationTrail(derivationID, auditor)
		assert.ErrorIs(t, err, ErrInvalidSignature)
	})

	t.Run("Broken Chain", func(t *testing.T) {
		// Restore E2, then modify E3's PreviousHash
		store.SaveEntry(nil, e2)
		brokenE3 := e3
		brokenE3.PreviousHash = []byte("BAD-HASH")
		// Resign so signature is valid but chain is broken
		sigPayload, _ := auditor.SerializeForSigning(brokenE3)
		brokenE3.Signature = ks.Sign(sigPayload)
		store.SaveEntry(nil, brokenE3)

		_, err := store.GetDerivationTrail(derivationID, auditor)
		assert.ErrorIs(t, err, ErrBrokenChain)
	})

	t.Run("Wrong Public Key", func(t *testing.T) {
		store.SaveEntry(nil, e3) // restore
		_, otherPriv, _ := ed25519.GenerateKey(nil)
		otherPub := otherPriv.Public().(ed25519.PublicKey)

		err := auditor.Verify(e1, otherPub, nil)
		assert.ErrorIs(t, err, ErrInvalidSignature)
	})
}
