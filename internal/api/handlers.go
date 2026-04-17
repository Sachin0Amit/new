package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/papi-ai/sovereign-core/internal/models"
	"github.com/papi-ai/sovereign-core/pkg/logger"
)

// Orchestrator defines the core cognitive interface required by the API.
type Orchestrator interface {
	SubmitTask(ctx context.Context, task *models.Task) (uuid.UUID, error)
	GetTask(ctx context.Context, id uuid.UUID) (*models.Task, error)
	GetTasks(ctx context.Context, status models.TaskStatus, limit int) ([]*models.Task, error)
}

// Handler manages the RESTful interaction with the Sovereign Core.
type Handler struct {
	orchestrator Orchestrator
	logger       logger.Logger
}

func NewHandler(core Orchestrator, l logger.Logger) *Handler {
	return &Handler{
		orchestrator: core,
		logger:       l,
	}
}

// HandleTasks manages the Submission and Retrieval of cognitive units.
func (h *Handler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var task models.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		id, err := h.orchestrator.SubmitTask(r.Context(), &task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"id": id.String()})

	case http.MethodGet:
		status := models.TaskStatus(r.URL.Query().Get("status"))
		tasks, _ := h.orchestrator.GetTasks(r.Context(), status, 100)
		json.NewEncoder(w).Encode(tasks)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleStatus provides a lean health probe for the CLI 'status' command.
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ACTIVE",
		"version": "v1.0.0-sovereign",
		"fleet":   "distributed-mesh",
	})
}
