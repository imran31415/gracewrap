# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **Minimum Go version is now 1.26** (required by current grpc, prometheus and golang.org/x releases). CI tests Go 1.26 and 1.27.
- Upgraded `google.golang.org/grpc` 1.59.0 → 1.85.0, `github.com/prometheus/client_golang` 1.17.0 → 1.25.0, `golang.org/x/net` 0.23.0 → 0.63.0, `google.golang.org/protobuf` 1.31.0 → 1.36.12 and all transitive dependencies across every module.

### Security
- Fixes 8 known vulnerabilities reachable from the library: GO-2026-6617, GO-2026-6612, GO-2026-6611, GO-2026-6603 (golang.org/x/net HTTP/2) and GO-2026-6443, GO-2026-6348, GO-2026-6061, GO-2026-4762 (grpc).

### Added
- Initial release of GraceWrap library
- Graceful shutdown for HTTP and gRPC servers
- Kubernetes health check endpoints (readiness/liveness)
- Prometheus metrics integration
- Signal handling (SIGTERM/SIGINT)
- In-flight request tracking and draining
- Environment-based configuration
- Comprehensive test suite (91%+ coverage)
- Race condition testing
- CI/CD pipeline with GitHub Actions
- Tag-driven release workflow that publishes GitHub Releases and the Go module proxy (see RELEASING.md)
- Examples for HTTP, gRPC, and mixed services

### Features
- `graceful.New()` - Create graceful wrapper
- `WrapHTTP()` - Wrap existing HTTP servers
- `WrapGRPC()` - Wrap existing gRPC servers
- `NewGRPCServer()` - Create gRPC server with interceptors
- `ServeGRPC()` - Start gRPC server with graceful shutdown
- `HealthHandler()` - Readiness probe endpoint
- `LivenessHandler()` - Liveness probe endpoint
- `MetricsHandler()` - Prometheus metrics endpoint
- `Wait()` - Block until shutdown signal
- `Shutdown()` - Manual shutdown trigger

### Documentation
- Comprehensive README with badges
- API documentation
- Usage examples
- Kubernetes integration guide
- Contributing guidelines
- Testing documentation

### Dependencies
- Go 1.26+
- Prometheus client library
- gRPC library
- No external runtime dependencies for core functionality
