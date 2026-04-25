package titan

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestTitanEngine(t *testing.T) {
	// In a real test environment, we might mock the C++ library
	// For now, we test the initialization logic.
	e := NewEngine("auto")
	defer e.Unload(context.Background())

	assert.NotNil(t, e)
	
	t.Run("GetStatus", func(t *testing.T) {
		status, err := e.GetStatus()
		assert.NoError(t, err)
		assert.Equal(t, "auto", status.HardwareAgent)
	})
}
