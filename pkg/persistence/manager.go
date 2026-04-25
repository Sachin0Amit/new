package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/Sachin0Amit/new/internal/models"
	"github.com/Sachin0Amit/new/pkg/logger"
)

// StorageQueryable defines the sub-interface for persistence management.
type StorageQueryable interface {
	QueryTasks(ctx context.Context, status models.TaskStatus, limit int) ([]*models.Task, error)
	Delete(ctx context.Context, key string) error
}

// CleanupManager handles the pruning and health of the Sovereign persistence layer.
type CleanupManager struct {
	storage StorageQueryable
	logger  logger.Logger
	TTL     time.Duration
}

// NewCleanupManager creates a manager with a specific task retention policy.
func NewCleanupManager(storage StorageQueryable, ttl time.Duration) *CleanupManager {
	return &CleanupManager{
		storage: storage,
		logger:  logger.New(),
		TTL:     ttl,
	}
}

// Start initiates the periodic background maintenance loop.
func (c *CleanupManager) Start(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Daily maintenance
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.RunCleanup(ctx)
		}
	}
}

// RunCleanup prunes archived tasks that exceed the retention policy.
func (c *CleanupManager) RunCleanup(ctx context.Context) {
	c.logger.Info("Starting persistence cleanup...")
	
	// Query COMPLETED tasks for potential pruning
	tasks, err := c.storage.QueryTasks(ctx, models.StatusCompleted, 500)
	if err != nil {
		c.logger.Error("Cleanup query failed", logger.ErrorF(err))
		return
	}

	pruned := 0
	for _, t := range tasks {
		if time.Since(t.CompletedAt()) > c.TTL {
			key := fmt.Sprintf("task:%s", t.ID.String())
			if err := c.storage.Delete(ctx, key); err != nil {
				c.logger.Warn("Failed to prune task", logger.String("id", t.ID.String()))
				continue
			}
			pruned++
		}
	}

	c.logger.Info("Cleanup cycle complete", logger.Int("pruned_count", pruned))
}

// IntegrityCheck (stub) would identify orphaned indices in subsequent phases.
func (c *CleanupManager) IntegrityCheck(ctx context.Context) error {
	c.logger.Info("Running database integrity check...")
	return nil
}

// Helper to bridge models.Task completion time (mock for interface compatibility)
func (c *CleanupManager) getCompletionTime(t *models.Task) time.Time {
    if t.Result != nil {
        return t.Result.Completed
    }
    return t.UpdatedAt
}
