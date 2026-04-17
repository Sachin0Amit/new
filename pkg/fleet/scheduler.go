package fleet

import (
	"sync"
)

// FleetScheduler orchestrates global resource pooling across the mesh.
type FleetScheduler struct {
	mu           sync.RWMutex
	resourceMap  map[string]NodeResourceStatus
	loadThreshold float64
}

// NodeResourceStatus tracks the compute gravity of a specific mesh node.
type NodeResourceStatus struct {
	Load      float64
	TaskCount int
}

// NewScheduler initializes the fleet orchestration engine.
func NewScheduler(threshold float64) *FleetScheduler {
	return &FleetScheduler{
		resourceMap:   make(map[string]NodeResourceStatus),
		loadThreshold: threshold,
	}
}

// UpdateNode record real-time resource telemetry from a peer.
func (s *FleetScheduler) UpdateNode(nodeID string, load float64, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resourceMap[nodeID] = NodeResourceStatus{Load: load, TaskCount: count}
}

// ShouldOffload determines if the local node should migrate a task to the fleet.
func (s *FleetScheduler) ShouldOffload(localLoad float64) bool {
	return localLoad > s.loadThreshold
}

// SelectMigrationTarget identifies the optimal node for task delegation.
func (s *FleetScheduler) SelectMigrationTarget() (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestNode string
	minGravity := 1e9

	for nodeID, status := range s.resourceMap {
		// Resource Gravity = Load * TaskCount (Simplistic but effective gravity model)
		gravity := status.Load * float64(status.TaskCount + 1)
		if gravity < minGravity {
			minGravity = gravity
			bestNode = nodeID
		}
	}

	if bestNode == "" {
		return "", false
	}
	return bestNode, true
}
