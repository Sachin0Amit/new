package plugins

import (
	"testing"

	"github.com/Sachin0Amit/new/internal/models"
	"github.com/Sachin0Amit/new/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// MockProcessorPlugin modifies the task payload for testing.
type MockProcessorPlugin struct {
	CapturedHook models.HookPoint
}

func (m *MockProcessorPlugin) Metadata() models.PluginManifest {
	return models.PluginManifest{
		ID:   "mock-processor",
		Name: "Mock Processor",
	}
}

func (m *MockProcessorPlugin) OnHook(hook models.HookPoint, task *models.Task) error {
	m.CapturedHook = hook
	if hook == models.HookPreProcessing {
		task.Payload["plugin_modified"] = true
	}
	return nil
}

func TestPluginHookInvocation(t *testing.T) {
	registry := NewRegistry(logger.New())
	plugin := &MockProcessorPlugin{}
	registry.Register(plugin)

	task := &models.Task{
		Payload: make(map[string]interface{}),
	}

	// 1. Invoke PreProcessing Hook
	registry.InvokeHook(models.HookPreProcessing, task)

	assert.Equal(t, models.HookPreProcessing, plugin.CapturedHook)
	assert.True(t, task.Payload["plugin_modified"].(bool))

	// 2. List manifests
	manifests := registry.ListActivePlugins()
	assert.Len(t, manifests, 1)
	assert.Equal(t, "mock-processor", manifests[0].ID)
}
