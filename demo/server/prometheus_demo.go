package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/imran31415/gracewrap"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run prometheus_demo.go [graceful|normal]")
		fmt.Println("  graceful: Run with GraceWrap and Prometheus metrics")
		fmt.Println("  normal:   Run without GraceWrap (no metrics)")
		fmt.Println("")
		fmt.Println("After starting, open http://localhost:8080/metrics to see Prometheus metrics")
		fmt.Println("Then send SIGTERM (Ctrl+C) to see graceful shutdown behavior in metrics")
		os.Exit(1)
	}

	mode := os.Args[1]

	fmt.Printf("🚀 Starting Prometheus demo in %s mode...\n", mode)
	fmt.Println("📊 Metrics available at: http://localhost:8080/metrics")
	fmt.Println("🏥 Health checks at: http://localhost:8080/health/ready")
	fmt.Println("🎯 Test endpoint at: http://localhost:8080/api/test")
	fmt.Println("💾 Database sim at: http://localhost:8080/api/database")
	fmt.Println("")
	fmt.Println("💡 Try these commands in another terminal:")
	fmt.Println("   curl http://localhost:8080/metrics")
	fmt.Println("   curl http://localhost:8080/health/ready")
	fmt.Println("   curl http://localhost:8080/api/test")
	fmt.Println("   curl http://localhost:8080/api/database")
	fmt.Println("")
	fmt.Println("🔄 Send SIGTERM (Ctrl+C) to trigger shutdown and watch metrics")
	fmt.Println("")

	switch mode {
	case "graceful":
		runGracefulDemo()
	case "normal":
		runNormalDemo()
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}
}

func runGracefulDemo() {
	fmt.Println("✅ Running WITH GraceWrap + Prometheus metrics")

	// Create graceful wrapper with metrics enabled
	graceful := gracewrap.New(&gracewrap.Config{
		DrainTimeout:      10 * time.Second,
		HardStopTimeout:   2 * time.Second,
		LoadBalancerDelay: 1 * time.Second,
		EnableMetrics:     true, // Enable Prometheus metrics
	})

	// Create server with realistic endpoints
	mux := http.NewServeMux()

	// API endpoint that simulates work
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond) // Simulate processing
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "success", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	// Database simulation endpoint (longer processing)
	mux.HandleFunc("/api/database", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate database query
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"data": "database_result", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	// Add health and metrics endpoints
	mux.Handle("/health/ready", graceful.HealthHandler())
	mux.Handle("/health/live", graceful.LivenessHandler())
	mux.Handle("/metrics", graceful.MetricsHandler())

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "GraceWrap Demo Server - Metrics enabled at /metrics")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Wrap with graceful shutdown
	if err := graceful.WrapHTTP(server); err != nil {
		log.Fatal(err)
	}

	fmt.Println("🎯 Server started with GraceWrap protection")
	fmt.Println("📊 Prometheus metrics enabled and available")
	fmt.Println("🔄 Monitoring graceful shutdown behavior...")

	// Wait for shutdown signal
	graceful.Wait(context.Background())

	fmt.Println("✅ Graceful shutdown completed - check final metrics!")
}

func runNormalDemo() {
	fmt.Println("❌ Running WITHOUT GraceWrap (no metrics, abrupt shutdown)")

	// Create regular server without graceful shutdown
	mux := http.NewServeMux()

	// Same endpoints but no metrics
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "success", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("/api/database", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"data": "database_result", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	// Basic health endpoint (always returns ready)
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ready\n")
	})

	// No metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Metrics not available without GraceWrap", http.StatusNotFound)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Normal Demo Server - No metrics, no graceful shutdown")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("⚡ Server started WITHOUT graceful shutdown")
	fmt.Println("❌ No metrics available")
	fmt.Println("💀 Shutdown will be abrupt when you press Ctrl+C")

	// Start server
	log.Fatal(server.ListenAndServe())
}
