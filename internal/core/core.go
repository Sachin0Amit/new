package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/Sachin0Amit/new/internal/auditor"
	"github.com/Sachin0Amit/new/internal/guard"
	"github.com/Sachin0Amit/new/internal/mesh"
	"github.com/Sachin0Amit/new/internal/reflex"
	"github.com/Sachin0Amit/new/internal/scheduler"
	"github.com/Sachin0Amit/new/internal/titan"
)

// Core is the Sovereign Intelligence Core's dependency injection root.
type Core struct {
	Titan     *titan.TitanEngine
	Guard     *guard.CapabilityEnforcer
	Mesh      mesh.KnowledgeMesh
	P2P       *mesh.Node
	Auditor   *auditor.ProofAuditor
	Reflex    *reflex.Detector
	Corrector *reflex.ReflexCorrector
	Scheduler *scheduler.TaskScheduler
}

// New initializes the system with all core modules using constructor injection.
func New(
	t *titan.TitanEngine,
	g *guard.CapabilityEnforcer,
	m mesh.KnowledgeMesh,
	p *mesh.Node,
	a *auditor.ProofAuditor,
	d *reflex.Detector,
	c *reflex.ReflexCorrector,
	s *scheduler.TaskScheduler,
) *Core {
	return &Core{
		Titan:     t,
		Guard:     g,
		Mesh:      m,
		P2P:       p,
		Auditor:   a,
		Reflex:    d,
		Corrector: c,
		Scheduler: s,
	}
}

// Start kicks off background processes for all modules.
func (c *Core) Start(ctx context.Context) error {
	fmt.Println("🚀 Sovereign Intelligence Core starting...")

	// 1. Start P2P Mesh
	if err := c.P2P.Start(ctx); err != nil {
		return fmt.Errorf("failed to start P2P node: %w", err)
	}

	// 2. Start Autonomous Reflex detector
	go c.Reflex.Start(ctx)

	// 3. Log startup
	_, _ = c.Auditor.CreateEntry("system", "startup", []byte(`{"status":"active"}`), nil)
	// In production, we'd persist this entry via c.Mesh or similar

	fmt.Printf("✅ Core initialized with NodeID: %s\n", c.P2P.Host.ID())
	return nil
}

// Shutdown gracefully stops all modules in reverse order of initialization.
func (c *Core) Shutdown(ctx context.Context) error {
	fmt.Println("🛑 Shutting down Sovereign Intelligence Core...")

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		c.Mesh.Close()
	}()

	go func() {
		defer wg.Done()
		c.Titan.Close()
	}()

	go func() {
		defer wg.Done()
		// P2P shutdown
		c.P2P.Host.Close()
	}()

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
