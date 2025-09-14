# GraceWrap

A Go library for adding graceful shutdown capabilities to your existing HTTP and gRPC services. Perfect for Kubernetes deployments where you need to handle pod termination and rolling updates gracefully.

## Features

- ✅ **Graceful Shutdown**: Handles SIGTERM/SIGINT signals properly
- ✅ **Kubernetes Ready**: Works perfectly with pod termination and rolling updates
- ✅ **Health Checks**: Built-in readiness and liveness probe endpoints
- ✅ **Request Tracking**: Tracks in-flight requests and waits for completion
- ✅ **Prometheus Metrics**: Optional metrics for monitoring
- ✅ **Easy Integration**: Wrap your existing services with minimal code changes
- ✅ **HTTP & gRPC Support**: Works with both HTTP and gRPC servers

## Quick Start

### Basic HTTP Server

```go
package main

import (
    "context"
    "net/http"
    "github.com/arsheenali/gracewrap"
)

func main() {
    // Your existing HTTP server
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World!"))
    })
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    // Wrap with graceful shutdown
    graceful := gracewrap.New(nil)
    graceful.WrapHTTP(server)

    // Add health endpoints
    mux.Handle("/health/ready", graceful.HealthHandler())
    mux.Handle("/health/live", graceful.LivenessHandler())

    // Wait for shutdown signal
    graceful.Wait(context.Background())
}
```

### Basic gRPC Server

```go
package main

import (
    "context"
    "github.com/arsheenali/gracewrap"
    "google.golang.org/grpc"
)

func main() {
    // Create graceful wrapper
    graceful := gracewrap.New(nil)

    // Create gRPC server with interceptors
    grpcServer := graceful.NewGRPCServer()
    
    // Register your services
    // grpcServer.RegisterService(...)

    // Start gRPC server
    graceful.ServeGRPC(":9090")

    // Wait for shutdown signal
    graceful.Wait(context.Background())
}
```

## Configuration

### Environment Variables

- `DRAIN_TIMEOUT_SECONDS`: How long to wait for in-flight requests (default: 25)
- `HARD_STOP_TIMEOUT_SECONDS`: Final cleanup timeout (default: 5)
- `ENABLE_METRICS`: Enable Prometheus metrics (default: false)

### Programmatic Configuration

```go
config := &gracewrap.Config{
    DrainTimeout:       30 * time.Second,
    HardStopTimeout:    5 * time.Second,
    EnableMetrics:      true,
    PrometheusRegistry: prometheus.DefaultRegisterer,
}

graceful := gracewrap.New(config)
```

## Kubernetes Integration

### Health Check Endpoints

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    ports:
    - containerPort: 8080
    livenessProbe:
      httpGet:
        path: /health/live
        port: 8080
      initialDelaySeconds: 10
      periodSeconds: 5
    readinessProbe:
      httpGet:
        path: /health/ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 2
    terminationGracePeriodSeconds: 35  # Should be > drain timeout
```

### Prometheus Metrics

If metrics are enabled, the following metrics are available:

- `gracewrap_inflight_requests`: Current number of in-flight requests
- `gracewrap_http_requests_total`: Total HTTP requests processed
- `gracewrap_grpc_requests_total`: Total gRPC requests processed
- `gracewrap_shutdown_duration_seconds`: Time taken for graceful shutdown
- `gracewrap_readiness_status`: Readiness status (1=ready, 0=not ready)
- `gracewrap_shutdowns_total`: Total number of shutdowns initiated

## API Reference

### Graceful

The main struct that wraps your services.

#### Methods

- `New(config *Config) *Graceful`: Create a new graceful wrapper
- `WrapHTTP(server *http.Server) error`: Wrap an existing HTTP server
- `WrapHTTPWithListener(server *http.Server, listener net.Listener) error`: Wrap HTTP server with existing listener
- `WrapGRPC(server *grpc.Server, listener net.Listener) error`: Wrap an existing gRPC server
- `NewGRPCServer(opts ...grpc.ServerOption) *grpc.Server`: Create gRPC server with interceptors
- `ServeGRPC(addr string, opts ...grpc.ServerOption) (*grpc.Server, net.Listener, error)`: Create and start gRPC server
- `Wait(ctx context.Context) error`: Wait for shutdown signal
- `Shutdown()`: Manually trigger shutdown
- `Ready() bool`: Get current readiness status
- `HealthHandler() http.Handler`: HTTP handler for readiness checks
- `LivenessHandler() http.Handler`: HTTP handler for liveness checks
- `MetricsHandler() http.Handler`: HTTP handler for Prometheus metrics

## Examples

See the `examples/` directory for complete working examples:

- `examples/http_server/`: Basic HTTP server with graceful shutdown
- `examples/grpc_server/`: Basic gRPC server with graceful shutdown
- `examples/mixed_service/`: Both HTTP and gRPC servers together

## How It Works

1. **Signal Handling**: Listens for SIGTERM/SIGINT signals
2. **Readiness Flip**: Marks service as not ready to stop new traffic
3. **Listener Close**: Closes all listeners to stop accepting new connections
4. **Graceful Shutdown**: Shuts down servers gracefully with timeout
5. **Request Draining**: Waits for in-flight requests to complete
6. **Final Cleanup**: Performs final cleanup with hard stop timeout

This ensures that:
- No new requests are accepted during shutdown
- Existing requests are allowed to complete
- The process exits cleanly within the configured timeouts
- Kubernetes can safely terminate the pod

## License

MIT
