package guard

import (
	"context"
	"fmt"
	"runtime"
)

// LogCheck records the outcome of a capability check to the audit trail.
func (e *CapabilityEnforcer) LogCheck(ctx context.Context, userID string, cap Capability, allowed bool) {
	_, file, line, _ := runtime.Caller(2) // Get the caller of Enforce()
	
	decision := "DENY"
	if allowed {
		decision = "ALLOW"
	}

	msg := fmt.Sprintf("[%s] Capability: %s | User: %s | Source: %s:%d", 
		decision, cap.String(), userID, file, line)

	fmt.Printf("[SECURITY] %s\n", msg)

	// In a real implementation, this would call auditor.CreateEntry()
	// and persist to the audit trail in BadgerDB.
}
