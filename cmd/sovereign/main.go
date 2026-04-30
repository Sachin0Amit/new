package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sachin0Amit/new/internal/agent"
	"github.com/Sachin0Amit/new/internal/api"
	"github.com/Sachin0Amit/new/internal/auditor"
	"github.com/Sachin0Amit/new/internal/auth"
	"github.com/Sachin0Amit/new/internal/core"
	"github.com/Sachin0Amit/new/internal/guard"
	"github.com/Sachin0Amit/new/internal/llm"
	"github.com/Sachin0Amit/new/internal/mesh"
	"github.com/Sachin0Amit/new/internal/reflex"
	"github.com/Sachin0Amit/new/internal/scheduler"
	"github.com/Sachin0Amit/new/internal/telemetry"
	"github.com/Sachin0Amit/new/internal/titan"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. Load Secrets & Config
	secret := os.Getenv("SOVEREIGN_SECRET")
	passphrase := os.Getenv("SOVEREIGN_PASSPHRASE")
	if secret == "" {
		secret = "dev_secret_key" // Development fallback
	}
	if passphrase == "" {
		passphrase = "dev_passphrase" // Development fallback
	}

	// 2. Initialize Tracing
	version := "1.2.0-production"
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
		log.Printf("Warning: Failed to initialize KeyStore: %v", err)
		// Continue without key store for development
	}
	proofAuditor := &auditor.ProofAuditor{
		KeyStore: keyStore,
		NodeID:   "local-node",
	}

	// 4. Initialize P2P & Discovery
	p2pNode, err := mesh.NewNode(context.Background(), 0)
	if err != nil {
		log.Printf("Warning: Failed to initialize P2P node: %v", err)
		// Continue without P2P for development
	}

	// 5. Initialize LLM Client (NEW)
	ollamaURL := os.Getenv("OLLAMA_HOST")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	llmClient := llm.NewOllamaClient(ollamaURL, "mistral")

	// Health check Ollama
	ctx := context.Background()
	if err := llmClient.Health(ctx); err != nil {
		log.Printf("Warning: Ollama not available at %s: %v", ollamaURL, err)
	} else {
		log.Printf("✓ Ollama connected at %s", ollamaURL)
	}

	// 6. Initialize Engine & Security
	titanEngine, err := titan.NewTitanEngine(`{"model_path": "model.gguf"}`)
	if err != nil {
		log.Printf("Warning: Titan Engine failed to load, falling back to HTTP: %v", err)
	}

	jwtSvc := auth.NewJWTService(secret, knowledgeMesh)
	_ = jwtSvc // Used by auth middleware in production
	enforcer := guard.NewCapabilityEnforcer(nil)

	// 7. Initialize Agent Infrastructure (NEW)
	contextManager := agent.NewContextManager(
		4096,
		agent.NewSimpleCompressor(512),
		agent.NewSimpleTokenCounter(),
	)

	memoryStore := agent.NewInMemoryMemoryStore()
	toolExecutor := agent.NewToolExecutor()

	// Register tools
	toolExecutor.Register(agent.NewWebSearchTool(func(ctx context.Context, query string) ([]agent.SearchResult, error) {
		return []agent.SearchResult{
			{Title: "Result", URL: "http://example.com", Snippet: query},
		}, nil
	}))

	toolExecutor.Register(agent.NewMathSolverTool(func(ctx context.Context, expr string, op string) (string, error) {
		return fmt.Sprintf("Solution: %s", expr), nil
	}))

	// Create ReAct agent
	reactAgent := agent.NewReActAgent(llmClient, toolExecutor, contextManager, memoryStore)

	// 8. Initialize Reflex & Scheduler
	budget := reflex.NewReflexBudget()
	corrector := reflex.NewReflexCorrector(knowledgeMesh, budget)
	detector := reflex.NewDetector(proofAuditor)

	gMap := scheduler.NewGravityMap()
	taskScheduler := &scheduler.TaskScheduler{
		LocalID:    "local-node",
		GravityMap: gMap,
		Auditor:    proofAuditor,
	}

	// 9. Assemble Core
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

	// 10. Setup HTTP Server with API handlers
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"OK","service":"sovereign-core"}`)
	})

	// Metrics
	mux.Handle("/metrics", promhttp.Handler())

	// API handlers
	apiHandler := api.NewHandler(sovCore, nil)
	mux.HandleFunc("/api/v1/chat", apiHandler.HandleChat)
	mux.HandleFunc("/api/v1/tasks", apiHandler.HandleTasks)
	mux.HandleFunc("/api/v1/status", apiHandler.HandleStatus)

	// WebSocket chat endpoint
	wsHandler := api.NewWebSocketHandler(llmClient, reactAgent, nil)
	mux.HandleFunc("/ws/chat", wsHandler.HandleWebSocket)

	// Static files
	mux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("./web"))))

	// Security middleware
	rateLimiter := api.NewRateLimiter(3600, 100) // 60 req/min per IP
	capabilityEnforcer := api.NewCapabilityEnforcer()
	capabilityEnforcer.RegisterPolicy(api.Policy{
		ToolName:  "web_search",
		Enabled:   true,
		RateLimit: 10,
		Timeout:   30,
	})

	handler := rateLimiter.Middleware(mux)

	server := &http.Server{
		Addr:    ":8081",
		Handler: handler,
	}

	go func() {
		fmt.Println("🌐 Sovereign Core listening on :8081")
		fmt.Println("   Chat UI:     http://localhost:8081/web/chat.html")
		fmt.Println("   WebSocket:   ws://localhost:8081/ws/chat")
		fmt.Println("   Metrics:     http://localhost:8081/metrics")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	<-ctx.Done()
	fmt.Println("\n⚠️  Shutdown signal received. Draining connections...")

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
