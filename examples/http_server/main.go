package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/imran31415/gracewrap"
)

func main() {
	// Create your existing HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from your service!\n")
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Simulate slow request
		fmt.Fprintf(w, "Slow request completed!\n")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Wrap your server with graceful shutdown
	graceful := gracewrap.New(&gracewrap.Config{
		DrainTimeout:    30 * time.Second,
		HardStopTimeout: 5 * time.Second,
		EnableMetrics:   true,
	})

	// Add health check endpoints
	server.Handler.(*http.ServeMux).Handle("/health/ready", graceful.HealthHandler())
	server.Handler.(*http.ServeMux).Handle("/health/live", graceful.LivenessHandler())
	server.Handler.(*http.ServeMux).Handle("/metrics", graceful.MetricsHandler())

	// Wrap your existing server
	if err := graceful.WrapHTTP(server); err != nil {
		log.Fatal(err)
	}

	log.Println("Server starting on :8080")
	log.Println("Try: curl http://localhost:8080/")
	log.Println("Try: curl http://localhost:8080/slow")
	log.Println("Health: curl http://localhost:8080/health/ready")
	log.Println("Metrics: curl http://localhost:8080/metrics")
	log.Println("Send SIGTERM to test graceful shutdown")

	// Wait for shutdown signal
	ctx := context.Background()
	if err := graceful.Wait(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server shutdown complete")
}
