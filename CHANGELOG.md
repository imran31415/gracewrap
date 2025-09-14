# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
- Go 1.21+ support
- Prometheus client library
- gRPC library
- No external runtime dependencies for core functionality

## [0.1.0] - 2024-XX-XX

### Added
- Initial release
