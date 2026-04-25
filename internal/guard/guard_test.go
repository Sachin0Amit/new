package guard

import (
	"context"
	"os"
	"testing"

	"github.com/Sachin0Amit/new/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestSecurityEnforcement(t *testing.T) {
	enforcer := NewCapabilityEnforcer(nil)
	dataRoot := t.TempDir()
	sandbox := NewSandbox(enforcer, dataRoot)

	t.Run("Capability Escalation (Observer -> Exec)", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), auth.ClaimsKey, &auth.Claims{
			UserID: "obs-1",
			Role:   "observer",
		})
		
		err := enforcer.Enforce(ctx, CapExecProcess)
		assert.ErrorIs(t, err, ErrCapabilityDenied)
	})

	t.Run("Path Traversal Detection", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), auth.ClaimsKey, &auth.Claims{
			UserID: "agent-1",
			Role:   "agent",
		})

		// Try to read /etc/passwd via traversal
		_, err := sandbox.SandboxedReadFile(ctx, dataRoot+"/../../../etc/passwd")
		assert.ErrorIs(t, err, ErrPathTraversal)
	})

	t.Run("Default Deny (Missing Context)", func(t *testing.T) {
		err := enforcer.Enforce(context.Background(), CapReadFiles)
		assert.ErrorIs(t, err, ErrCapabilityDenied)
	})

	t.Run("Grant Override", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), auth.ClaimsKey, &auth.Claims{
			UserID: "user-x",
			Role:   "none",
		})

		// Initially denied
		assert.Error(t, enforcer.Enforce(ctx, CapReadFiles))

		// Grant and check
		enforcer.Grant("user-x", CapReadFiles)
		assert.NoError(t, enforcer.Enforce(ctx, CapReadFiles))
	})

	t.Run("Symlink Traversal", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), auth.ClaimsKey, &auth.Claims{
			UserID: "admin-1",
			Role:   "admin",
		})

		// Create a symlink pointing outside the root
		externalFile := t.TempDir() + "/secret.txt"
		os.WriteFile(externalFile, []byte("TOP SECRET"), 0644)
		
		symlinkPath := dataRoot + "/link_to_external"
		os.Symlink(externalFile, symlinkPath)

		_, err := sandbox.SandboxedReadFile(ctx, symlinkPath)
		assert.ErrorIs(t, err, ErrPathTraversal)
	})
}
