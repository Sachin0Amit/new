package reflex

import (
	"context"
	"fmt"

	"github.com/Sachin0Amit/new/internal/mesh"
)

type Learner struct {
	mesh mesh.KnowledgeMesh
}

func NewLearner(m mesh.KnowledgeMesh) *Learner {
	return &Learner{mesh: m}
}

func (l *Learner) RecordSuccess(ctx context.Context, anomaly AnomalyType, action string) error {
	key := fmt.Sprintf("reflex:success:%s:%s", anomaly, action)
	
	var count int
	l.mesh.Retrieve(ctx, key, &count)
	count++
	
	return l.mesh.Store(ctx, key, count)
}

func (l *Learner) GetBestAction(ctx context.Context, anomaly AnomalyType) string {
	// In a real implementation, iterate through possible actions and find the one with the highest success count in Badger.
	// For now, return a default based on anomaly type.
	switch anomaly {
	case ConsistencyAnomaly: return "rollback_with_hint"
	case TimeoutAnomaly:     return "simplify_params"
	case ConfidenceAnomaly:  return "strategy_swap"
	default:                 return "abort"
	}
}
