package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/imran31415/gracewrap"
)

func TestMixedServiceExample(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping example test in short mode")
	}

	// Test that both HTTP and gRPC servers start and shut down gracefully
	graceful := gracewrap.New(&gracewrap.Config{
		DrainTimeout:    100 * time.Millisecond,
		HardStopTimeout: 50 * time.Millisecond,
		EnableMetrics:   true,
	})

	// HTTP service setup (like in main)
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	httpServer := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: httpMux,
	}

	// Add health endpoints
	httpMux.Handle("/health/ready", graceful.HealthHandler())
	httpMux.Handle("/health/live", graceful.LivenessHandler())
	httpMux.Handle("/metrics", graceful.MetricsHandler())

	// Wrap HTTP server
	if err := graceful.WrapHTTP(httpServer); err != nil {
		t.Fatalf("failed to wrap HTTP server: %v", err)
	}

	// gRPC service setup
	_ = graceful.NewGRPCServer()
	// In a real app, you'd register services here

	// Start gRPC server
	_, grpcListener, err := graceful.ServeGRPC("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start gRPC server: %v", err)
	}

	if grpcListener == nil {
		t.Fatal("expected gRPC listener to be returned")
	}

	// Give services time to start
	time.Sleep(50 * time.Millisecond)

	// Test graceful shutdown of both services
	done := make(chan struct{})
	go func() {
		defer close(done)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		graceful.Wait(ctx)
	}()

	// Trigger shutdown
	graceful.Shutdown()

	// Wait for shutdown to complete
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown took too long")
	}
}
