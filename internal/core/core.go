package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/Sachin0Amit/new/internal/auditor"
	"github.com/Sachin0Amit/new/internal/guard"
	"github.com/Sachin0Amit/new/internal/mesh"
	"github.com/Sachin0Amit/new/internal/reflex"
	"github.com/Sachin0Amit/new/internal/titan"
)

// Core is the Sovereign Intelligence Core's dependency injection root.
type Core struct {
	Titan   titan.Engine
	Guard   guard.Guard
	Mesh    mesh.KnowledgeMesh
	Auditor auditor.Auditor
	Reflex  reflex.SelfHealer
}

// New initializes the system with all 5 core modules using constructor injection.
func New(t titan.Engine, g guard.Guard, m mesh.KnowledgeMesh, a auditor.Auditor, r reflex.SelfHealer) *Core {
	return &Core{
		Titan:   t,
		Guard:   g,
		Mesh:    m,
		Auditor: a,
		Reflex:  r,
	}
}

// Start kicks off background processes for all modules.
func (c *Core) Start(ctx context.Context) error {
	fmt.Println("🚀 Sovereign Intelligence Core starting...")
	
	// Start the autonomous self-healing loop
	c.Reflex.Start(ctx)
	
	c.Auditor.Log(ctx, auditor.Entry{
		Actor:    "system",
		Action:   "startup",
		Resource: "core",
		Status:   "active",
	})
	
	return nil
}

// Shutdown gracefully stops all modules in reverse order of initialization.
func (c *Core) Shutdown(ctx context.Context) error {
	fmt.Println("🛑 Shutting down Sovereign Intelligence Core...")
	
	var wg sync.WaitGroup
	wg.Add(2)
	
	go func() {
		defer wg.Done()
		c.Mesh.Close()
	}()
	
	go func() {
		defer wg.Done()
		c.Titan.Unload(ctx)
	}()
	
	c.Auditor.Flush()
	
	// Wait for cleanup with timeout handled by ctx
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		fmt.Println("✅ Clean shutdown complete.")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out")
	}
}
