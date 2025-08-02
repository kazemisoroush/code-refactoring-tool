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

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf docs/
	@echo "Clean completed."

.PHONY: test lint mock swagger build-api run-api serve serve-detached stop logs docker-build clean ci

ci: mock test lint clean swagger
