package guard

import (
	"context"
	"errors"
)

var (
	ErrUnauthorized = errors.New("unauthorized: insufficient permissions")
	ErrInvalidToken = errors.New("invalid or expired token")
)

// User represents a security principal in the Sovereign system.
type User struct {
	ID    string
	Role  string
	Scope []string
}

// Guard defines the security boundary for the Sovereign system.
type Guard interface {
	// Authenticate validates the provided credentials and returns a User.
	Authenticate(ctx context.Context, token string) (*User, error)
	
	// Authorize checks if a User is permitted to perform an action on a resource.
	Authorize(ctx context.Context, user *User, action string, resource string) error
	
	// SignToken generates a secure token for the given User.
	SignToken(user *User) (string, error)
}
