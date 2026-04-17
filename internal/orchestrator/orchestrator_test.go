package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/papi-ai/sovereign-core/internal/models"
	"github.com/papi-ai/sovereign-core/internal/api"
	"github.com/papi-ai/sovereign-core/pkg/security"
	"github.com/papi-ai/sovereign-core/pkg/p2p"
)

// MockEngine implements models.InferenceEngine for testing.
type MockEngine struct{}

func (m *MockEngine) Infer(ctx context.Context, payload map[string]interface{}) (*models.TaskResult, error) {
	return &models.TaskResult{
		Data:    map[string]interface{}{"result": "success"},
		Metrics: models.InferenceMetrics{TokensPerSecond: 100},
		Completed: time.Now(),
	}, nil
}

func (m *MockEngine) GetStatus() (models.InferenceMetrics, error) {
	return models.InferenceMetrics{}, nil
}

// MockStorage implements models.StorageManager for testing.
type MockStorage struct {
	saved map[string]interface{}
}

func (m *MockStorage) Save(ctx context.Context, key string, data interface{}) error {
	m.saved[key] = data
	return nil
}

func (m *MockStorage) Load(ctx context.Context, key string, out interface{}) error {
	return nil
}

// MockSecurity implements security.SecurityManager for testing.
type MockSecurity struct{}

var _ security.SecurityManager = (*MockSecurity)(nil)

func (m *MockSecurity) Allow() bool                                { return true }
func (m *MockSecurity) Sanitize(input string) string               { return input }
func (m *MockSecurity) Validate(p map[string]interface{}) error { return nil }

func TestOrchestratorLifecycle(t *testing.T) {
	engine := &MockEngine{}
	storage := &MockStorage{saved: make(map[string]interface{})}
	sec := &MockSecurity{}
	hub := api.NewHub()
	gossip := p2p.NewGossipNode(nil)
	orch := New(context.Background(), engine, storage, sec, hub, gossip)

	task := &models.Task{
		Type:    "derivation",
		Payload: map[string]interface{}{"prompt": "pi to 100 digits"},
	}

	id, err := orch.SubmitTask(context.Background(), task)
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	if id == uuid.Nil {
		t.Fatal("Task ID is nil")
	}

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	retrieved, err := orch.GetTask(context.Background(), id)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if retrieved.Status != models.StatusCompleted {
		t.Errorf("Expected status COMPLETED, got %v", retrieved.Status)
	}

	if retrieved.Result.Data["result"] != "success" {
		t.Errorf("Expected data 'success', got %v", retrieved.Result.Data["result"])
	}
}
