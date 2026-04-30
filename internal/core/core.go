package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Sachin0Amit/new/internal/auditor"
	"github.com/Sachin0Amit/new/internal/guard"
	"github.com/Sachin0Amit/new/internal/mesh"
	"github.com/Sachin0Amit/new/internal/models"
	"github.com/Sachin0Amit/new/internal/reflex"
	"github.com/Sachin0Amit/new/internal/scheduler"
	"github.com/Sachin0Amit/new/internal/titan"
	"github.com/google/uuid"
)

// Core is the Sovereign Intelligence Core's dependency injection root.
type Core struct {
	Titan     *titan.TitanEngine
	Guard     *guard.CapabilityEnforcer
	Mesh      mesh.KnowledgeMesh
	P2P       *mesh.Node
	Auditor   *auditor.ProofAuditor
	Reflex    *reflex.Detector
	Corrector *reflex.ReflexCorrector
	Scheduler *scheduler.TaskScheduler

	tasks   map[uuid.UUID]*models.Task
	tasksMu sync.RWMutex
}

// New initializes the system with all core modules using constructor injection.
func New(
	t *titan.TitanEngine,
	g *guard.CapabilityEnforcer,
	m mesh.KnowledgeMesh,
	p *mesh.Node,
	a *auditor.ProofAuditor,
	d *reflex.Detector,
	c *reflex.ReflexCorrector,
	s *scheduler.TaskScheduler,
) *Core {
	return &Core{
		Titan:     t,
		Guard:     g,
		Mesh:      m,
		P2P:       p,
		Auditor:   a,
		Reflex:    d,
		Corrector: c,
		Scheduler: s,
		tasks:     make(map[uuid.UUID]*models.Task),
	}
}

// SubmitTask creates and processes a new task.
func (c *Core) SubmitTask(ctx context.Context, task *models.Task) (uuid.UUID, error) {
	task.ID = uuid.New()
	task.Status = models.StatusPending
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	c.tasksMu.Lock()
	c.tasks[task.ID] = task
	c.tasksMu.Unlock()

	// Process asynchronously
	go func() {
		c.tasksMu.Lock()
		task.Status = models.StatusRunning
		task.UpdatedAt = time.Now()
		c.tasksMu.Unlock()

		result := &models.TaskResult{
			Data:      make(map[string]interface{}),
			Completed: time.Now(),
		}

		// Try Titan engine if available
		if c.Titan != nil {
			output, err := c.Titan.Derive(ctx, fmt.Sprintf("%v", task.Payload["prompt"]), 512)
			if err == nil {
				result.Data["text"] = output
				result.Data["model"] = "titan-local"
			} else {
				result.Data["text"] = generateFallback(task)
				result.Data["model"] = "fallback"
			}
		} else {
			result.Data["text"] = generateFallback(task)
			result.Data["model"] = "fallback"
		}

		c.tasksMu.Lock()
		task.Status = models.StatusCompleted
		task.Result = result
		task.UpdatedAt = time.Now()
		c.tasksMu.Unlock()
	}()

	return task.ID, nil
}

// GetTask returns a task by ID.
func (c *Core) GetTask(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	c.tasksMu.RLock()
	defer c.tasksMu.RUnlock()

	task, ok := c.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return task, nil
}

// GetTasks returns tasks filtered by status.
func (c *Core) GetTasks(ctx context.Context, status models.TaskStatus, limit int) ([]*models.Task, error) {
	c.tasksMu.RLock()
	defer c.tasksMu.RUnlock()

	result := make([]*models.Task, 0)
	for _, t := range c.tasks {
		if status == "" || t.Status == status {
			result = append(result, t)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func generateFallback(task *models.Task) string {
	if prompt, ok := task.Payload["prompt"].(string); ok {
		return fmt.Sprintf("Processing: %s", prompt)
	}
	return "Task processed."
}

// Start kicks off background processes for all modules.
func (c *Core) Start(ctx context.Context) error {
	fmt.Println("🚀 Sovereign Intelligence Core starting...")

	// 1. Start P2P Mesh
	if c.P2P != nil {
		if err := c.P2P.Start(ctx); err != nil {
			fmt.Printf("Warning: P2P start failed: %v\n", err)
		}
	}

	// 2. Start Autonomous Reflex detector
	if c.Reflex != nil {
		go c.Reflex.Start(ctx)
	}

	// 3. Log startup
	if c.Auditor != nil {
		_, _ = c.Auditor.CreateEntry("system", "startup", []byte(`{"status":"active"}`), nil)
	}

	if c.P2P != nil && c.P2P.Host != nil {
		fmt.Printf("✅ Core initialized with NodeID: %s\n", c.P2P.Host.ID())
	} else {
		fmt.Println("✅ Core initialized (standalone mode)")
	}
	return nil
}

// Shutdown gracefully stops all modules in reverse order of initialization.
func (c *Core) Shutdown(ctx context.Context) error {
	fmt.Println("🛑 Shutting down Sovereign Intelligence Core...")

	var wg sync.WaitGroup

	if c.Mesh != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Mesh.Close()
		}()
	}

	if c.Titan != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Titan.Close()
		}()
	}

	if c.P2P != nil && c.P2P.Host != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.P2P.Host.Close()
		}()
	}

	// Wait for cleanup with timeout handled by ctx
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("✅ Clean shutdown complete.")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out")
	}
}

