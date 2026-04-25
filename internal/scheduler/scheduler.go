package scheduler

import (
	"context"
	"fmt"
	"log"

	"github.com/Sachin0Amit/new/internal/auditor"
)

type Task struct {
	ID        string
	Payload   []byte
	Status    string
}

type TaskScheduler struct {
	LocalID    string
	GravityMap *GravityMap
	Auditor    *auditor.ProofAuditor
}

func (s *TaskScheduler) Schedule(ctx context.Context, task Task, localGravity float64) (string, error) {
	// (1) If local node gravity < 0.7, run locally
	if localGravity < 0.7 {
		s.logDecision(task.ID, s.LocalID, "local")
		return s.LocalID, nil
	}

	// (2) Otherwise, query the GravityMap for the 3 nodes with lowest gravity
	candidates := s.GravityMap.GetLowest(3)
	for _, nodeID := range candidates {
		if nodeID == s.LocalID {
			continue
		}

		// (3) Send the task to the best remote node
		// In production, this would call a libp2p service
		log.Printf("[SCHEDULER] Offloading task %s to remote node %s", task.ID, nodeID)
		s.logDecision(task.ID, nodeID, "remote")
		return nodeID, nil
	}

	// (4) Fallback to local
	s.logDecision(task.ID, s.LocalID, "fallback_local")
	return s.LocalID, nil
}

func (s *TaskScheduler) logDecision(taskID, nodeID, strategy string) {
	msg := fmt.Sprintf("Task %s scheduled to %s via %s", taskID, nodeID, strategy)
	log.Println("[SCHEDULER]", msg)
	// Integration with auditor.Log would go here
}
