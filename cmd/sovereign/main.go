package main

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/papi-ai/sovereign-core/internal/api"
	"github.com/papi-ai/sovereign-core/internal/orchestrator"
	"github.com/papi-ai/sovereign-core/internal/storage"
	"github.com/papi-ai/sovereign-core/internal/titan"
	"github.com/papi-ai/sovereign-core/pkg/logger"
	"github.com/papi-ai/sovereign-core/pkg/p2p"
	"github.com/papi-ai/sovereign-core/pkg/security"
	"github.com/papi-ai/sovereign-core/web"
)

func main() {
	// Initialize high-performance logger
	l := logger.New()
	l.Info("Sovereign Intelligence Core Initializing...")

	// Create a context that is canceled on OS interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize the Sovereign Storage (BadgerDB)
	store, err := storage.New("./data/sovereign")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close(ctx)

	// Initialize the Titan Inference Engine (C++ Bound)
	engine := titan.NewEngine("auto")
	defer engine.Unload(ctx)

	// Initialize the Security Guard
	guard := security.DefaultGuard()

	// Initialize the Telemetry Hub
	hub := api.NewHub()
	go api.StartBroadcaster(ctx, hub)

	// Initialize the P2P Discovery & Gossip (Distributed fleet)
	nodeID := uuid.New()
	p2pPort := 9090
	discovery := p2p.NewDiscoveryEngine(nodeID, p2pPort)
	gossip := p2p.NewGossipNode(discovery)
	
	if err := discovery.Start(ctx); err != nil {
		l.Error("P2P Discovery failed to start", logger.ErrorF(err))
	}
	go gossip.Listen(ctx, p2pPort)
	go gossip.Start(ctx)

	// Initialize the Intellectual Orchestrator with Mesh Sync
	core := orchestrator.New(ctx, engine, store, guard, hub, gossip)
	
	// Recover interrupted derivations
	if err := core.RecoverTasks(ctx); err != nil {
		l.Error("Initial state reconciliation failed", logger.ErrorF(err))
	}
	
	// Initialize the Dispatcher (Legacy wrapper for monitoring)
	dispatcher := orchestrator.NewDispatcher(l)

	// Start the WebSocket server for Admin Telemetry
	go func() {
		// Serve embedded React UI
		dist, _ := fs.Sub(web.AdminAssets, "admin/dist")
		http.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.FS(dist))))
		
		// Serve Main Website (Root)
		http.Handle("/", http.FileServer(http.Dir("./")))

		// API Endpoints
		h := api.NewHandler(core, l)
		http.HandleFunc("/api/v1/status", h.HandleStatus)
		http.HandleFunc("/api/v1/tasks", h.HandleTasks)
		http.HandleFunc("/ws/telemetry", hub.Serve)
		
		l.Info("Administrative Command Center active on :8081/admin")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			l.Error("WebSocket server failed", logger.ErrorF(err))
		}
	}()
	
	l.Info("Starting local inference services...")
	if err := dispatcher.Start(ctx); err != nil {
		log.Fatalf("Fatal system error: %v", err)
	}

	l.Info("Sovereign Intelligence is ACTIVE and SECURE.")
	
	// Wait for shutdown
	<-ctx.Done()
	l.Info("Shutting down Sovereign Core...")
	core.Shutdown()
	dispatcher.Shutdown()
}
