package fleet

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSchedulerGravity(t *testing.T) {
	scheduler := NewScheduler(0.8)

	// Node Alpha: High Load
	scheduler.UpdateNode("alpha", 0.9, 5)
	
	// Node Beta: Low Load
	scheduler.UpdateNode("beta", 0.1, 0)
	
	// Node Gamma: Medium Load
	scheduler.UpdateNode("gamma", 0.3, 2)

	best, ok := scheduler.SelectMigrationTarget()
	assert.True(t, ok)
	assert.Equal(t, "beta", best, "Beta should be selected as the lowest gravity node")
}

func TestShouldOffload(t *testing.T) {
	scheduler := NewScheduler(0.7)
	
	assert.False(t, scheduler.ShouldOffload(0.5))
	assert.True(t, scheduler.ShouldOffload(0.8))
}
