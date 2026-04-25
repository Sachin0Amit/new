package scheduler

import (
	"sync"
)

type NodeGravity struct {
	CPUUsage    float64
	MemUsagePct float64
	ActiveTasks int
	MaxTasks    int
}

func (n NodeGravity) Score() float64 {
	// gravity = (cpu_load * 0.4) + (memory_used_pct * 0.3) + (active_derivations / max_concurrent * 0.3)
	maxTasks := float64(n.MaxTasks)
	if maxTasks == 0 {
		maxTasks = 1
	}
	score := (n.CPUUsage * 0.4) + (n.MemUsagePct * 0.3) + (float64(n.ActiveTasks)/maxTasks * 0.3)
	if score > 1.0 {
		return 1.0
	}
	return score
}

type GravityMap struct {
	mu     sync.RWMutex
	scores map[string]float64
}

func NewGravityMap() *GravityMap {
	return &GravityMap{
		scores: make(map[string]float64),
	}
}

func (m *GravityMap) Update(nodeID string, score float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scores[nodeID] = score
}

func (m *GravityMap) GetLowest(n int) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type nodeScore struct {
		id    string
		score float64
	}
	var list []nodeScore
	for id, s := range m.scores {
		list = append(list, nodeScore{id, s})
	}

	// Simple selection sort for n lowest
	for i := 0; i < len(list) && i < n; i++ {
		min := i
		for j := i + 1; j < len(list); j++ {
			if list[j].score < list[min].score {
				min = j
			}
		}
		list[i], list[min] = list[min], list[i]
	}

	var ids []string
	for i := 0; i < n && i < len(list); i++ {
		ids = append(ids, list[i].id)
	}
	return ids
}
