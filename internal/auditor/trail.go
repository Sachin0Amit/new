package auditor

import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditEntry struct {
	EntryID      string    `json:"entry_id"`
	DerivationID string    `json:"derivation_id"`
	NodeID       string    `json:"node_id"`
	Timestamp    time.Time `json:"timestamp"`
	StepType     string    `json:"step_type"`
	StepPayload  []byte    `json:"step_payload"`
	PreviousHash []byte    `json:"previous_hash"`
	Signature    []byte    `json:"signature"`
}

type ProofAuditor struct {
	KeyStore *KeyStore
	NodeID   string
}

func (a *ProofAuditor) CreateEntry(derivationID, stepType string, payload []byte, prevHash []byte) (AuditEntry, error) {
	entry := AuditEntry{
		EntryID:      uuid.New().String(),
		DerivationID: derivationID,
		NodeID:       a.NodeID,
		Timestamp:    time.Now().UTC(),
		StepType:     stepType,
		StepPayload:  payload,
		PreviousHash: prevHash,
	}

	sigPayload, err := a.SerializeForSigning(entry)
	if err != nil {
		return AuditEntry{}, err
	}

	entry.Signature = a.KeyStore.Sign(sigPayload)
	return entry, nil
}

func (a *ProofAuditor) SerializeForSigning(e AuditEntry) ([]byte, error) {
	// Exclude signature from payload
	temp := struct {
		EntryID      string    `json:"entry_id"`
		DerivationID string    `json:"derivation_id"`
		NodeID       string    `json:"node_id"`
		Timestamp    int64     `json:"timestamp"`
		StepType     string    `json:"step_type"`
		StepPayload  []byte    `json:"step_payload"`
		PreviousHash []byte    `json:"previous_hash"`
	}{
		EntryID:      e.EntryID,
		DerivationID: e.DerivationID,
		NodeID:       e.NodeID,
		Timestamp:    e.Timestamp.UnixNano(),
		StepType:     e.StepType,
		StepPayload:  e.StepPayload,
		PreviousHash: e.PreviousHash,
	}
	return json.Marshal(temp)
}

func ComputeHash(entry AuditEntry) []byte {
	data, _ := json.Marshal(entry)
	h := sha256.Sum256(data)
	return h[:]
}
