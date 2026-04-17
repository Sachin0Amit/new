package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// 1. Test Defaults
	cfg, err := Load("")
	assert.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "cpu", cfg.Inference.Device)

	// 2. Test Environment Variable Overrides
	os.Setenv("SOVEREIGN_SERVER_PORT", "9090")
	os.Setenv("SOVEREIGN_INFERENCE_DEVICE", "cuda")
	defer os.Unsetenv("SOVEREIGN_SERVER_PORT")
	defer os.Unsetenv("SOVEREIGN_INFERENCE_DEVICE")

	cfgEnv, err := Load("")
	assert.NoError(t, err)
	assert.Equal(t, 9090, cfgEnv.Server.Port)
	assert.Equal(t, "cuda", cfgEnv.Inference.Device)
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tempFile := "test_config.yaml"
	content := []byte(`
server:
  port: 7070
storage:
  data_dir: "/tmp/sovereign"
`)
	err := os.WriteFile(tempFile, content, 0644)
	assert.NoError(t, err)
	defer os.Remove(tempFile)

	cfg, err := Load(tempFile)
	assert.NoError(t, err)
	assert.Equal(t, 7070, cfg.Server.Port)
	assert.Equal(t, "/tmp/sovereign", cfg.Storage.DataDir)
}
