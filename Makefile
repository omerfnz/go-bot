.PHONY: build test test-unit test-integration test-coverage lint fmt run clean deps help

# Variables
BINARY_NAME=serp-bot
MAIN_PATH=cmd/serp-bot/main.go
BIN_DIR=bin
COVERAGE_FILE=coverage.out

# Default target
help:
	@echo "Go-SERP-Bot Makefile Commands:"
	@echo "  make build             - Build binary"
	@echo "  make test              - Run all tests"
	@echo "  make test-unit         - Run unit tests only"
	@echo "  make test-integration  - Run integration tests"
	@echo "  make test-coverage     - Generate coverage report"
	@echo "  make lint              - Run linter"
	@echo "  make fmt               - Format code"
	@echo "  make run               - Run application"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make deps              - Download dependencies"

# Build binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"

# Run all tests
test:
	@echo "Running all tests..."
	@go test -v -cover ./...

# Run unit tests only (skip integration tests)
test-unit:
	@echo "Running unit tests..."
	@go test -v -short ./...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@go test -v -run Integration ./...

# Generate coverage report
test-coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE)
	@echo "Coverage report generated: $(COVERAGE_FILE)"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .
	@echo "Code formatted"

# Run application
run:
	@echo "Running $(BINARY_NAME)..."
	@go run $(MAIN_PATH) start

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -rf logs
	@rm -f data/stats.json
	@rm -f $(COVERAGE_FILE)
	@echo "Clean complete"

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

