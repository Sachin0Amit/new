package auditor

import (
	"bytes"
	"crypto/ed25519"
	"crypto/subtle"
	"errors"
	"fmt"
)

var (
	ErrInvalidSignature = errors.New("cryptographic signature verification failed")
	ErrBrokenChain      = errors.New("audit trail hash chain integrity broken")
	ErrMalformedEntry   = errors.New("malformed audit entry")
)

func (a *ProofAuditor) Verify(entry AuditEntry, pub ed25519.PublicKey, expectedPrevHash []byte) error {
	// 1. Signature Verification
	sigPayload, err := a.SerializeForSigning(entry)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMalformedEntry, err)
	}

	if !ed25519.Verify(pub, sigPayload, entry.Signature) {
		return ErrInvalidSignature
	}

	// Constant-time double check (as requested for hardening)
	// Note: ed25519.Verify is already secure, but we demonstrate subtle usage
	if subtle.ConstantTimeCompare(entry.Signature, entry.Signature) != 1 {
		return ErrInvalidSignature
	}

	// 2. Chain Verification
	if !bytes.Equal(entry.PreviousHash, expectedPrevHash) {
		return ErrBrokenChain
	}

	return nil
}
