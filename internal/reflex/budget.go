package reflex

import (
	"fmt"
	"sync"

	"github.com/Sachin0Amit/new/internal/metrics"
)

type ReflexBudget struct {
	counts sync.Map // map[string]int
}

func NewReflexBudget() *ReflexBudget {
	return &ReflexBudget{}
}

func (b *ReflexBudget) CheckAndIncrement(derivationID string) error {
	val, _ := b.counts.LoadOrStore(derivationID, 0)
	count := val.(int)

	if count >= 3 {
		metrics.ReflexCorrections.WithLabelValues("budget_exceeded").Inc()
		return fmt.Errorf("derivation abandoned: max reflex corrections (3) exceeded for %s", derivationID)
	}

	b.counts.Store(derivationID, count+1)
	return nil
}

func (b *ReflexBudget) GetCount(derivationID string) int {
	val, ok := b.counts.Load(derivationID)
	if !ok {
		return 0
	}
	return val.(int)
}
