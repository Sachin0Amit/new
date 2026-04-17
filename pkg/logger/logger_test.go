package logger

import (
	"testing"
)

func TestLoggingOperations(t *testing.T) {
	// Initialize with development config
	err := Init(Config{
		Level:      "debug",
		Production: false,
	})
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer Sync()

	// Verify logging methods don't panic
	Info("test info message", String("key", "value"))
	Error("test error message", Int("code", 500))
	Debug("test debug message (may not show if level is info)")
}

func TestUninitializedLogger(t *testing.T) {
	// Should not panic even if Init wasn't called (uses fallback)
	global = nil
	Info("message with uninitialized global")
	
	l := L()
	if l == nil {
		t.Error("Expected fallback logger, got nil")
	}
}
