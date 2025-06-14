.PHONY: build run test clean fmt vet deps help

# Binary name
BINARY_NAME := consensus
BUILD_DIR := ./bin

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Run all tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...

# Run tests with detailed coverage report
test-coverage-html:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run all quality checks
check: fmt vet test
	@echo "All checks passed!"

# Development setup
setup: deps
	@echo "Setting up development environment..."
	@echo "Make sure you have the following environment variables set:"
	@echo "  - OPENAI_API_KEY"
	@echo "  - ANTHROPIC_API_KEY"
	@echo "  - GEMINI_API_KEY"

# Show help
help:
	@echo "Available targets:"
	@echo "  build            Build the binary"
	@echo "  test             Run all tests"
	@echo "  test-verbose     Run tests with verbose output"
	@echo "  test-coverage    Run tests with coverage"
	@echo "  test-coverage-html Generate HTML coverage report"
	@echo "  fmt              Format Go code"
	@echo "  vet              Run go vet"
	@echo "  deps             Download and tidy dependencies"
	@echo "  clean            Clean build artifacts"
	@echo "  check            Run fmt, vet, and test"
	@echo "  setup            Setup development environment"
	@echo "  help             Show this help message"
