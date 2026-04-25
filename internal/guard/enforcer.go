package guard

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/Sachin0Amit/new/internal/auth"
)

var (
	ErrCapabilityDenied = errors.New("security: system action denied - missing capability")
	ErrPathTraversal    = errors.New("security: path traversal attempt detected")
)

type CapabilityEnforcer struct {
	db       *badger.DB
	overrides sync.Map // Map[userID]uint32
}

func NewCapabilityEnforcer(db *badger.DB) *CapabilityEnforcer {
	return &CapabilityEnforcer{db: db}
}

func (e *CapabilityEnforcer) Enforce(ctx context.Context, cap Capability) error {
	claims, ok := ctx.Value(auth.ClaimsKey).(*auth.Claims)
	if !ok {
		e.LogCheck(ctx, "Unknown", cap, false)
		return ErrCapabilityDenied
	}

	// 1. Get base capabilities from role
	var userCaps uint32
	switch claims.Role {
	case "observer": userCaps = uint32(RoleObserver)
	case "agent":    userCaps = uint32(RoleAgent)
	case "admin":    userCaps = uint32(RoleAdmin)
	case "sovereign": userCaps = uint32(RoleAdmin)
	default:         userCaps = 0
	}

	// 2. Overlay persistent overrides from Badger/Map
	if val, ok := e.overrides.Load(claims.UserID); ok {
		userCaps |= val.(uint32)
	}

	allowed := (userCaps & uint32(cap)) != 0
	e.LogCheck(ctx, claims.UserID, cap, allowed)

	if !allowed {
		return fmt.Errorf("%w: %s", ErrCapabilityDenied, cap.String())
	}
	return nil
}

func (e *CapabilityEnforcer) Grant(userID string, cap Capability) {
	val, _ := e.overrides.LoadOrStore(userID, uint32(0))
	e.overrides.Store(userID, val.(uint32)|uint32(cap))
	// In production, also persist to BadgerDB
}

func (e *CapabilityEnforcer) Revoke(userID string, cap Capability) {
	if val, ok := e.overrides.Load(userID); ok {
		e.overrides.Store(userID, val.(uint32) & ^uint32(cap))
	}
}
