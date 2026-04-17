package audit

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"time"

	"github.com/papi-ai/sovereign-core/internal/models"
	"github.com/papi-ai/sovereign-core/pkg/errors"
)

// Auditor manages the reasoning trace and proof-of-derivation for cognitive tasks.
type Auditor interface {
	Record(stage, action string, metadata map[string]interface{})
	Sign(result *models.TaskResult) ([]byte, error)
}

// SovereignAuditor is a local-first implementation of the intellectual transparency engine.
type SovereignAuditor struct {
	trail    []models.AuditEntry
	privKey  ed25519.PrivateKey
	pubKey   ed25519.PublicKey
}

// NewAuditor initializes a new trace and generates a transient node identity for signing.
func NewAuditor() (*SovereignAuditor, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to generate derivation keys", err)
	}

	return &SovereignAuditor{
		trail:   make([]models.AuditEntry, 0),
		privKey: priv,
		pubKey:  pub,
	}, nil
}

// Record appends a reasoning step to the current intellectual trace.
func (a *SovereignAuditor) Record(stage, action string, metadata map[string]interface{}) {
	a.trail = append(a.trail, models.AuditEntry{
		Timestamp: time.Now(),
		Stage:     stage,
		Action:    action,
		Metadata:  metadata,
	})
}

// Sign finalizes the derivation with a cryptographic proof-of-authenticity.
func (a *SovereignAuditor) Sign(result *models.TaskResult) ([]byte, error) {
	result.AuditTrail = a.trail
	
	// Canonical JSON serialization for stable signature
	data, err := json.Marshal(result.Data)
	if err != nil {
		return nil, err
	}

	signature := ed25519.Sign(a.privKey, data)
	result.Signature = signature
	
	return signature, nil
}

// VerifyDerivation ensures a task result hasn't been tampered with since its creation.
func VerifyDerivation(pubKey ed25519.PublicKey, result *models.TaskResult) (bool, error) {
	data, err := json.Marshal(result.Data)
	if err != nil {
		return false, err
	}
	
	return ed25519.Verify(pubKey, data, result.Signature), nil
}
