package auditor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type cognitiveAuditor struct {
	entries []Entry
	mu      sync.Mutex
}

func NewAuditor() Auditor {
	return &cognitiveAuditor{
		entries: make([]Entry, 0),
	}
}

func (a *cognitiveAuditor) Log(ctx context.Context, entry Entry) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	a.entries = append(a.entries, entry)
	fmt.Printf("[AUDIT] %s | %s | %s -> %s\n", entry.Timestamp.Format(time.RFC3339), entry.Actor, entry.Action, entry.Resource)
	return nil
}

func (a *cognitiveAuditor) GetHistory(ctx context.Context, actor string, limit int) ([]Entry, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	var results []Entry
	for i := len(a.entries) - 1; i >= 0; i-- {
		if len(results) >= limit {
			break
		}
		if actor == "" || a.entries[i].Actor == actor {
			results = append(results, a.entries[i])
		}
	}
	return results, nil
}

func (a *cognitiveAuditor) Flush() error {
	// In a real system, this would write to a persistent log file or database
	return nil
}
