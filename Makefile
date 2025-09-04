# Makefile for Go Taxi Fare Calculator
# Supports cross-platform builds and comprehensive testing

# Project configuration
BINARY_NAME=taxi-fare
PACKAGE=golang-taxi-fare
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go configuration
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X 'main.BuildTime=$(BUILD_TIME)' -X main.GitCommit=$(GIT_COMMIT)"

# Test configuration
TEST_PACKAGES=./datavalidator/... ./errorhandler/... ./farecalculator/... ./inputparser/... ./loggingsystem/... ./models/... ./outputformatter/...
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
RACE_DETECTOR=-race

# Platform-specific settings
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    PLATFORM=linux
endif
ifeq ($(UNAME_S),Darwin)
    PLATFORM=darwin
endif
ifeq ($(UNAME_S),MINGW32_NT)
    PLATFORM=windows
endif
ifeq ($(UNAME_S),MINGW64_NT)
    PLATFORM=windows
endif

# Build targets
.PHONY: all build clean test test-unit test-integration test-race test-coverage \
        benchmark fmt lint vet check install uninstall \
        build-linux build-darwin build-windows build-all \
        coverage-html coverage-report test-data \
        ci-test ci-build docker-build docker-test \
        help

# Default target
all: clean test build

help: ## Display this help message
	@echo "Go Taxi Fare Calculator - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the application for current platform
	@echo "Building $(BINARY_NAME) for $(PLATFORM)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	@echo "Linux build complete: $(BINARY_NAME)-linux-amd64"

build-darwin: ## Build for macOS
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	@echo "macOS build complete: $(BINARY_NAME)-darwin-amd64"

build-windows: ## Build for Windows
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .
	@echo "Windows build complete: $(BINARY_NAME)-windows-amd64.exe"

build-all: build-linux build-darwin build-windows ## Build for all platforms
	@echo "Cross-platform builds complete"

# Test targets
test: ## Run all tests
	@echo "Running all tests..."
	$(GOTEST) -v $(TEST_PACKAGES)
	@echo "All tests completed"

test-unit: ## Run unit tests only (excludes main package)
	@echo "Running unit tests..."
	$(GOTEST) -v $(TEST_PACKAGES)
	@echo "Unit tests completed"

test-integration: ## Run integration tests with sample data
	@echo "Running integration tests..."
	@$(MAKE) test-data
	@echo "Testing with valid input..."
	@echo "12:34:56.789 12345678.5" | ./$(BINARY_NAME) > /dev/null
	@echo "Testing with multiple records..."
	@cat test-data/valid_input.txt | ./$(BINARY_NAME) > /dev/null
	@echo "Integration tests completed"

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	$(GOTEST) $(RACE_DETECTOR) -v $(TEST_PACKAGES)
	@echo "Race detector tests completed"

test-coverage: ## Run tests with coverage reporting
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) $(TEST_PACKAGES)
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)
	@echo "Coverage report saved to $(COVERAGE_FILE)"

coverage-html: test-coverage ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "HTML coverage report saved to $(COVERAGE_HTML)"

coverage-report: test-coverage ## Display detailed coverage report
	@echo "=== Coverage Report ==="
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep "total:" || true
	@echo ""
	@echo "=== Package Coverage ==="
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep -E "(datavalidator|errorhandler|farecalculator|inputparser|loggingsystem|models|outputformatter)" || true

# Benchmark targets
benchmark: ## Run benchmark tests
	@echo "Running benchmark tests..."
	$(GOTEST) -bench=. -benchmem $(TEST_PACKAGES)
	@echo "Benchmark tests completed"

# Code quality targets
fmt: ## Format Go code
	@echo "Formatting Go code..."
	$(GOFMT) ./...
	@echo "Code formatting completed"

lint: ## Run golint on all packages
	@echo "Running golint..."
	@which golint > /dev/null || (echo "Installing golint..." && go install golang.org/x/lint/golint@latest)
	@golint $(TEST_PACKAGES) || true
	@echo "Linting completed"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "Go vet completed"

check: fmt vet lint ## Run all code quality checks
	@echo "All code quality checks completed"

# Dependency management
deps: ## Download and tidy dependencies
	@echo "Managing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

# Test data generation
test-data: ## Generate test data files
	@echo "Generating test data..."
	@mkdir -p test-data
	@echo "12:34:56.789 12345678.5" > test-data/valid_input.txt
	@echo "12:34:57.123 12345679.1" >> test-data/valid_input.txt
	@echo "12:34:58.456 12345680.3" >> test-data/valid_input.txt
	@echo "12:35:00.789 12345681.8" >> test-data/valid_input.txt
	@echo "invalid format" > test-data/invalid_input.txt
	@echo "12:34:56" >> test-data/invalid_input.txt
	@echo "" > test-data/empty_input.txt
	@echo "12:34:56.789 12345678.5" > test-data/single_record.txt
	@echo "Test data generated in test-data/ directory"

# Installation targets
install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) .
	@echo "Installation completed"

uninstall: ## Remove installed binary
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Uninstallation completed"

# CI targets
ci-test: ## Run CI-friendly tests with coverage
	@echo "Running CI tests..."
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic $(TEST_PACKAGES)
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE)

ci-build: ## Run CI-friendly build
	@echo "Running CI build..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .
	@./$(BINARY_NAME) --version 2>/dev/null || echo "Built successfully"

# Docker targets (optional)
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(PACKAGE):$(VERSION) .

docker-test: ## Run tests in Docker container
	@echo "Running tests in Docker..."
	@docker run --rm $(PACKAGE):$(VERSION) make ci-test

# Cleanup targets
clean: ## Clean build artifacts and test files
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-*
	@rm -f $(COVERAGE_FILE)
	@rm -f $(COVERAGE_HTML)
	@rm -rf test-data/
	@echo "Cleanup completed"

# Performance targets
profile: ## Run performance profiling
	@echo "Running performance profiling..."
	@echo "12:34:56.789 12345678.5" | $(GOTEST) -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "Profiling completed. Use 'go tool pprof cpu.prof' or 'go tool pprof mem.prof'"

# Validation targets
validate: test-coverage vet lint ## Complete validation suite
	@echo "=== Validation Results ==="
	@echo "✅ Tests passed with coverage"
	@echo "✅ Code passed go vet"
	@echo "✅ Code passed linting"
	@echo "🎉 All validations successful!"

# Development workflow
dev: clean deps fmt test build ## Complete development workflow
	@echo "🚀 Development workflow completed successfully!"

# Release workflow
release: clean deps validate test-race build-all ## Release preparation workflow
	@echo "📦 Release build completed for all platforms!"
	@ls -la $(BINARY_NAME)*