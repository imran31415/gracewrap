package main

import (
	"context"
	"testing"
	"time"

	"github.com/arsheenali/gracewrap"
)

func TestGRPCServerExample(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping example test in short mode")
	}

	// Test that the gRPC server starts and shuts down gracefully
	graceful := gracewrap.New(&gracewrap.Config{
		DrainTimeout:    100 * time.Millisecond,
		HardStopTimeout: 50 * time.Millisecond,
		EnableMetrics:   true,
	})

	// Create gRPC server like in main()
	grpcServer := graceful.NewGRPCServer()

	// Register the service (simplified for test)
	RegisterMyServiceServer(grpcServer, &myService{})

	// Start the gRPC server
	_, listener, err := graceful.ServeGRPC("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start gRPC server: %v", err)
	}

	if listener == nil {
		t.Fatal("expected listener to be returned")
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

func TestMyService(t *testing.T) {
	svc := &myService{}
	resp, err := svc.SayHello(context.Background(), &HelloRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("SayHello failed: %v", err)
	}
	if resp.Message != "Hello Test" {
		t.Fatalf("expected 'Hello Test', got %q", resp.Message)
	}
}
