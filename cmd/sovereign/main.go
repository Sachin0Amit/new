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

	"github.com/Sachin0Amit/new/internal/auditor"
	"github.com/Sachin0Amit/new/internal/core"
	"github.com/Sachin0Amit/new/internal/guard"
	"github.com/Sachin0Amit/new/internal/mesh"
	"github.com/Sachin0Amit/new/internal/reflex"
	"github.com/Sachin0Amit/new/internal/titan"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. Initialize all implementations
	t := titan.NewEngine("auto")
	
	g := guard.NewGuard(os.Getenv("SOVEREIGN_SECRET"))
	
	m, err := mesh.NewKnowledgeMesh("./data/mesh")
	if err != nil {
		log.Fatalf("Failed to initialize Knowledge Mesh: %v", err)
	}
	
	a := auditor.NewAuditor()
	
	r := reflex.NewSelfHealer(1 * time.Minute)

	// 2. Dependency Injection via internal/core
	sovCore := core.New(t, g, m, a, r)

	// 3. System Startup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := sovCore.Start(ctx); err != nil {
		log.Fatalf("Core startup failed: %v", err)
	}

	// 4. Setup HTTP Server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		status, _ := sovCore.Titan.GetStatus()
		fmt.Fprintf(w, "Sovereign Engine Status: %+v", status)
	})

	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	go func() {
		fmt.Println("🌐 Server listening on :8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// 5. Graceful Shutdown with 30-second drain timeout
	<-ctx.Done()
	fmt.Println("\n⚠️ Shutdown signal received. Draining connections...")

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
