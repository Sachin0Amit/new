package guard

import (
	"context"
	"fmt"
	"strings"
)

type sovereignGuard struct {
	secret string
}

func NewGuard(secret string) Guard {
	return &sovereignGuard{secret: secret}
}

func (g *sovereignGuard) Authenticate(ctx context.Context, token string) (*User, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	// Mock implementation for demonstration
	if strings.HasPrefix(token, "sov_") {
		return &User{
			ID:   "admin",
			Role: "sovereign",
			Scope: []string{"all"},
		}, nil
	}
	return nil, ErrInvalidToken
}

func (g *sovereignGuard) Authorize(ctx context.Context, user *User, action string, resource string) error {
	if user.Role == "sovereign" {
		return nil
	}
	return ErrUnauthorized
}

func (g *sovereignGuard) SignToken(user *User) (string, error) {
	return fmt.Sprintf("sov_%s_%s", user.ID, g.secret), nil
}
