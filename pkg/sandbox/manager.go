package sandbox

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/papi-ai/sovereign-core/pkg/errors"
)

// ExecutionRecord captures the results and metrics of a sandboxed execution.
type ExecutionRecord struct {
	ID         uuid.UUID     `json:"id"`
	Command    string        `json:"command"`
	Stdout     string        `json:"stdout"`
	Stderr     string        `json:"stderr"`
	ExitCode   int           `json:"exit_code"`
	Duration   time.Duration `json:"duration"`
	TimedOut   bool          `json:"timed_out"`
}

// SandboxManager handles the restricted execution of external processes.
type SandboxManager struct {
	Timeout time.Duration
}

// NewManager creates a sandbox with a default timeout security policy.
func NewManager(timeout time.Duration) *SandboxManager {
	return &SandboxManager{
		Timeout: timeout,
	}
}

// Execute runs a command within a restricted environment.
func (s *SandboxManager) Execute(ctx context.Context, name string, args ...string) (*ExecutionRecord, error) {
	record := &ExecutionRecord{
		ID:      uuid.New(),
		Command: name,
	}

	// Create sub-context with strict sandbox timeout
	execCtx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, name, args...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	record.Duration = time.Since(start)

	record.Stdout = stdout.String()
	record.Stderr = stderr.String()

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			record.TimedOut = true
			return record, errors.New(errors.CodeInternal, "sandbox execution timed out", err)
		}

		if exitError, ok := err.(*exec.ExitError); ok {
			record.ExitCode = exitError.ExitCode()
		} else {
			return record, errors.New(errors.CodeInternal, "failed to initiate sandbox", err)
		}
	}

	return record, nil
}
