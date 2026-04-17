package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/papi-ai/sovereign-core/internal/models"
)

func TestStorageLifecycle(t *testing.T) {
	tmpDir := "./test_data"
	defer os.RemoveAll(tmpDir)

	s, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer s.Close(context.Background())

	id := uuid.New()
	task := &models.Task{
		ID:        id,
		Type:      "derivation",
		Status:    models.StatusCompleted,
		CreatedAt: time.Now(),
	}

	err = s.SaveTask(context.Background(), task)
	if err != nil {
		t.Fatalf("Failed to save task: %v", err)
	}

	loaded, err := s.LoadTask(context.Background(), id)
	if err != nil {
		t.Fatalf("Failed to load task: %v", err)
	}

	if loaded.ID != id {
		t.Errorf("Expected task ID %v, got %v", id, loaded.ID)
	}
}
