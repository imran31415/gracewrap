package gracewrap

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

// Graceful wraps your existing services with graceful shutdown capabilities.
// It handles Kubernetes pod termination, rolling updates, and provides
// health check endpoints.
type Graceful struct {
	config Config
	logger *log.Logger

	// State management
	readyMu sync.RWMutex
	ready   bool
	started time.Time

	// In-flight request tracking
	inflight struct {
		mu sync.Mutex
		n  int64
		cv *sync.Cond
	}

	// Tracked servers
	httpServers []*http.Server
	grpcServers []*grpc.Server
	listeners   []net.Listener

	// Shutdown control
	stopOnce sync.Once
	metrics  *metrics
}

// New creates a new Graceful wrapper with the given configuration.
// If config is nil, sensible defaults are used.
func New(config *Config) *Graceful {
	if config == nil {
		defaultConfig := DefaultConfig()
		config = &defaultConfig
	}

	g := &Graceful{
		config:  *config,
		ready:   true,
		started: time.Now(),
	}

	// Setup logger
	if g.config.Logger != nil {
		g.logger = g.config.Logger
	} else {
		g.logger = log.New(os.Stdout, "[gracewrap] ", log.LstdFlags|log.Lmicroseconds)
	}

	// Setup metrics if enabled
	if g.config.EnableMetrics {
		g.metrics = newMetrics(g.config.PrometheusRegistry)
	}

	// Initialize condition variable
	g.inflight.cv = sync.NewCond(&g.inflight.mu)

	return g
}

// WrapHTTP wraps an existing HTTP server with graceful shutdown capabilities.
// The server will be started in a goroutine and tracked for graceful shutdown.
func (g *Graceful) WrapHTTP(server *http.Server) error {
	// Wrap the handler with request tracking
	if server.Handler != nil {
		server.Handler = g.httpMiddleware(server.Handler)
	}

	// Start the server
	go func() {
		g.logger.Printf("HTTP server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			g.logger.Printf("HTTP server error: %v", err)
		}
	}()

	g.httpServers = append(g.httpServers, server)
	return nil
}

// WrapHTTPWithListener wraps an HTTP server that's already bound to a listener.
func (g *Graceful) WrapHTTPWithListener(server *http.Server, listener net.Listener) error {
	// Wrap the handler with request tracking
	if server.Handler != nil {
		server.Handler = g.httpMiddleware(server.Handler)
	}

	// Start the server
	go func() {
		g.logger.Printf("HTTP server starting on %s", listener.Addr())
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			g.logger.Printf("HTTP server error: %v", err)
		}
	}()

	g.httpServers = append(g.httpServers, server)
	g.listeners = append(g.listeners, listener)
	return nil
}

// WrapGRPC wraps an existing gRPC server with graceful shutdown capabilities.
func (g *Graceful) WrapGRPC(server *grpc.Server, listener net.Listener) error {
	// Note: This is a limitation - we can't add interceptors to an existing server
	// Users should create their gRPC server with our interceptors from the start
	g.logger.Printf("Warning: gRPC server already created. Consider using NewGRPCServer() for full integration.")

	// Start the server
	go func() {
		g.logger.Printf("gRPC server starting on %s", listener.Addr())
		if err := server.Serve(listener); err != nil {
			g.logger.Printf("gRPC server error: %v", err)
		}
	}()

	g.grpcServers = append(g.grpcServers, server)
	g.listeners = append(g.listeners, listener)
	return nil
}

// NewGRPCServer creates a new gRPC server with our interceptors pre-installed.
// Use this instead of grpc.NewServer() for full graceful shutdown integration.
func (g *Graceful) NewGRPCServer(opts ...grpc.ServerOption) *grpc.Server {
	opts = append(opts,
		grpc.ChainUnaryInterceptor(g.grpcUnaryInterceptor),
		grpc.ChainStreamInterceptor(g.grpcStreamInterceptor),
	)
	return grpc.NewServer(opts...)
}

// ServeGRPC creates a gRPC server with our interceptors and starts it.
func (g *Graceful) ServeGRPC(addr string, opts ...grpc.ServerOption) (*grpc.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	server := g.NewGRPCServer(opts...)

	go func() {
		g.logger.Printf("gRPC server starting on %s", addr)
		if err := server.Serve(listener); err != nil {
			g.logger.Printf("gRPC server error: %v", err)
		}
	}()

	g.grpcServers = append(g.grpcServers, server)
	g.listeners = append(g.listeners, listener)
	return server, listener, nil
}

// Wait blocks until a shutdown signal is received, then performs graceful shutdown.
// This is the main method you call after setting up your services.
func (g *Graceful) Wait(ctx context.Context) error {
	// Setup signal handling
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-ctx.Done():
		g.logger.Printf("Context canceled; initiating graceful shutdown")
		g.shutdown()
	case sig := <-sigCh:
		g.logger.Printf("Received signal %v; initiating graceful shutdown", sig)
		g.shutdown()
	}

	return nil
}

// Shutdown manually triggers graceful shutdown.
// This is useful for testing or when you want to shutdown programmatically.
func (g *Graceful) Shutdown() {
	g.shutdown()
}

// Ready returns the current readiness status.
func (g *Graceful) Ready() bool {
	g.readyMu.RLock()
	defer g.readyMu.RUnlock()
	return g.ready
}

// HealthHandler returns an HTTP handler for health checks.
// Use this for Kubernetes liveness and readiness probes.
func (g *Graceful) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.Ready() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ready\n"))
		} else {
			http.Error(w, "draining", http.StatusServiceUnavailable)
		}
	})
}

// LivenessHandler returns an HTTP handler for liveness checks.
// This always returns 200 as long as the process is running.
func (g *Graceful) LivenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("alive\n"))
	})
}

// MetricsHandler returns an HTTP handler for Prometheus metrics.
// Only available if metrics are enabled.
func (g *Graceful) MetricsHandler() http.Handler {
	if !g.config.EnableMetrics || g.metrics == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "metrics not enabled", http.StatusNotFound)
		})
	}
	return promhttp.HandlerFor(g.metrics.gatherer, promhttp.HandlerOpts{})
}
