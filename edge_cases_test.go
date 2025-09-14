package gracewrap

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// TestWaitForInflightTimeout tests timeout behavior
func TestWaitForInflightTimeout(t *testing.T) {
	g := New(nil)

	// Add in-flight requests
	g.incInflight()
	g.incInflight()

	// Test with past deadline (should return false immediately)
	pastDeadline := time.Now().Add(-1 * time.Second)
	ok := g.waitForInflight(pastDeadline)

	if ok {
		t.Error("Expected timeout with past deadline")
	}

	// Clean up
	g.decInflight()
	g.decInflight()
}

// TestTrackedStreamMethods tests gRPC stream wrapper methods
func TestTrackedStreamMethods(t *testing.T) {
	stream := &testServerStream{}
	tracked := &trackedStream{
		ServerStream: stream,
		graceful:     New(nil),
	}

	// Test RecvMsg
	if err := tracked.RecvMsg(nil); err != nil {
		t.Errorf("RecvMsg failed: %v", err)
	}

	// Test SendMsg
	if err := tracked.SendMsg("test"); err != nil {
		t.Errorf("SendMsg failed: %v", err)
	}
}

// TestUpdateReadinessFalse tests setting readiness to false
func TestUpdateReadinessFalse(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EnableMetrics = true
	cfg.PrometheusRegistry = prometheus.NewRegistry()
	g := New(&cfg)

	// Test setting to false (covers the false branch)
	g.setReady(false)
	if g.Ready() {
		t.Error("Expected ready to be false")
	}
}

// TestPeerAddrEdgeCases tests peerAddr with edge cases
func TestPeerAddrEdgeCases(t *testing.T) {
	// Test with no peer in context
	ctx := context.Background()
	if addr := peerAddr(ctx); addr != "" {
		t.Errorf("Expected empty addr for no peer, got %q", addr)
	}

	// Test with peer but nil address
	p := &peer.Peer{Addr: nil}
	ctx = peer.NewContext(context.Background(), p)
	if addr := peerAddr(ctx); addr != "" {
		t.Errorf("Expected empty addr for nil address, got %q", addr)
	}
}

// TestServeGRPCError tests ServeGRPC with invalid address
func TestServeGRPCError(t *testing.T) {
	g := New(nil)

	// Invalid address should return error
	_, _, err := g.ServeGRPC("invalid:address")
	if err == nil {
		t.Error("Expected error for invalid address")
	}
}

// testServerStream implements grpc.ServerStream for testing
type testServerStream struct {
	grpc.ServerStream
}

func (t *testServerStream) RecvMsg(m interface{}) error { return nil }
func (t *testServerStream) SendMsg(m interface{}) error { return nil }
