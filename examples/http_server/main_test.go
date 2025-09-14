package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/imran31415/gracewrap"
)

func TestHTTPServerExample(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping example test in short mode")
	}

	// Test that the server starts and can handle requests
	graceful := gracewrap.New(&gracewrap.Config{
		DrainTimeout:    100 * time.Millisecond,
		HardStopTimeout: 50 * time.Millisecond,
		EnableMetrics:   true,
	})

	// Create the same server as in main()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello from your service!\n"))
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		// Don't actually sleep in test
		_, _ = w.Write([]byte("Slow request completed!\n"))
	})

	server := &http.Server{
		Addr:    "127.0.0.1:0", // Use random port
		Handler: mux,
	}

	// Add health check endpoints like in main
	mux.Handle("/health/ready", graceful.HealthHandler())
	mux.Handle("/health/live", graceful.LivenessHandler())
	mux.Handle("/metrics", graceful.MetricsHandler())

	// Wrap the server
	if err := graceful.WrapHTTP(server); err != nil {
		t.Fatalf("failed to wrap HTTP server: %v", err)
	}

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	// Test graceful shutdown
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
