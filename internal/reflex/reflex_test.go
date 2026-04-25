package reflex

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Sachin0Amit/new/internal/auditor"
	"github.com/Sachin0Amit/new/internal/mesh"
	"github.com/stretchr/testify/assert"
)

func TestAutonomousReflex(t *testing.T) {
	tmpDir := t.TempDir()
	m, _ := mesh.NewKnowledgeMesh(tmpDir)
	budget := NewReflexBudget()
	corrector := NewReflexCorrector(m, budget)
	detector := NewDetector(nil) // Auditor not needed for confidence check

	t.Run("Confidence Anomaly Trigger", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]interface{}{
			"confidence": 0.4,
		})
		
		entry := auditor.AuditEntry{
			DerivationID: "d-1",
			EntryID:      "s-1",
			StepPayload:  payload,
			Signature:    []byte("valid"),
		}

		// Inject event manually into detector's logic
		detector.checkStep(entry)
		
		select {
		case event := <-detector.output:
			assert.Equal(t, ConfidenceAnomaly, event.Type)
			err := corrector.HandleAnomaly(context.Background(), event)
			assert.NoError(t, err)
		default:
			t.Fatal("Expected anomaly event was not triggered")
		}
	})

	t.Run("Budget Exhaustion", func(t *testing.T) {
		derivationID := "d-budget"
		event := AnomalyEvent{DerivationID: derivationID, Type: TimeoutAnomaly}

		// First 3 should pass
		for i := 0; i < 3; i++ {
			err := corrector.HandleAnomaly(context.Background(), event)
			assert.NoError(t, err)
		}

		// 4th should fail
		err := corrector.HandleAnomaly(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max reflex corrections (3) exceeded")
	})

	t.Run("Chain Break Critical Failure", func(t *testing.T) {
		event := AnomalyEvent{DerivationID: "d-compromised", Type: ChainBreakAnomaly}
		err := corrector.HandleAnomaly(context.Background(), event)
		assert.NoError(t, err)

		var status string
		m.Retrieve(context.Background(), "derivation:status:d-compromised", &status)
		assert.Equal(t, "COMPROMISED", status)
	})
}
