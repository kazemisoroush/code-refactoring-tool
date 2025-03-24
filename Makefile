# test
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests passed."

lint:
	@echo "Running linter..."
	@golangci-lint run
	@echo "Linter passed."

ci: test lint

mockgen:
	@echo "Generating mocks..."
	@go generate ./...
	@echo "Mocks generated."