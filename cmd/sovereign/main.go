package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

const banner = `
╔══════════════════════════════════════════════╗
║      SOVEREIGN INTELLIGENCE CORE v1.0       ║
║    Local-First · Privacy-Preserving · Fast   ║
╚══════════════════════════════════════════════╝
`

func main() {
	fmt.Print(banner)

	// Initialize high-performance logger
	l := logger.New()
	l.Info("Sovereign Intelligence Core Initializing...")

	// Determine port
	port := os.Getenv("SOVEREIGN_PORT")
	if port == "" {
		port = "8081"
	}

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

	// Build HTTP mux with CORS
	mux := http.NewServeMux()

	// Serve embedded React UI
	dist, _ := fs.Sub(web.AdminAssets, "admin/dist")
	mux.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.FS(dist))))

	// Serve Main Website (Root) with proper caching
	fileServer := http.FileServer(http.Dir("./"))
	mux.Handle("/", addCacheHeaders(fileServer))

	// API Endpoints
	h := api.NewHandler(core, l)
	mux.HandleFunc("/api/v1/status", h.HandleStatus)
	mux.HandleFunc("/api/v1/tasks", h.HandleTasks)
	mux.HandleFunc("/api/v1/chat", h.HandleChat)
	mux.HandleFunc("/ws/telemetry", hub.Serve)

	// Finance Endpoints
	fh := api.NewFinanceHandler(l)
	mux.HandleFunc("/api/v1/finance/market", fh.HandleMarketData)
	mux.HandleFunc("/api/v1/finance/indicators", fh.HandleIndicators)
	mux.HandleFunc("/api/v1/finance/predict", fh.HandlePrediction)
	mux.HandleFunc("/api/v1/finance/risk", fh.HandleRisk)
	mux.HandleFunc("/api/v1/finance/intelligence", fh.HandleIntelligence)
	mux.HandleFunc("/api/v1/finance/trader/status", fh.HandleTraderStatus)
	mux.HandleFunc("/api/v1/finance/trader/toggle", fh.HandleTraderToggle)
	mux.HandleFunc("/api/v1/finance/trader/order", fh.HandleManualOrder)
	mux.HandleFunc("/api/v1/finance/trader/login", fh.HandleBrokerLogin)

	// Wrap with CORS middleware
	handler := corsMiddleware(mux)

	// Start HTTP server with graceful shutdown
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		l.Info(fmt.Sprintf("🌐 Web Interface: http://localhost:%s", port))
		l.Info(fmt.Sprintf("📡 Admin Dashboard: http://localhost:%s/admin", port))
		l.Info(fmt.Sprintf("🔌 API Endpoint: http://localhost:%s/api/v1", port))
		l.Info(fmt.Sprintf("📊 WebSocket Telemetry: ws://localhost:%s/ws/telemetry", port))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("HTTP server failed", logger.ErrorF(err))
		}
	}()

	l.Info("Starting local inference services...")
	if err := dispatcher.Start(ctx); err != nil {
		log.Fatalf("Fatal system error: %v", err)
	}

	l.Info("✅ Sovereign Intelligence is ACTIVE and SECURE.")

	// Wait for shutdown signal
	<-ctx.Done()
	l.Info("Shutting down Sovereign Core...")

	// Graceful shutdown with 10s timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		l.Error("HTTP server shutdown error", logger.ErrorF(err))
	}

	core.Shutdown()
	dispatcher.Shutdown()
	l.Info("Sovereign Core shutdown complete. Goodbye.")
}

// corsMiddleware adds permissive CORS headers for local development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// addCacheHeaders sets cache headers for static files.
func addCacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		next.ServeHTTP(w, r)
	})
}
