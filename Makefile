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

build-api:
	@echo "Building API server..."
	@go build -o bin/api-server ./cmd/api/
	@echo "API server built."

run-api: swagger build-api
	@echo "Running API server..."
	@./bin/api-server

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf docs/
	@echo "Clean completed."

.PHONY: test lint mock swagger build-api run-api clean ci

ci: mock test lint clean swagger
