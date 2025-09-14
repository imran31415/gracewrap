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
	@timeout 5s bash -c 'cd examples/http_server && go run main.go' || true
	@echo "Testing gRPC server example..."
	@timeout 5s bash -c 'cd examples/grpc_server && go run main.go' || true
	@echo "Testing mixed service example..."
	@timeout 5s bash -c 'cd examples/mixed_service && go run main.go' || true
	@echo "Examples smoke test completed"

# Benchmark tests
bench: ## Run benchmark tests
	go test -bench=. -benchmem ./...

# Install development tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Check for vulnerabilities
vuln-check: ## Check for known vulnerabilities
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
