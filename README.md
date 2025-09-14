# GraceWrap

[![Go Version](https://img.shields.io/github/go-mod/go-version/imran31415/gracewrap)](https://golang.org/)
[![Build Status](https://github.com/imran31415/gracewrap/workflows/CI/badge.svg)](https://github.com/imran31415/gracewrap/actions)
[![codecov](https://codecov.io/gh/imran31415/gracewrap/branch/main/graph/badge.svg)](https://codecov.io/gh/imran31415/gracewrap)
[![Go Report Card](https://goreportcard.com/badge/github.com/imran31415/gracewrap)](https://goreportcard.com/report/github.com/imran31415/gracewrap)
[![GoDoc](https://godoc.org/github.com/imran31415/gracewrap?status.svg)](https://godoc.org/github.com/imran31415/gracewrap)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/release/imran31415/gracewrap.svg)](https://github.com/imran31415/gracewrap/releases)

A Go library for adding graceful shutdown capabilities to your existing HTTP and gRPC services. Perfect for Kubernetes deployments where you need to handle pod termination and rolling updates gracefully.

## 🚨 **The Problem**

**Kubernetes routinely terminates pods** during deployments, scaling, and node maintenance. By default, Go applications don't handle this well:

1. **SIGTERM ignored**: Most Go services don't listen for termination signals
2. **Abrupt shutdown**: After 30 seconds, Kubernetes sends SIGKILL, immediately terminating the process
3. **Request failures**: In-flight requests get killed mid-processing, causing:
   - Database transactions to rollback
   - API responses to never reach clients  
   - File operations to be left incomplete
   - User-visible 502/503 errors during deployments

**Result**: Every Kubernetes deployment causes request failures and potential data loss.

*Based on [Graceful shutdown in Go with Kubernetes](https://medium.com/insiderengineering/graceful-shutdown-in-go-with-kubernetes-7d9cfdd518d4) by Rıdvan Berkay Çetin*

## ✅ **The Solution**

GraceWrap implements proper Kubernetes pod lifecycle management:
- Listens for SIGTERM signals
- Coordinates with readiness probes  
- Waits for in-flight requests to complete
- Prevents request failures during pod termination

## ✨ Features

- **Graceful Shutdown**: Handles SIGTERM/SIGINT signals properly
- **Kubernetes Ready**: Works with pod termination and rolling updates
- **Health Checks**: Built-in readiness and liveness probe endpoints
- **Request Tracking**: Tracks in-flight requests and waits for completion
- **Easy Integration**: Wrap your existing services with minimal code changes
- **HTTP & gRPC Support**: Works with both HTTP and gRPC servers

## 📦 Installation

```bash
go get github.com/imran31415/gracewrap
```

## 🚀 Quick Start

### Basic HTTP Server

```go
package main

import (
    "context"
    "net/http"
    "github.com/imran31415/gracewrap"
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
    "github.com/imran31415/gracewrap"
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

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DRAIN_TIMEOUT_SECONDS` | How long to wait for in-flight requests | 25 |
| `HARD_STOP_TIMEOUT_SECONDS` | Final cleanup timeout | 5 |
| `LOAD_BALANCER_DELAY_SECONDS` | Delay for load balancer coordination | 1 |
| `ENABLE_METRICS` | Enable Prometheus metrics | false |

### Programmatic Configuration

```go
config := &gracewrap.Config{
    DrainTimeout:       30 * time.Second,
    HardStopTimeout:    5 * time.Second,
    LoadBalancerDelay:  2 * time.Second,  // Custom delay for your environment
    EnableMetrics:      true,
    PrometheusRegistry: prometheus.DefaultRegisterer,
}

graceful := gracewrap.New(config)
```

### Load Balancer Delay Configuration

The `LoadBalancerDelay` prevents race conditions during shutdown:

- **Default (1s)**: Works for most environments
- **Increase (2-5s)**: For slow load balancers or service mesh
- **Decrease (0-500ms)**: For fast environments or testing
- **Zero (0s)**: Disables the delay entirely

```bash
# Environment variable
export LOAD_BALANCER_DELAY_SECONDS=2
```

## ☸️ Kubernetes Integration

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

When metrics are enabled, the following metrics are available at `/metrics`:

| Metric | Type | Description |
|--------|------|-------------|
| `gracewrap_inflight_requests` | Gauge | Current number of in-flight requests |
| `gracewrap_http_requests_total` | Counter | Total HTTP requests processed |
| `gracewrap_grpc_requests_total` | Counter | Total gRPC requests processed |
| `gracewrap_shutdown_duration_seconds` | Histogram | Time taken for graceful shutdown |
| `gracewrap_readiness_status` | Gauge | Readiness status (1=ready, 0=not ready) |
| `gracewrap_shutdowns_total` | Counter | Total number of shutdowns initiated |

## 📚 API Reference

### Graceful

The main struct that wraps your services.

#### Methods

| Method | Description |
|--------|-------------|
| `New(config *Config) *Graceful` | Create a new graceful wrapper |
| `WrapHTTP(server *http.Server) error` | Wrap an existing HTTP server |
| `WrapHTTPWithListener(server *http.Server, listener net.Listener) error` | Wrap HTTP server with existing listener |
| `WrapGRPC(server *grpc.Server, listener net.Listener) error` | Wrap an existing gRPC server |
| `NewGRPCServer(opts ...grpc.ServerOption) *grpc.Server` | Create gRPC server with interceptors |
| `ServeGRPC(addr string, opts ...grpc.ServerOption) (*grpc.Server, net.Listener, error)` | Create and start gRPC server |
| `Wait(ctx context.Context) error` | Wait for shutdown signal |
| `Shutdown()` | Manually trigger shutdown |
| `Ready() bool` | Get current readiness status |
| `HealthHandler() http.Handler` | HTTP handler for readiness checks |
| `LivenessHandler() http.Handler` | HTTP handler for liveness checks |
| `MetricsHandler() http.Handler` | HTTP handler for Prometheus metrics |

## 🔧 Development

```bash
# Run all tests
make test

# Run proof test
make proof

# Generate coverage report
make coverage
```

## 📖 Examples

See the `examples/` directory for complete working examples:

- **[HTTP Server](examples/http_server/)**: Basic HTTP server with graceful shutdown
- **[gRPC Server](examples/grpc_server/)**: Basic gRPC server with graceful shutdown  
- **[Mixed Service](examples/mixed_service/)**: Both HTTP and gRPC servers together

## 🧪 Proof of Value

Prove GraceWrap prevents request failures with our definitive test:

```bash
# Run the proof test that shows GraceWrap prevents request failures
make proof
```

**Results**: Without GraceWrap: 5-94% of requests killed (depending on processing time) | With GraceWrap: 0% killed

See **[proof_tests/PROOF_OF_VALUE.md](proof_tests/PROOF_OF_VALUE.md)** for detailed results and analysis.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📝 How It Works

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

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Made with ❤️ for the Kubernetes community