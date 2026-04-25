package reflex

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Sachin0Amit/new/internal/auditor"
)

type AnomalyType string

const (
	ConsistencyAnomaly AnomalyType = "consistency"
	TimeoutAnomaly     AnomalyType = "timeout"
	ConfidenceAnomaly  AnomalyType = "confidence"
	ChainBreakAnomaly  AnomalyType = "chain_break"
)

type AnomalyEvent struct {
	DerivationID string
	Type         AnomalyType
	StepID       string
	Message      string
}

type Detector struct {
	auditor *auditor.ProofAuditor
	events  chan auditor.AuditEntry
	output  chan AnomalyEvent
}

func NewDetector(aud *auditor.ProofAuditor) *Detector {
	return &Detector{
		auditor: aud,
		events:  make(chan auditor.AuditEntry, 100),
		output:  make(chan AnomalyEvent, 100),
	}
}

func (d *Detector) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case entry := <-d.events:
			d.checkStep(entry)
		}
	}
}

func (d *Detector) checkStep(entry auditor.AuditEntry) {
	// 1. Confidence Anomaly
	var payload map[string]interface{}
	if err := json.Unmarshal(entry.StepPayload, &payload); err == nil {
		if conf, ok := payload["confidence"].(float64); ok && conf < 0.6 {
			d.output <- AnomalyEvent{
				DerivationID: entry.DerivationID,
				Type:         ConfidenceAnomaly,
				StepID:       entry.EntryID,
				Message:      "confidence score below threshold",
			}
		}
	}

	// 2. Chain Break Anomaly (Simplified check, assumes we track last hash)
	// In production, we'd verify the signature and PreviousHash
	if len(entry.Signature) == 0 {
		d.output <- AnomalyEvent{
			DerivationID: entry.DerivationID,
			Type:         ChainBreakAnomaly,
			StepID:       entry.EntryID,
			Message:      "audit trail signature missing or invalid",
		}
	}

	// 3. Timeout Anomaly
	// Assume entry contains a field for computation time
	if start, ok := payload["start_time"].(float64); ok {
		if time.Since(time.Unix(0, int64(start))) > 10*time.Second {
			d.output <- AnomalyEvent{
				DerivationID: entry.DerivationID,
				Type:         TimeoutAnomaly,
				StepID:       entry.EntryID,
				Message:      "step execution time exceeded 10s",
			}
		}
	}
}
