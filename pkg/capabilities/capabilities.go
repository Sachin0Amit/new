package capabilities

import (
	"github.com/Sachin0Amit/new/pkg/errors"
)

// Capability defines a specific permission granted to a Sovereign task.
type Capability string

const (
	CapCodeExec    Capability = "CAP_CODE_EXEC"
	CapFileRead    Capability = "CAP_FILE_READ"
	CapFileWrite   Capability = "CAP_FILE_WRITE"
	CapNetworkIO   Capability = "CAP_NETWORK_IO"
	CapSystemAdmin Capability = "CAP_SYSTEM_ADMIN"
)

// Enforcer handles the authorization of capabilities for specific tasks.
type Enforcer struct {
	globalWhitelist map[Capability]bool
}

// NewEnforcer creates a new capability guard with a default security policy.
func NewEnforcer() *Enforcer {
	return &Enforcer{
		globalWhitelist: map[Capability]bool{
			CapCodeExec:  true,
			CapFileRead:  true,
			CapFileWrite: false, // Restricted by default
			CapNetworkIO: false, // Extreme isolation
		},
	}
}

// Authorize checks if a task payload contains and is permitted to use a capability.
func (e *Enforcer) Authorize(requested []Capability) error {
	for _, cap := range requested {
		allowed, ok := e.globalWhitelist[cap]
		if !ok || !allowed {
			return errors.New(errors.CodeValidation, "unauthorized capability: "+string(cap), nil)
		}
	}
	return nil
}

// MapToStrings converts a slice of capabilities to strings for serialization.
func MapToStrings(caps []Capability) []string {
	res := make([]string, len(caps))
	for i, c := range caps {
		res[i] = string(c)
	}
	return res
}
