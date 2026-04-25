package reflex

import (
	"testing"

	"github.com/Sachin0Amit/new/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestReflexEvaluationToolFailure(t *testing.T) {
	engine := NewEngine(3)

	task := &models.Task{
		Type: "tool_use",
		Result: &models.TaskResult{
			Data: map[string]interface{}{
				"exit_code": 1,
				"stderr":    "permission denied",
			},
		},
	}

	eval := engine.Evaluate(task)
	assert.Equal(t, ActionRetry, eval.Action)
	assert.Contains(t, eval.Correction, "permission denied")
}

func TestReflexEvaluationMaxDepth(t *testing.T) {
	engine := NewEngine(2)

	task := &models.Task{
		ReflexDepth: 2,
	}

	eval := engine.Evaluate(task)
	assert.Equal(t, ActionAbort, eval.Action)
}

func TestReflexEvaluationReasoningAnomaly(t *testing.T) {
	engine := NewEngine(3)

	task := &models.Task{
		Type: "derivation",
		Result: &models.TaskResult{
			AuditTrail: []models.AuditEntry{
				{Stage: "INFERENCE", Action: "Cognitive error detected"},
			},
		},
	}

	eval := engine.Evaluate(task)
	assert.Equal(t, ActionReDerive, eval.Action)
	assert.Equal(t, "LOGIC_AUDITOR", eval.ValidatorID)
}

func TestReflexEvaluationSyntaxError(t *testing.T) {
	engine := NewEngine(3)

	task := &models.Task{
		Type: "derivation",
		Result: &models.TaskResult{
			Data: map[string]interface{}{
				"output": "```python\nprint('hello')\n", // Missing closing block
			},
		},
	}

	eval := engine.Evaluate(task)
	assert.Equal(t, ActionReDerive, eval.Action)
	assert.Equal(t, "SYNTAX_CHECKER", eval.ValidatorID)
}

func TestReflexEvaluationPropheticDiscovery(t *testing.T) {
	engine := NewEngine(3)

	task := &models.Task{
		Type: "derivation",
		Result: &models.TaskResult{
			Data: map[string]interface{}{
				"output": "I don't have information on the specific fleet topology.",
			},
		},
	}

	eval := engine.Evaluate(task)
	assert.Equal(t, ActionReDerive, eval.Action)
	assert.Equal(t, "CONTEXT_DISCOVERY", eval.ValidatorID)
}

func TestReflexEvaluationConsistency(t *testing.T) {
	engine := NewEngine(3)

	task := &models.Task{
		Type: "derivation",
		Result: &models.TaskResult{
			AuditTrail: []models.AuditEntry{
				{Stage: "SYMBOLIC", Action: "Contradiction found in premise A"},
			},
		},
	}

	eval := engine.Evaluate(task)
	assert.Equal(t, ActionReDerive, eval.Action)
	assert.Equal(t, "LOGIC_AUDITOR", eval.ValidatorID)
}
