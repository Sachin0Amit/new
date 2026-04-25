package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sachin0Amit/new/internal/auditor"
	"github.com/Sachin0Amit/new/internal/auth"
	"github.com/Sachin0Amit/new/internal/core"
	"github.com/Sachin0Amit/new/internal/guard"
	"github.com/Sachin0Amit/new/internal/mathsolver"
	"github.com/Sachin0Amit/new/internal/mesh"
	"github.com/Sachin0Amit/new/internal/metrics"
	"github.com/Sachin0Amit/new/internal/reflex"
	"github.com/Sachin0Amit/new/internal/scheduler"
	"github.com/Sachin0Amit/new/internal/telemetry"
	"github.com/Sachin0Amit/new/internal/titan"
)

func main() {
	// 1. Load Secrets & Config
	secret := os.Getenv("SOVEREIGN_SECRET")
	passphrase := os.Getenv("SOVEREIGN_PASSPHRASE")
	if secret == "" || passphrase == "" {
		log.Fatal("CRITICAL: SOVEREIGN_SECRET and SOVEREIGN_PASSPHRASE must be set")
	}

	// 2. Initialize Tracing
	version := "1.2.0-hardened"
	tp, err := telemetry.InitTracer(version)
	if err != nil {
		log.Printf("Warning: Failed to initialize tracer: %v", err)
	} else {
		defer tp.Shutdown(context.Background())
	}

	// 3. Initialize Base Modules (Storage & Audit)
	knowledgeMesh, err := mesh.NewKnowledgeMesh("./data/mesh")
	if err != nil {
		log.Fatalf("Failed to initialize Knowledge Mesh: %v", err)
	}
	defer knowledgeMesh.Close()

	keyStore, err := auditor.NewKeyStore("./data/sovereign.key", passphrase)
	if err != nil {
		log.Fatalf("Failed to initialize KeyStore: %v", err)
	}
	proofAuditor := &auditor.ProofAuditor{
		KeyStore: keyStore,
		NodeID:   "local-node", // In production, this would be the P2P ID
	}

	// 4. Initialize P2P & Discovery
	p2pNode, err := mesh.NewNode(context.Background(), 0)
	if err != nil {
		log.Fatalf("Failed to initialize P2P node: %v", err)
	}

	// 5. Initialize Engine & Security
	titanEngine, err := titan.NewTitanEngine(`{"model_path": "model.gguf"}`)
	if err != nil {
		log.Printf("Warning: Titan Engine failed to load, falling back to HTTP: %v", err)
		// Fallback logic handled in bridge.go
	}

	jwtSvc := auth.NewJWTService(secret, knowledgeMesh)
	enforcer := guard.NewCapabilityEnforcer(nil) // Badger instance managed via mesh

	// 6. Initialize Reflex & Scheduler
	budget := reflex.NewReflexBudget()
	corrector := reflex.NewReflexCorrector(knowledgeMesh, budget)
	detector := reflex.NewDetector(proofAuditor)

	gMap := scheduler.NewGravityMap()
	taskScheduler := &scheduler.TaskScheduler{
		LocalID:    string(p2pNode.Host.ID()),
		GravityMap: gMap,
		Auditor:    proofAuditor,
	}

	// 7. Assemble Core
	sovCore := core.New(
		titanEngine,
		enforcer,
		knowledgeMesh,
		p2pNode,
		proofAuditor,
		detector,
		corrector,
		taskScheduler,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := sovCore.Start(ctx); err != nil {
		log.Fatalf("Core startup failed: %v", err)
	}

	// 8. Setup HTTP Server
	mux := http.NewServeMux()

	// Observability
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/health", sovCore.HealthHandler)

	// Math Solver Integration
	mathClient := mathsolver.NewClient("http://localhost:8000")
	mux.HandleFunc("/api/v1/solve", func(w http.ResponseWriter, r *http.Request) {
		// Example tool delegation
		res, err := mathClient.Solve(r.Context(), "/solve/algebra", mathsolver.SolveRequest{
			Expression: "x^2 + 2x + 1",
			Variables:  []string{"x"},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(res)
	})

	// Security Middleware Chain
	authMiddleware := auth.AuthMiddleware(jwtSvc)
	csrfMiddleware := auth.CSRFMiddleware()
	rateLimitMiddleware := auth.RateLimitMiddleware()

	finalHandler := rateLimitMiddleware(csrfMiddleware(authMiddleware(mux)))

	server := &http.Server{
		Addr:    ":8081",
		Handler: finalHandler,
	}

	go func() {
		fmt.Printf("🌐 Sovereign Core listening on :8081 [Node: %s]\n", p2pNode.Host.ID())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	<-ctx.Done()
	
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("HTTP shutdown error: %v\n", err)
	}
	
	if err := sovCore.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Core shutdown error: %v\n", err)
	}

	fmt.Println("👋 Sovereign Core stopped.")
}
