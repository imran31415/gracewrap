package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arsheenali/gracewrap"
)

func main() {
	// Create graceful wrapper with environment-based config
	graceful := gracewrap.New(nil) // Uses default config

	// Your existing HTTP service
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status": "ok", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: httpMux,
	}

	// Add health endpoints to your existing mux
	httpMux.Handle("/health/ready", graceful.HealthHandler())
	httpMux.Handle("/health/live", graceful.LivenessHandler())
	httpMux.Handle("/metrics", graceful.MetricsHandler())

	// Wrap your existing HTTP server
	if err := graceful.WrapHTTP(httpServer); err != nil {
		log.Fatal(err)
	}

	// Your existing gRPC service
	_ = graceful.NewGRPCServer()
	// Register your gRPC services here...
	// grpcServer.RegisterService(...)

	// Start gRPC server
	_, _, err := graceful.ServeGRPC(":9090")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Mixed service starting:")
	log.Println("  HTTP: http://localhost:8080/api/status")
	log.Println("  gRPC: localhost:9090")
	log.Println("  Health: http://localhost:8080/health/ready")
	log.Println("  Metrics: http://localhost:8080/metrics")
	log.Println("Send SIGTERM to test graceful shutdown")

	// Wait for shutdown signal
	ctx := context.Background()
	if err := graceful.Wait(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Mixed service shutdown complete")
}
