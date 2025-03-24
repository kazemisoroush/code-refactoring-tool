test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests passed."

lint:
	@echo "Running linter..."
	@golangci-lint -v run
	@echo "Linter passed."

ci: test lint

mock:
	@echo "Generating mocks..."
	@go generate ./...
	@echo "Mocks generated."