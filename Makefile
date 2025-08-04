# Code Refactoring Tool - Makefile

# Show available commands
help:
	@echo "Code Refactoring Tool - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  test        - Run all tests"
	@echo "  lint        - Run golangci-lint"
	@echo "  mock        - Generate mocks using go generate"
	@echo "  build       - Build application binaries"
	@echo "  swagger     - Generate Swagger documentation"
	@echo ""
	@echo "Docker & Local Development:"
	@echo "  serve       - Start local development environment"
	@echo "  serve-detached - Start development environment in background"
	@echo "  stop        - Stop local development environment"
	@echo "  logs        - Follow API logs"
	@echo "  docker-build - Build production Docker image"
	@echo ""
	@echo "CI/CD:"
	@echo "  ci          - Run complete CI pipeline (mock, test, lint, build, swagger)"
	@echo ""
	@echo "Utility:"
	@echo "  clean       - Clean build artifacts (keeps documentation)"
	@echo "  help        - Show this help message"
	@echo ""

# Run tests for the main application
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests passed."

lint:
	@echo "Running linter..."
	@golangci-lint -v run
	@echo "Linter passed."

mock:
	@echo "Generating mocks..."
	@go generate ./...
	@echo "Mocks generated."

swagger:
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null 2>&1 || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@swag init -g cmd/api/main.go -o docs/
	@echo "Swagger documentation generated."

# Local development
serve:
	@echo "Starting local development environment..."
	@docker-compose up --build
	@echo "Environment started. API available at http://localhost:8080"

serve-detached:
	@echo "Starting local development environment in background..."
	@docker-compose up -d --build
	@echo "Environment started. API available at http://localhost:8080"

stop:
	@echo "Stopping local development environment..."
	@docker-compose down
	@echo "Environment stopped."

logs:
	@echo "Following API logs..."
	@docker-compose logs -f api

# Production Docker build
docker-build:
	@echo "Building production Docker image..."
	@docker build -t code-refactoring-tool:latest .
	@echo "Docker image built."

# Build the application binaries
build:
	@echo "Building application binaries..."
	@mkdir -p bin/
	@go build -o bin/api -ldflags="-s -w" ./cmd/api
	@echo "API binary built at bin/api"
	@echo "Binary size: $$(du -h bin/api | cut -f1)"
	@echo "Build completed."

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@echo "✅ Clean completed!"

# Make help the default target
.DEFAULT_GOAL := help

.PHONY: help test lint mock swagger build serve serve-detached stop logs docker-build clean ci

ci: mock test lint build swagger
	@echo "🎉 CI pipeline completed successfully!"
