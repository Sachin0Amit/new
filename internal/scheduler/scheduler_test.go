package scheduler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFleetScheduling(t *testing.T) {
	gMap := NewGravityMap()
	
	// Simulate 5 nodes
	nodes := []string{"node-1", "node-2", "node-3", "node-4", "node-5"}
	
	// Saturate 4 of them (gravity=0.95)
	for i := 0; i < 4; i++ {
		gMap.Update(nodes[i], 0.95)
	}
	// Node 5 is idle
	gMap.Update("node-5", 0.1)

	sched := &TaskScheduler{
		LocalID:    "node-1",
		GravityMap: gMap,
	}

	t.Run("Offload to Idle Node", func(t *testing.T) {
		task := Task{ID: "task-1"}
		// Local gravity is 0.95, should offload
		target, err := sched.Schedule(context.Background(), task, 0.95)
		
		assert.NoError(t, err)
		assert.Equal(t, "node-5", target, "Should have scheduled to the idle node-5")
	})

	t.Run("Local Execution when Idle", func(t *testing.T) {
		task := Task{ID: "task-2"}
		// Local gravity is 0.2, should stay local
		target, err := sched.Schedule(context.Background(), task, 0.2)
		
		assert.NoError(t, err)
		assert.Equal(t, "node-1", target, "Should have scheduled locally")
	})
}

func TestMigrationTrigger(t *testing.T) {
	gMap := NewGravityMap()
	gMap.Update("node-idle", 0.1)
	
	migrator := &TaskMigrator{
		LocalID:    "node-heavy",
		GravityMap: gMap,
	}

	t.Run("Migration Triggered (High Gravity, Low Progress)", func(t *testing.T) {
		target, err := migrator.MaybeMigrate(context.Background(), "derivation-1", 0.98, 0.1)
		assert.NoError(t, err)
		assert.Equal(t, "node-idle", target)
	})

	t.Run("Migration Skipped (High Progress)", func(t *testing.T) {
		target, err := migrator.MaybeMigrate(context.Background(), "derivation-2", 0.98, 0.6)
		assert.NoError(t, err)
		assert.Equal(t, "node-heavy", target, "Should not migrate if progress > 30%")
	})
}
