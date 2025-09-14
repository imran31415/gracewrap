package gracewrap

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// helper to create Graceful with metrics enabled and isolated registry
func newTestGraceful(t *testing.T) *Graceful {
	t.Helper()
	cfg := DefaultConfig()
	cfg.EnableMetrics = true
	cfg.PrometheusRegistry = prometheus.NewRegistry()
	g := New(&cfg)
	return g
}

func TestReadinessAndLivenessHandlers(t *testing.T) {
	g := newTestGraceful(t)

	// Ready true
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	g.HealthHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("ready")) {
		t.Fatalf("expected body to contain 'ready' got %q", rr.Body.String())
	}

	// Flip to not ready
	g.setReady(false)
	rr = httptest.NewRecorder()
	g.HealthHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}

	// Liveness always OK
	rr = httptest.NewRecorder()
	g.LivenessHandler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHTTPMiddlewareInflight(t *testing.T) {
	g := newTestGraceful(t)

	var inHandler sync.WaitGroup
	inHandler.Add(1)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inHandler.Done()
		// simulate some work
		time.Sleep(50 * time.Millisecond)
		io.WriteString(w, "ok")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	g.httpMiddleware(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

// fakeHTTPServer implements just Shutdown used by shutdown.go
type fakeHTTPServer struct {
	shutdownCalled int
	mu             sync.Mutex
}

func (f *fakeHTTPServer) Shutdown(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.shutdownCalled++
	return nil
}

// fakeListener to satisfy Close
type fakeListener struct{}

func (f *fakeListener) Close() error              { return nil }
func (f *fakeListener) Addr() net.Addr            { return &net.IPAddr{} }
func (f *fakeListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }

func TestShutdownSequence(t *testing.T) {
	g := newTestGraceful(t)
	g.config.DrainTimeout = 100 * time.Millisecond
	g.config.HardStopTimeout = 0

	// install fake HTTP server
	httpSrv := &http.Server{}
	g.httpServers = append(g.httpServers, httpSrv)
	g.listeners = append(g.listeners, &fakeListener{})

	// install fake gRPC server using real type but not serving
	// we simply ensure Stop/GracefulStop paths do not panic by creating a server
	grpcSrv := g.NewGRPCServer()
	g.grpcServers = append(g.grpcServers, grpcSrv)

	// bump inflight then decrement in background to test wait
	g.inflight.mu.Lock()
	g.inflight.n = 1
	g.inflight.mu.Unlock()
	go func() {
		time.Sleep(20 * time.Millisecond)
		g.decInflight()
	}()

	// trigger shutdown
	g.shutdown()

	if g.Ready() {
		t.Fatalf("expected not ready after shutdown")
	}
}

func TestMetricsHandlerDisabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EnableMetrics = false
	g := New(&cfg)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	g.MetricsHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when metrics disabled, got %d", rr.Code)
	}
}

func TestMetricsHandlerEnabled(t *testing.T) {
	g := newTestGraceful(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	g.MetricsHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for metrics, got %d", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Fatalf("expected some metrics output")
	}
}
