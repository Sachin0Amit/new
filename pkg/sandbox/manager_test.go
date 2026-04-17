package sandbox

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSandboxExecution(t *testing.T) {
	manager := NewManager(2 * time.Second)
	
	// Test successful command
	record, err := manager.Execute(context.Background(), "echo", "hello sovereign")
	assert.NoError(t, err)
	assert.Equal(t, "hello sovereign\n", record.Stdout)
	assert.Equal(t, 0, record.ExitCode)

	// Test timeout
	record, err = manager.Execute(context.Background(), "sleep", "5")
	assert.Error(t, err)
	assert.True(t, record.TimedOut)
}

func TestUnsuccessfulCommand(t *testing.T) {
	manager := NewManager(2 * time.Second)
	
	// Test non-existent command
	_, err := manager.Execute(context.Background(), "non-existent-cmd")
	assert.Error(t, err)
}
