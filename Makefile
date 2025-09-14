# Makefile for gracewrap

.PHONY: help test test-race test-short coverage build clean lint fmt vet examples ci-test

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Test targets
test: ## Run all tests
	go test -v ./...

test-race: ## Run tests with race detector
	go test -race -v ./...

test-short: ## Run tests in short mode (skip slow tests)
	go test -short -v ./...

# Coverage
coverage: ## Generate test coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

coverage-func: ## Show coverage by function
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Build targets
build: ## Build the library
	go build ./...

examples: ## Build all examples
	cd examples/http_server && go build .
	cd examples/grpc_server && go build .
	cd examples/mixed_service && go build .

# Code quality
lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

# Cleanup
clean: ## Clean build artifacts
	go clean ./...
	rm -f coverage.out coverage.html

# CI target that runs everything
ci-test: fmt vet test test-race coverage ## Run all CI checks

# Development helpers
mod-tidy: ## Tidy go modules
	go mod tidy

mod-download: ## Download go modules
	go mod download

# Check if examples work (basic smoke test)
test-examples: ## Test that examples compile and don't panic immediately
	@echo "Testing HTTP server example..."
	@cd examples/http_server && go test -v .
	@echo "Testing gRPC server example..."
	@cd examples/grpc_server && go test -v .
	@echo "Testing mixed service example..."
	@cd examples/mixed_service && go test -v .
	@echo "All examples tests passed!"

# Benchmark tests
bench: ## Run benchmark tests
	go test -bench=. -benchmem ./...

# Proof testing
proof: ## Run proof test that demonstrates GraceWrap prevents request failures
	@echo "🎯 Running Kubernetes in-flight request protection proof..."
	cd proof_tests && go test -v -run TestKubernetesInFlightRequestProof
	@echo "✅ Proof test completed - check proof_tests/results/ for output"

# Prometheus demo
demo-metrics: ## Run interactive Prometheus metrics demonstration
	@echo "🎯 Starting Prometheus metrics demonstration..."
	@echo "📖 See demo/README.md for full instructions"
	cd demo && ./metrics_demo.sh

demo-server-graceful: ## Start demo server with GraceWrap and metrics
	@echo "🛡️ Starting graceful demo server with Prometheus metrics..."
	cd demo/server && go run prometheus_demo.go graceful

demo-server-normal: ## Start demo server without GraceWrap (no metrics)
	@echo "⚡ Starting normal demo server (no graceful shutdown)..."
	cd demo/server && go run prometheus_demo.go normal

demo-load-light: ## Generate light load for metrics demo
	cd demo/loadgen && go run load_generator.go light

demo-load-heavy: ## Generate heavy load for metrics demo
	cd demo/loadgen && go run load_generator.go heavy

demo-monitoring: ## Start Prometheus + Grafana monitoring stack
	@echo "📊 Starting Prometheus + Grafana monitoring stack..."
	cd demo && docker-compose up -d
	@echo "✅ Monitoring stack started:"
	@echo "   Grafana: http://localhost:3000 (admin/admin)"
	@echo "   Prometheus: http://localhost:9090"

# Install development tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Check for vulnerabilities
vuln-check: ## Check for known vulnerabilities
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
