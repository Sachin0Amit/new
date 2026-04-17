package reflex

import (
	"fmt"
	"strings"

	"github.com/papi-ai/sovereign-core/internal/models"
)

// ReflexAction defines the corrective measure to be taken by the orchestrator.
type ReflexAction string

const (
	ActionNone     ReflexAction = "NONE"
	ActionRetry    ReflexAction = "RETRY"
	ActionReDerive ReflexAction = "RE_DERIVE"
	ActionAbort    ReflexAction = "ABORT"
)

// Validator defines the interface for specialized reflex evaluation logic.
type Validator interface {
	ID() string
	Validate(task *models.Task) *Evaluation
}

// ReflexEngine implements the intellectual self-monitoring layer.
type ReflexEngine struct {
	MaxDepth   int
	Validators []Validator
}

// Evaluation represents the result of a reflex analysis.
type Evaluation struct {
	Action      ReflexAction `json:"action"`
	Reason      string       `json:"reason"`
	Correction  string       `json:"correction"` // Feedback to inject into the next iteration
	ValidatorID string       `json:"validator_id,omitempty"`
}

// NewEngine initializes a reflex loop with a specific safety depth limit and default validators.
func NewEngine(maxDepth int) *ReflexEngine {
	return &ReflexEngine{
		MaxDepth: maxDepth,
		Validators: []Validator{
			&ToolValidator{},
			&SyntaxValidator{},
			&PropheticValidator{},
			&ConsistencyValidator{},
		},
	}
}

// Evaluate analyzes a task's state and determines if an autonomous reflex is required.
func (e *ReflexEngine) Evaluate(task *models.Task) Evaluation {
	if task.ReflexDepth >= e.MaxDepth {
		return Evaluation{Action: ActionAbort, Reason: "Max reflex depth reached"}
	}

	for _, v := range e.Validators {
		eval := v.Validate(task)
		if eval != nil && eval.Action != ActionNone {
			eval.ValidatorID = v.ID()
			return *eval
		}
	}

	return Evaluation{Action: ActionNone}
}

// --- Validator Implementations ---

// ToolValidator handles execution errors in sandboxed tools.
type ToolValidator struct{}

func (v *ToolValidator) ID() string { return "TOOL_ENFORCER" }
func (v *ToolValidator) Validate(task *models.Task) *Evaluation {
	if task.Type == "tool_use" && task.Result != nil {
		stderr, _ := task.Result.Data["stderr"].(string)
		exitCode, _ := task.Result.Data["exit_code"].(int)

		if exitCode != 0 || stderr != "" {
			return &Evaluation{
				Action:     ActionRetry,
				Reason:     "Tool execution failed",
				Correction: fmt.Sprintf("Environment signaled a failure (Exit Code: %d). Error: %s. Please refine the syntax and retry.", exitCode, stderr),
			}
		}
	}
	return nil
}

// SyntaxValidator detects broken code or response formats.
type SyntaxValidator struct{}

func (v *SyntaxValidator) ID() string { return "SYNTAX_CHECKER" }
func (v *SyntaxValidator) Validate(task *models.Task) *Evaluation {
	if task.Result == nil {
		return nil
	}

	output, _ := task.Result.Data["output"].(string)
	
	// Basic check for unclosed code blocks which often indicates truncation or mistake
	if strings.Count(output, "```")%2 != 0 {
		return &Evaluation{
			Action:     ActionReDerive,
			Reason:     "Detected unclosed markdown code block",
			Correction: "Your previous response contained an unclosed triple-backtick block. Please provide a complete and well-formatted response.",
		}
	}

	// Panic/Index out of bounds indicator in generated logs/output
	if strings.Contains(strings.ToLower(output), "panic:") || strings.Contains(strings.ToLower(output), "index out of range") {
		return &Evaluation{
			Action:     ActionRetry,
			Reason:     "Detected runtime-like error in output content",
			Correction: "The generated output appears to contain a runtime error or logic crash. Re-evaluate the algorithm.",
		}
	}

	return nil
}

// PropheticValidator detects indicators of missing knowledge or context.
type PropheticValidator struct{}

func (v *PropheticValidator) ID() string { return "CONTEXT_DISCOVERY" }
func (v *PropheticValidator) Validate(task *models.Task) *Evaluation {
	if task.Result == nil {
		return nil
	}

	output, _ := task.Result.Data["output"].(string)
	markers := []string{
		"i don't have information",
		"i don't know",
		"context provided is insufficient",
		"unable to find",
		"unknown subject",
	}

	for _, marker := range markers {
		if strings.Contains(strings.ToLower(output), marker) {
			return &Evaluation{
				Action:     ActionReDerive,
				Reason:     "Prophetic Context Discovery triggered: Engine signaling information gap",
				Correction: "The engine requires additional context. I will expand the RAG search horizon for this derivation.",
			}
		}
	}
	return nil
}

// ConsistencyValidator checks for logical contradictions in reasoning steps.
type ConsistencyValidator struct{}

func (v *ConsistencyValidator) ID() string { return "LOGIC_AUDITOR" }
func (v *ConsistencyValidator) Validate(task *models.Task) *Evaluation {
	if task.Result == nil {
		return nil
	}

	for _, step := range task.Result.AuditTrail {
		msg := strings.ToUpper(step.Action)
		if strings.Contains(msg, "CONTRADICTION") || strings.Contains(msg, "INCONSISTENCY") || strings.Contains(msg, "ERROR") {
			return &Evaluation{
				Action:     ActionReDerive,
				Reason:     "Logical inconsistency detected in audit trail: " + step.Action,
				Correction: "A contradiction was flagged during the reasoning process. Please restart the derivation with a focus on logical consistency.",
			}
		}
	}
	return nil
}
